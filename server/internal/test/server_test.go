package test

// import (
// 	"context"
// 	"crypto/rand"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"reflect"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/buger/jsonparser"
// 	"github.com/caarlos0/env/v9"
// 	"github.com/golang-migrate/migrate/v4"
// 	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
// 	_ "github.com/golang-migrate/migrate/v4/source/file"
// 	"github.com/redis/go-redis/v9"
// 	"github.com/stretchr/testify/require"
// 	"github.com/stretchr/testify/suite"
// 	"google.golang.org/grpc"

// 	pb "github.com/Karzoug/goph_keeper/common/grpc"
// 	"github.com/Karzoug/goph_keeper/pkg/logger/slog/discard"
// 	"github.com/Karzoug/goph_keeper/server/internal/config"
// 	dgrpc "github.com/Karzoug/goph_keeper/server/internal/delivery/grpc"
// 	rtasks "github.com/Karzoug/goph_keeper/server/internal/delivery/rtask"
// 	"github.com/Karzoug/goph_keeper/server/internal/repository/mail/smtp"
// 	rtaskc "github.com/Karzoug/goph_keeper/server/internal/repository/rtask"
// 	"github.com/Karzoug/goph_keeper/server/internal/repository/storage/sqlite"
// 	"github.com/Karzoug/goph_keeper/server/internal/service"
// )

// type ServerTestSuite struct {
// 	suite.Suite
// 	parentCtx       context.Context
// 	cancel          context.CancelFunc
// 	m               *migrate.Migrate
// 	storageFilename string
// 	redisURI        string
// 	address         string
// 	mailhogAddr     string
// 	cfg             *config.Config
// }

// func (suite *ServerTestSuite) SetupSuite() {
// 	logger := discard.NewDiscardLogger()

// 	suite.storageFilename = suite.T().TempDir() + "/temp.db"
// 	suite.address = "127.0.0.1:8080"
// 	suite.redisURI = "redis://localhost:6379/3"
// 	suite.mailhogAddr = "127.0.0.1:8025"

// 	suite.T().Setenv("GOPHKEEPER_SERVICE_TOKEN_SECRET_KEY", "H#d0xYP2KTk7iZo8O*")
// 	suite.T().Setenv("GOPHKEEPER_GRPC_ADDRESS", suite.address)
// 	suite.T().Setenv("GOPHKEEPER_SERVICE_STORAGE_URI", "file:"+suite.storageFilename)
// 	suite.T().Setenv("GOPHKEEPER_RTASK_STORAGE_URI", suite.redisURI)
// 	suite.T().Setenv("GOPHKEEPER_SMTP_HOST", "0.0.0.0")
// 	suite.T().Setenv("GOPHKEEPER_SMTP_PORT", "1025")

// 	cfg, err := buildConfig()
// 	require.NoError(suite.T(), err)
// 	suite.cfg = cfg

// 	storage, err := sqlite.New(cfg.Service.Storage)
// 	require.NoError(suite.T(), err)

// 	suite.m, err = migrate.New(
// 		"file://./../../migrations",
// 		"sqlite://"+suite.storageFilename)
// 	require.NoError(suite.T(), err)

// 	rtaskClient, err := rtaskc.New(cfg.RTask.Storage.URI, logger)
// 	require.NoError(suite.T(), err)

// 	smtpClient, err := smtp.New(cfg.SMTP)
// 	require.NoError(suite.T(), err)

// 	s, err := service.New(cfg.Service, storage, rtaskClient, smtpClient, service.WithSLogger(logger))
// 	require.NoError(suite.T(), err)

// 	grpcServer := dgrpc.New(cfg.GRPC, s, logger)
// 	rtaskServer, err := rtasks.New(cfg.RTask, s, logger)

// 	suite.parentCtx, suite.cancel = context.WithTimeout(context.Background(), 30*time.Second)

// 	go grpcServer.Run(suite.parentCtx)
// 	go rtaskServer.Run(suite.parentCtx)

// 	// wait to start
// 	err = WaitPort(suite.parentCtx, "tcp", "8080")
// 	require.NoError(suite.T(), err)
// 	err = WaitPort(suite.parentCtx, "tcp", "6379")
// 	require.NoError(suite.T(), err)
// }

// func (suite *ServerTestSuite) SetupTest() {
// 	suite.m.Up()
// }

// func (suite *ServerTestSuite) TearDownTest() {
// 	suite.m.Down()

