package main

// Basic imports
import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/buger/jsonparser"
	"github.com/pioz/faker"
	"github.com/rs/xid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/test/internal/fork"
)

type VaultSuite struct {
	suite.Suite

	host        string
	port        string
	mailpitHost string
	mailpitPort string
	redisHost   string
	redisPort   string
	binaryPath  string
	envs        []string
	token       string

	conn       *grpc.ClientConn
	grpcClient pb.GophKeeperServiceClient

	serverProcess *fork.BackgroundProcess
}

// SetupSuite bootstraps suite dependencies
func (suite *VaultSuite) SetupSuite() {
	err := suite.getEnvs()
	if err != nil {
		suite.T().Errorf("Невозможно запустить тесты, не установлены переменные окружения: %s", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = fork.WaitPort(ctx, "tcp", suite.redisHost+":"+suite.redisPort)
	if err != nil {
		suite.T().Errorf("Не удалось дождаться пока порт %s станет доступен для запроса: %s", suite.redisPort, err)
		return
	}
	err = fork.WaitPort(ctx, "tcp", suite.mailpitHost+":"+suite.mailpitPort)
	if err != nil {
		suite.T().Errorf("Не удалось дождаться пока порт %s станет доступен для запроса: %s", suite.mailpitPort, err)
		return
	}

	suite.envs = os.Environ()
	suite.envs = append(suite.envs, "GOPHKEEPER_SERVICE_TOKEN_SECRET_KEY="+faker.StringWithSize(20))
	suite.serverUp(ctx, suite.envs)

	config := &tls.Config{
		ServerName:         suite.host,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS13,
	}

	addr := suite.host + ":" + suite.port
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(credentials.NewTLS(config)), grpc.WithBlock())
	if err != nil {
		suite.Require().Error(err, "gRPC dial error")
		return
	}

	grpcClient := pb.NewGophKeeperServiceClient(conn)
	suite.conn = conn
	suite.grpcClient = grpcClient
}

func (suite *VaultSuite) serverUp(ctx context.Context, envs []string) {
	p := fork.NewBackgroundProcess(context.Background(),
		suite.binaryPath,
		fork.WithEnv(envs...),
	)

	err := p.Start(ctx)
	if err != nil {
		suite.T().Errorf("Невозможно запустить процесс командой %s: %s. Переменные окружения: %+v", p, err, envs)
		return
	}

	err = fork.WaitPort(ctx, "tcp", suite.host+":"+suite.port)
	if err != nil {
		suite.T().Errorf("Не удалось дождаться пока порт %s станет доступен для запроса: %s", suite.port, err)
		if out := p.Stderr(ctx); len(out) > 0 {
			suite.T().Logf("Получен STDERR лог агента:\n\n%s\n\n", string(out))
		}
		if out := p.Stdout(ctx); len(out) > 0 {
			suite.T().Logf("Получен STDOUT лог агента:\n\n%s\n\n", string(out))
		}
		return
	}

	suite.serverProcess = p
}

func (suite *VaultSuite) serverDown() {
	if suite.serverProcess == nil {
		return
	}

	exitCode, err := suite.serverProcess.Stop(syscall.SIGINT, syscall.SIGKILL)
	if err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return
		}
		suite.T().Logf("Не удалось остановить процесс с помощью сигнала ОС: %s", err)
		return
	}

	if exitCode > 0 {
		suite.T().Logf("Процесс завершился с не нулевым статусом %d", exitCode)
	}

	// try to read stdout/stderr
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if out := suite.serverProcess.Stderr(ctx); len(out) > 0 {
		suite.T().Logf("Получен STDERR лог агента:\n\n%s\n\n", string(out))
	}
	if out := suite.serverProcess.Stdout(ctx); len(out) > 0 {
		suite.T().Logf("Получен STDOUT лог агента:\n\n%s\n\n", string(out))
	}
}

func (suite *VaultSuite) TearDownSuite() {
	suite.serverDown()

	// delete all mails in mailpit
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/api/v1/messages", suite.mailpitHost+":"+suite.mailpitPort), http.NoBody)
	if err != nil {
		suite.T().Logf("Create request error: %s", err)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		suite.T().Logf("Delete all mails from mailpit error: %s", err)
		return
	}
	defer resp.Body.Close()
}

