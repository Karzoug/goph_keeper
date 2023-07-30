package client

import (
	"context"
	"errors"

	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/metadata"

	"github.com/Karzoug/goph_keeper/client/internal/config"
	"github.com/Karzoug/goph_keeper/client/internal/model/auth"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage/badger"
	"github.com/Karzoug/goph_keeper/client/pkg/crypto"
	pb "github.com/Karzoug/goph_keeper/common/grpc"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

type credentialsStorage interface {
	SetCredentials(email, token string, encrKey []byte) error
	GetCredentials() (email, token string, encrKey []byte, err error)
	DeleteCredentials() error
	Close() error
}

type storage interface {
	ListVaultItems() ([]vault.Item, error)
	ListVaultItemsNames() ([]string, error)
	SetVaultItem(item vault.Item) error
	SetVaultItems(items []vault.Item) error
	Close() error
}

type credentials struct {
	email    string
	authHash auth.Hash
	token    string
	encrKey  vault.EncryptionKey
}

type Client struct {
	cfg    *config.Config
	logger *slog.Logger

	storage            storage
	credentialsStorage credentialsStorage
	credentials        credentials
	conn               *grpc.ClientConn
	grpcClient         pb.GophKeeperServiceClient
}

func New(cfg *config.Config, logger *slog.Logger) (*Client, error) {
	const op = "create client"

	c := &Client{
		cfg:    cfg,
		logger: logger,
	}

	// TODO: add other storages, add other credentials storages
	bs, err := badger.New()
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	c.storage = bs
	c.credentialsStorage = bs

	_ = c.getCredentials()

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

func (c *Client) HasLocalCredintials() bool {
	return len(c.credentials.email) > 0 && c.credentials.encrKey != nil
}

func (c *Client) HasToken() bool {
	return len(c.credentials.token) > 0
}

func (c *Client) setPasswordHashes(email string, password []byte) error {
	const op = "set password hashes"

	defer crypto.Wipe(password) // prevent long-term storage of the password in memory

	hash, err := auth.NewHash([]byte(email), password)
	if err != nil {
		return e.Wrap(op, err)
	}

	encrKey, err := vault.NewEncryptionKey([]byte(email), password)
	if err != nil {
		return e.Wrap(op, err)
	}

	c.credentials = credentials{
		email:    email,
		encrKey:  encrKey,
		authHash: hash,
	}

	return e.Wrap(op,
		c.credentialsStorage.SetCredentials(email, "", encrKey))
}

func (c *Client) setToken(token string) error {
	const op = "set token"

	c.credentials.authHash = nil
	c.credentials.token = token

	if !c.HasLocalCredintials() {
		return e.Wrap(op, errors.New("there aren't local credentials"))
	}

	return e.Wrap(op,
		c.credentialsStorage.SetCredentials(c.credentials.email, token, c.credentials.encrKey))
}

func (c *Client) clearCredentials(ctx context.Context) error {
	const op = "clear credentials"

	c.credentials = credentials{}

	return e.Wrap(op,
		c.credentialsStorage.DeleteCredentials())
}

func (c *Client) getCredentials() bool {
	email, token, encrKey, err := c.credentialsStorage.GetCredentials()
	if err != nil {
		return false
	}

	if len(email) == 0 || len(encrKey) == 0 {
		return false
	}
	c.credentials = credentials{
		email:   email,
		token:   token,
		encrKey: encrKey,
	}

	return true
}

func (c *Client) newContextWithAuthData(ctx context.Context) (context.Context, error) {
	if len(c.credentials.token) == 0 {
		return ctx, pb.ErrEmptyAuthData
	}
	md := metadata.New(map[string]string{"token": c.credentials.token})
	return metadata.NewOutgoingContext(ctx, md), nil
}
