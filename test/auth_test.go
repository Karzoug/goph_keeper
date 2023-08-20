package main

// Basic imports
import (
	"context"
	"crypto/rand"
	"time"

	"github.com/pioz/faker"

	pb "github.com/Karzoug/goph_keeper/common/grpc/server"
)

type AuthSuite struct {
	commonTestSuite
}

func (suite *AuthSuite) TestAuth() {
	ctx, cancel := context.WithTimeout(context.Background(), testsTimeout)
	defer cancel()

	suite.Run("register: bad arguments", func() {
		authHash := make([]byte, 32)
		_, err := rand.Read(authHash)
		suite.Require().NoError(err, "Generate random auth hash error")

		// the same email
		_, err = suite.grpcClient.Register(ctx, &pb.RegisterRequest{
			Email: suite.email,
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
		_, err := suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: faker.FreeEmail(),
			Hash:  suite.authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrUserNotExists)

		// bad email
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: faker.Username() + "@superpupkinmupkin.io",
			Hash:  suite.authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidEmailFormat)

		// bad email
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: faker.Username() + "io.com",
			Hash:  suite.authHash,
		})
		suite.Assert().ErrorIs(err, pb.ErrInvalidEmailFormat)

		// empty email
		_, err = suite.grpcClient.Login(ctx, &pb.LoginRequest{
			Email: "",
			Hash:  suite.authHash,
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
		_, err := suite.grpcClient.ListVaultItems(ctx, &pb.ListVaultItemsRequest{})
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
			Email: suite.email,
			Hash:  suite.authHash,
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
