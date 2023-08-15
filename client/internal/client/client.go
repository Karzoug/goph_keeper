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
	"github.com/Karzoug/goph_keeper/client/internal/model"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage/native"
	sqlite "github.com/Karzoug/goph_keeper/client/internal/repository/storage/sqllite"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

const (
	createClientTimeout = 5 * time.Second
	syncTimeout         = 5 * time.Second
	syncInterval        = 5 * time.Minute
)

type clientCredentialsStorage interface {
	// SetCredentials adds or updates email, token and encryption key.
	SetCredentials(context.Context, model.Credentials) error
	//GetCredentials returns email and encryption key, or an error if they are not found.
	// It can also return a token if it exists.
	GetCredentials(context.Context) (model.Credentials, error)
	// DeleteCredentials deletes all credentials: email, token and encryption key.
	DeleteCredentials(context.Context) error
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
	closedCh           chan struct{}
}

func New(cfg *config.Config, logger *slog.Logger) (*Client, error) {
	const op = "create client"

	ctx, cancel := context.WithTimeout(context.Background(), createClientTimeout)
	defer cancel()

	c := &Client{
		cfg:      cfg,
		logger:   logger.With(slog.String("from", "client")),
		closedCh: make(chan struct{}),
	}

	ss, err := sqlite.New(ctx)
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	c.storage = ss

	c.credentialsStorage = ss
	if cfg.CredentialsStorageType != storage.Database {
		ns, err := native.New(cfg.CredentialsStorageType)
		if err != nil {
			return nil, e.Wrap(op, err)
		}
		c.credentialsStorage = ns
	}

	if err := c.restoreCredentials(ctx); err != nil {
		c.logger.Error(op, err)
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

	go c.runSyncLoop()

	return c, nil
}

func (c *Client) runSyncLoop() {
	ticker := time.NewTicker(syncInterval)
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithTimeout(context.TODO(), syncTimeout)
		defer cancel()
		_ = c.SyncVaultItems(ctx)

		select {
		case <-c.closedCh:
			return
		case <-ticker.C:
		}
	}
}

func (c *Client) Version() string {
	return c.cfg.Version
}

func (c *Client) RootPath() string {
	return c.cfg.RootPath
}

func (c *Client) Close() error {
	const op = "close client"

	close(c.closedCh)
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
