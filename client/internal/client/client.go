package client

import (
	"context"
	"time"

	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Karzoug/goph_keeper/client/internal/config"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage/badger"
	"github.com/Karzoug/goph_keeper/client/pkg/crypto"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

type clientCredentialsStorage interface {
	// SetCredentials adds or updates email, token and encryption key.
	SetCredentials(email, token string, encrKey []byte) error
	//GetCredentials returns email and encryption key, or an error if they are not found.
	// It can also return a token if it exists.
	GetCredentials() (email, token string, encrKey []byte, err error)
	// DeleteCredentials deletes all credentials: email, token and encryption key.
	DeleteCredentials() error
	// Close closes storage if applicable.
	Close() error
}

type clientStorage interface {
	GetOwner(ctx context.Context) (string, error)
	SetOwner(ctx context.Context, email string) error
	ClearVault(ctx context.Context) error

	ListVaultItems(context.Context) ([]vault.Item, error)
	ListVaultItemsNames(context.Context) ([]string, error)
	ListModifiedVaultItems(context.Context) ([]vault.Item, error)
	GetVaultItem(ctx context.Context, name string) (vault.Item, error)
	SetVaultItem(ctx context.Context, item vault.Item) error
	MoveVaultItemToConflict(ctx context.Context, name string) error
	GetLastServerUpdatedAt(ctx context.Context) (time.Time, error)

	Close() error
}

type Client struct {
	cfg    *config.Config
	logger *slog.Logger

	storage            clientStorage
	credentialsStorage clientCredentialsStorage
	credentials        credentials
	conn               *grpc.ClientConn
	grpcClient         pb.GophKeeperServiceClient
}

func New(cfg *config.Config, logger *slog.Logger) (*Client, error) {
	const op = "create client"

	c := &Client{
		cfg:    cfg,
		logger: logger.With(slog.String("from", "client")),
	}

	// TODO: add other storages, add other credentials storages
	bs, err := badger.New()
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	c.credentialsStorage = ss // TODO: add other credentials storages
	c.storage = ss

	_ = c.restoreCredentials()

	// TODO: add TLS
	c.conn, err = grpc.Dial(c.cfg.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	c.grpcClient = pb.NewGophKeeperServiceClient(c.conn)

	return c, nil
}

func (c *Client) Version() string {
	return c.cfg.Version
}

func (c *Client) Close() error {
	const op = "close client"

	return e.Wrap(op, c.storage.Close())
}