// 	opt, err := redis.ParseURL(suite.redisURI)
// 	suite.Require().NoError(err)
// 	cl := redis.NewClient(opt)
// 	cl.FlushDB(suite.parentCtx)
// 	cl.Close()

// 	req, err := http.NewRequestWithContext(suite.parentCtx,
// 		http.MethodDelete,
// 		fmt.Sprintf("http://%s/api/v1/messages", suite.mailhogAddr),
// 		http.NoBody,
// 	)
// 	resp, err := http.DefaultClient.Do(req)
// 	require.NoError(suite.T(), err)
// 	defer resp.Body.Close()
// }

// func (suite *ServerTestSuite) TearDownSuite() {
// 	suite.cancel()
// }

// func TestServerTestSuite(t *testing.T) {
// 	suite.Run(t, new(ServerTestSuite))
// }

// func (suite *ServerTestSuite) TestRegister() {
// 	ctx, cancel := context.WithTimeout(suite.parentCtx, 5*time.Second)
// 	defer cancel()

// 	conn, err := grpc.Dial(suite.address, grpc.WithInsecure(), grpc.WithBlock())
// 	require.NoError(suite.T(), err)
// 	defer conn.Close()

// 	grpcClient := pb.NewGophKeeperServiceClient(conn)

// 	// generate random auth hash
// 	authHash := make([]byte, 32)
// 	_, err = rand.Read(authHash)
// 	require.NoError(suite.T(), err)

// 	email := "somebody@some.com"

// 	// register
// 	_, err = grpcClient.Register(ctx, &pb.RegisterRequest{
// 		Email: email,
// 		Hash:  authHash,
// 	})
// 	suite.NoError(err)

// 	// try to login, email not confirmed
// 	_, err = grpcClient.Login(ctx, &pb.LoginRequest{
// 		Email: email,
// 		Hash:  authHash,
// 	})
// 	require.Error(suite.T(), err)
// 	require.ErrorIs(suite.T(), err, pb.ErrUserEmailNotVerified)

// 	getBodyLastMail := func(toEmail string) ([]byte, error) {
// 		resp, err := http.Get(fmt.Sprintf("http://%s/api/v2/search?kind=to&query=%s&limit=1", suite.mailhogAddr, toEmail))
// 		if err != nil {
// 			return nil, err
// 		}
// 		defer resp.Body.Close()

// 		if resp.StatusCode != http.StatusOK {
// 			return nil, errors.New("code not found")
// 		}
// 		body, err := io.ReadAll(resp.Body)
// 		if err != nil {
// 			return nil, err
// 		}

// 		return body, nil
// 	}

// 	// waiting for the mail
// 	var body []byte
// 	for attempt := 0; attempt < 20; attempt++ {
// 		body, err = getBodyLastMail(email)
// 		if err != nil {
// 			time.Sleep(500 * time.Millisecond)
// 			continue
// 		}

// 		count, err := jsonparser.GetInt(body, "count")
// 		if err != nil || count == 0 {
// 			time.Sleep(500 * time.Millisecond)
// 			continue
// 		}
// 	}
// 	require.NoError(suite.T(), err)
// 	html, err := jsonparser.GetString(body, "items", "[0]", "Content", "Body")
// 	require.NoError(suite.T(), err)

// 	// looking for the verification code in the mail

//// TODO: change to search in text version

// 	substr := `id=3D"email_code">`
// 	pos := strings.Index(html, substr)
// 	require.NotEqual(suite.T(), -1, pos)
// 	codeInMail := html[pos+len(substr) : pos+len(substr)+suite.cfg.Service.Email.CodeLength]

// 	// login with the verification code, got token
// 	resp, err := grpcClient.Login(ctx, &pb.LoginRequest{
// 		Email:     email,
// 		Hash:      authHash,
// 		EmailCode: codeInMail,
// 	})
// 	require.NoError(suite.T(), err)
// 	require.NotEmpty(suite.T(), resp.Token)
// }

// func buildConfig() (*config.Config, error) {
// 	cfg := new(config.Config)

// 	opts := env.Options{
// 		Prefix: "GOPHKEEPER_",
// 		FuncMap: map[reflect.Type]env.ParserFunc{
// 			reflect.TypeOf(cfg.Env): config.EnvTypeParserFunc,
// 		},
// 	}

// 	return cfg, env.ParseWithOptions(cfg, opts)
// }
