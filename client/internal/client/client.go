package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"
	"time"

	"log/slog"

	"google.golang.org/grpc"
	gcreds "google.golang.org/grpc/credentials"

	"github.com/Karzoug/goph_keeper/client/internal/config"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	sqlite "github.com/Karzoug/goph_keeper/client/internal/repository/storage/sqllite"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

const createClientTimeout = 5 * time.Second

type clientCredentialsStorage interface {
	// SetCredentials adds or updates email, token and encryption key.
	SetCredentials(ctx context.Context, email, token, encrKey string) error
	//GetCredentials returns email and encryption key, or an error if they are not found.
	// It can also return a token if it exists.
	GetCredentials(context.Context) (email, token, encrKey string, err error)
	// DeleteCredentials deletes all credentials: email, token and encryption key.
	DeleteCredentials(context.Context) error
	// Close closes storage if applicable.
	Close() error
}

type clientStorage interface {
	GetOwner(ctx context.Context) (string, error)
	SetOwner(ctx context.Context, email string) error
	ClearVault(ctx context.Context) error

	ListVaultItems(context.Context) ([]vault.Item, error)
	ListVaultItemsIDName(context.Context) ([]vault.IDName, error)
	ListModifiedVaultItems(context.Context) ([]vault.Item, error)
	GetVaultItem(ctx context.Context, id string) (vault.Item, error)
	SetVaultItem(ctx context.Context, item vault.Item) error
	MoveVaultItemToConflict(ctx context.Context, id string) error
	GetLastServerUpdatedAt(ctx context.Context) (int64, error)

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

	ctx, cancel := context.WithTimeout(context.Background(), createClientTimeout)
	defer cancel()

	c := &Client{
		cfg:    cfg,
		logger: logger.With(slog.String("from", "client")),
	}

	ss, err := sqlite.New(ctx)
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	c.credentialsStorage = ss // TODO: add other credentials storages
	c.storage = ss

	if err := c.restoreCredentials(ctx); err != nil {
		c.logger.Debug(op, err)
	}

	cs, err := loadTLSCredentials(cfg.Host, cfg.CertFilename)
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	addr := cfg.Host + ":" + cfg.Port
	c.conn, err = grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(cs))
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

func loadTLSCredentials(host, certFilename string) (gcreds.TransportCredentials, error) {
	const op = "load TLS credentials"

	config := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS13,
	}

	if len(certFilename) == 0 {
		return gcreds.NewTLS(config), nil
	}

	certPool := x509.NewCertPool()
	f, err := os.Open(certFilename)
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	if !certPool.AppendCertsFromPEM(b) {
		return nil, e.Wrap(op, err)
	}
	config.RootCAs = certPool

	return gcreds.NewTLS(config), nil
}