func (suite *VaultSuite) TestVault() {
	ctx, cancel := context.WithTimeout(context.Background(), testsTimeout)
	defer cancel()

	// generate random auth hash
	authHash := make([]byte, 32)
	_, err := rand.Read(authHash)
	suite.Require().NoError(err, "Generate random auth hash error")
	suite.T().Logf("Generate random auth hash: len %d", len(authHash))

	email := faker.SafeEmail()
	suite.T().Logf("Generate random email: %s", email)

	suite.Run("register and login", func() {
		// register
		_, err = suite.grpcClient.Register(ctx, &pb.RegisterRequest{
			Email: email,
			Hash:  authHash,
		})
		if err != nil {
			suite.Require().NoError(err, "gRPC user register error", err)
		}

		// try to login, email not confirmed
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: email,
			Hash:  authHash,
		})
		suite.Require().ErrorIs(err, pb.ErrUserEmailNotVerified)

		getIdLastMail := func(toEmail string) ([]byte, error) {
			resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/search?query=to:\"%s\"&limit=1", suite.mailpitHost+":"+suite.mailpitPort, toEmail))
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, errors.New("code not found")
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			return body, nil
		}

		getBodyLastMail := func(id string) ([]byte, error) {
			resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/message/%s", suite.mailpitHost+":"+suite.mailpitPort, id))
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, errors.New("code not found")
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			return body, nil
		}

		// waiting for the mail
		var body []byte
		for attempt := 0; attempt < 20; attempt++ {
			body, err = getIdLastMail(email)
			if err != nil {
				time.Sleep(500 * time.Millisecond)
				continue
			}

			var count int64
			count, err = jsonparser.GetInt(body, "messages_count")
			if err != nil || count == 0 {
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}
		suite.Require().NoError(err, "Not found any mails in mailpit to the email %s", email)
		id, err := jsonparser.GetString(body, "messages", "[0]", "ID")
		suite.Require().NoError(err, "Wrong response format from mailpit")

		body, err = getBodyLastMail(id)
		suite.Require().NoError(err, "Not found mail by id in mailpit")

		text, err := jsonparser.GetString(body, "Text")
		suite.Require().NoError(err, "Wrong response format from mailpit")

		// looking for the verification code in the mail
		substr := `activate your account: `
		pos := strings.Index(text, substr)
		suite.Require().NotEqual(-1, pos, "Verification code not found in mail")
		codeInMail := text[pos+len(substr) : pos+len(substr)+emailCodeLength]

		suite.T().Logf("Found verification code in mail: %s", codeInMail)

		// login with the verification code, got token
		resp, err := suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email:     email,
			Hash:      authHash,
			EmailCode: codeInMail,
		})
		suite.Require().NoError(err, "gRPC user login with mail verification code error", err)
		suite.Assert().NotEqual(0, len(resp.Token), "Token not found in response")

		suite.token = resp.Token
	})

	itemId := xid.New().String()
	var setServerUpdatedAt int64

	suite.Run("add vault item", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		now := time.Now().Unix()

		respSet, err := suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:              itemId,
				Name:            faker.String(),
				Itype:           pb.IType(vault.Text),
				Value:           []byte(faker.ArticleWithParagraphCount(faker.IntInRange(0, 10))),
				ServerUpdatedAt: now,
			},
		})
		suite.Require().NoError(err, "gRPC add vault item error", err)
		setServerUpdatedAt = respSet.ServerUpdatedAt
		suite.Assert().LessOrEqual(now, setServerUpdatedAt, "returned server update time must be equal or greater than time of request")

		respList, err := suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: now - 1, // -1 to avoid case server update time equal time of request
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Assert().Len(respList.Items, 1, "returned wrong number of vault items")

		respList, err = suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: setServerUpdatedAt,
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Assert().Len(respList.Items, 0, "returned wrong number of vault items")
	})

	suite.Run("update vault item", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		updatedName := faker.String()

		respSet, err := suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:              itemId,
				Name:            updatedName,
				Itype:           pb.IType(vault.Text),
				Value:           []byte(faker.ArticleWithParagraphCount(faker.IntInRange(0, 10))),
				ServerUpdatedAt: setServerUpdatedAt,
			},
		})
		suite.Require().NoError(err, "gRPC update vault item error", err)
		setServerUpdatedAt = respSet.ServerUpdatedAt

		respList, err := suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{
			Since: setServerUpdatedAt - 1,
		})
		suite.Require().NoError(err, "gRPC list vault items error", err)
		suite.Assert().Len(respList.Items, 1, "returned wrong number of vault items")
		suite.Assert().Equal(updatedName, respList.Items[0].Name, "returned wrong vault item name")
	})

	suite.Run("update vault item with conflict version", func() {
		ctx := newContextWithAuthData(ctx, suite.token)

		_, err := suite.grpcClient.SetVaultItem(ctx, &pb.SetVaultItemRequest{
			Item: &pb.VaultItem{
				Id:              itemId,
				Name:            faker.String(),
				Itype:           pb.IType(vault.Text),
				Value:           []byte(faker.ArticleWithParagraphCount(faker.IntInRange(0, 10))),
				ServerUpdatedAt: setServerUpdatedAt - 10, // client has old version
			},
		})
		suite.Assert().ErrorIs(err, pb.ErrVaultItemConflictVersion)
	})

}

func (suite *VaultSuite) getEnvs() error {
	var ok bool
	if suite.host, ok = os.LookupEnv("TEST_GRPC_HOST"); !ok {
		return errors.New("TEST_GRPC_HOST not set")
	}
	if suite.port, ok = os.LookupEnv("TEST_GRPC_PORT"); !ok {
		return errors.New("TEST_GRPC_PORT not set")
	}
	if suite.mailpitHost, ok = os.LookupEnv("TEST_MAIL_HOST"); !ok {
		return errors.New("TEST_MAIL_HOST not set")
	}
	if suite.mailpitPort, ok = os.LookupEnv("TEST_MAIL_PORT"); !ok {
		return errors.New("TEST_MAIL_PORT not set")
	}
	if suite.redisHost, ok = os.LookupEnv("TEST_REDIS_HOST"); !ok {
		return errors.New("TEST_REDIS_HOST not set")
	}
	if suite.redisPort, ok = os.LookupEnv("TEST_REDIS_PORT"); !ok {
		return errors.New("TEST_REDIS_PORT not set")
	}
	if suite.binaryPath, ok = os.LookupEnv("TEST_BINARY_PATH"); !ok {
		return errors.New("TEST_BINARY_PATH not set")
	}

	return nil
}
