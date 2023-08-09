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
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/test/internal/fork"
)

type AuthSuite struct {
	suite.Suite

	host        string
	port        string
	mailpitHost string
	mailpitPort string
	redisHost   string
	redisPort   string
	binaryPath  string
	envs        []string

	conn       *grpc.ClientConn
	grpcClient pb.GophKeeperServiceClient

	serverProcess *fork.BackgroundProcess
}

// SetupSuite bootstraps suite dependencies
func (suite *AuthSuite) SetupSuite() {
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

func (suite *AuthSuite) serverUp(ctx context.Context, envs []string) {
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

func (suite *AuthSuite) serverDown() {
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

func (suite *AuthSuite) TearDownSuite() {
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

func (suite *AuthSuite) TestAuth() {
	ctx, cancel := context.WithTimeout(context.Background(), testsTimeout)
	defer cancel()

	// generate random auth hash
	authHash := make([]byte, 32)
	_, err := rand.Read(authHash)
	suite.Require().NoError(err, "Generate random auth hash error")
	suite.T().Logf("Generate random auth hash: len %d", len(authHash))

	email := faker.SafeEmail()
	suite.T().Logf("Generate random email: %s", email)

	suite.Run("register", func() {
		// register
		_, err = suite.grpcClient.Register(ctx, &pb.RegisterRequest{
			Email: email,
			Hash:  authHash,
		})
		suite.Require().NoError(err, "gRPC user register error", err)

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
	})

	suite.Run("register: bad arguments", func() {
		authHash := make([]byte, 32)
		_, err := rand.Read(authHash)
		suite.Require().NoError(err, "Generate random auth hash error")

		// the same email
		_, err = suite.grpcClient.Register(ctx, &pb.RegisterRequest{
			Email: email,
			Hash:  authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrUserAlreadyExists)

		// bad email
		_, err = suite.grpcClient.Register(ctx, &pb.RegisterRequest{
			Email: faker.Username() + "@superpupkinmupkin.io",
			Hash:  authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidEmailFormat)

		// bad email
		_, err = suite.grpcClient.Register(ctx, &pb.RegisterRequest{
			Email: faker.Username() + "io.com",
			Hash:  authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidEmailFormat)

		// empty email
		_, err = suite.grpcClient.Register(ctx, &pb.RegisterRequest{
			Email: "",
			Hash:  authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidEmailFormat)

		// empty hash
		_, err = suite.grpcClient.Register(ctx, &pb.RegisterRequest{
			Email: faker.SafeEmail(),
			Hash:  []byte{},
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidHashFormat)
	})

	suite.Run("login: bad arguments", func() {
		// email/user not exist
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: faker.FreeEmail(),
			Hash:  authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrUserNotExists)

		// bad email
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: faker.Username() + "@superpupkinmupkin.io",
			Hash:  authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidEmailFormat)

		// bad email
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: faker.Username() + "io.com",
			Hash:  authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidEmailFormat)

		// empty email
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: "",
			Hash:  authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidEmailFormat)

		// empty hash
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: faker.SafeEmail(),
			Hash:  []byte{},
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidHashFormat)
	})

	suite.Run("get items: bad auth", func() {
		_, err = suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{})
		suite.Assert().ErrorIs(err, pb.ErrEmptyAuthData)
	})

	suite.Run("restart server with token lifetime equals 2 sec", func() {
		suite.serverDown()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		envs := make([]string, len(suite.envs)+1)
		copy(envs, suite.envs)
		envs = append(envs, "GOPHKEEPER_SERVICE_TOKEN_LIFETIME=2s")

		suite.serverUp(ctx, envs)
	})

	suite.Run("token lifetime work check", func() {
		// login with the verification code, got token
		resp, err := suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: email,
			Hash:  authHash,
		})
		suite.Require().NoError(err, "gRPC existed user login error", err)
		suite.Assert().NotEqual(0, len(resp.Token), "Token not found in response")

		ctx := newContextWithAuthData(ctx, resp.Token)
		_, err = suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{})
		suite.Assert().NoError(err, "Login with valid token error", err)

		time.Sleep(3 * time.Second)
		_, err = suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{})
		suite.Assert().ErrorIs(err, pb.ErrUserNeedAuthentication)
	})
}

func (suite *AuthSuite) getEnvs() error {
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
