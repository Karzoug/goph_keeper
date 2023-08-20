package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/Karzoug/goph_keeper/common/grpc/server"
)

const (
	testsTimeout    time.Duration = 120 * time.Second
	emailCodeLength               = 6
)

type client struct {
	conn       *grpc.ClientConn
	grpcClient pb.GophKeeperServerClient
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestAuth(t *testing.T) {
	suite.Run(t, new(AuthSuite))
}

func TestVault(t *testing.T) {
	suite.Run(t, new(VaultSuite))
}

func TestSyncVault(t *testing.T) {
	suite.Run(t, new(SyncVaultSuite))
}

func newContextWithAuthData(ctx context.Context, token string) context.Context {
	md := metadata.New(map[string]string{"token": token})
	return metadata.NewOutgoingContext(ctx, md)
}
