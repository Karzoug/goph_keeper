package client

import (
	"context"
	"errors"

	"github.com/Karzoug/goph_keeper/client/internal/model"
	"github.com/Karzoug/goph_keeper/client/internal/model/auth"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/client/pkg/crypto"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

type credentials struct {
	model.Credentials
	AuthHash auth.Hash
}

// HasLocalCredintials indicates whether credentials for the application to work locally.
func (c *Client) HasLocalCredintials() bool {
	return len(c.credentials.Email) > 0 && c.credentials.EncrKey.Hash != nil
}

// HasToken indicates whether token for the application to work online.
func (c *Client) HasToken() bool {
	return len(c.credentials.Token) > 0
}

// buildPasswordHashes builds auth hash and encryption key from given email and password.
//
// Warning(!): wipes given password slice to prevent long-term storage in memory.
func buildPasswordHashes(ctx context.Context, email string, password []byte) (auth.Hash, vault.EncryptionKey, error) {
	const op = "build password hashes"

	defer crypto.Wipe(password) // prevent long-term storage of the password in memory

	hash, err := auth.NewHash([]byte(email), password)
	if err != nil {
		return nil, vault.EncryptionKey{}, e.Wrap(op, err)
	}

	encrKey, err := vault.NewEncryptionKey([]byte(email), password)
	if err != nil {
		return nil, vault.EncryptionKey{}, e.Wrap(op, err)
	}

	return hash, encrKey, nil
}

// setCredentialsForOwnerOnly sets credentials if only the local vault (storage) owner email
// is equal the given email.
func (c *Client) setCredentialsForOwnerOnly(ctx context.Context, email string, hash auth.Hash, encrKey vault.EncryptionKey) error {
	const op = "set credentials for owner only"

	owner, err := c.storage.GetOwner(ctx)
	if err != nil {
		if !errors.Is(err, storage.ErrRecordNotFound) {
			return e.Wrap(op, err)
		}

		// case: owner db is not set
		err := c.storage.ClearVault(ctx)
		if err != nil {
			return e.Wrap(op, err)
		}
		err = c.storage.SetOwner(ctx, email)
		if err != nil {
			return e.Wrap(op, err)
		}
		owner = email
	}

	if owner != email {
		return ErrUserNeedAuthentication
	}

	c.credentials = credentials{
		Credentials: model.Credentials{
			Email:   email,
			EncrKey: encrKey,
		},
		AuthHash: hash,
	}

	if err := c.credentialsStorage.SetCredentials(ctx, c.credentials.Credentials); err != nil {
		return e.Wrap(op, err)
	}

	return nil
}

// setCredentialsForced sets credentials.
//
// Warning(!): if the local vault (storage) owner email is not equal the given email
// method clear all data in storage.
func (c *Client) setCredentialsForced(ctx context.Context, email string, hash auth.Hash, encrKey vault.EncryptionKey) error {
	const op = "set credentials"

	owner, err := c.storage.GetOwner(ctx)
	if err != nil {
		if !errors.Is(err, storage.ErrRecordNotFound) {
			return e.Wrap(op, err)
		}
		owner = ""
	}

	if owner != email {
		err := c.storage.ClearVault(ctx)
		if err != nil {
			return e.Wrap(op, err)
		}
		err = c.storage.SetOwner(ctx, email)
		if err != nil {
			return e.Wrap(op, err)
		}
	}

	c.credentials = credentials{
		Credentials: model.Credentials{
			Email:   email,
			EncrKey: encrKey,
		},
		AuthHash: hash,
	}

	if err := c.credentialsStorage.SetCredentials(ctx, c.credentials.Credentials); err != nil {
		return e.Wrap(op, err)
	}

	return nil
}

func (c *Client) setToken(ctx context.Context, token string) error {
	const op = "set token"

	c.credentials.AuthHash = nil
	c.credentials.Token = token

	if !c.HasLocalCredintials() {
		return e.Wrap(op, ErrUserNeedAuthentication)
	}

	return e.Wrap(op, c.credentialsStorage.SetCredentials(ctx, c.credentials.Credentials))
}

func (c *Client) clearToken(ctx context.Context) error {
	const op = "clear token"

	c.credentials.Token = ""

	return e.Wrap(op,
		c.credentialsStorage.SetCredentials(ctx, c.credentials.Credentials))
}

func (c *Client) clearCredentials(ctx context.Context) error {
	const op = "clear credentials"

	c.credentials = credentials{}

	return e.Wrap(op,
		c.credentialsStorage.DeleteCredentials(ctx))
}

func (c *Client) restoreCredentials(ctx context.Context) error {
	const op = "restore credentials"

	creds, err := c.credentialsStorage.GetCredentials(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return nil
		}
		return e.Wrap(op, err)
	}
	// encrKey, err := vault.EncryptionKeyFromString(encrKeyString)
	// if err != nil {
	// 	return e.Wrap(op, err)
	// }

	c.credentials = credentials{
		Credentials: creds,
	}
	return nil
}
