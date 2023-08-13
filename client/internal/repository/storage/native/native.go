package native

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/99designs/keyring"

	"github.com/Karzoug/goph_keeper/client/internal/model"
	"github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

const appKey = "GOPHKEEPER"

type nativeStorage struct {
	keyring keyring.Keyring
}

func New(t storage.Type) (*nativeStorage, error) {
	const op = "create native storage"

	ring, err := keyring.Open(keyring.Config{
		AllowedBackends: []keyring.BackendType{keyring.BackendType(t)},
	})
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	return &nativeStorage{
		keyring: ring,
	}, nil
}

func (ns nativeStorage) SetCredentials(ctx context.Context, creds model.Credentials) error {
	const op = "native: set credentials"

	b := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(b)
	if err := enc.Encode(creds); err != nil {
		return e.Wrap(op, err)
	}

	if err := ns.keyring.Set(keyring.Item{
		Key:  appKey,
		Data: b.Bytes(),
	}); err != nil {
		return e.Wrap(op, err)
	}

	return nil
}

func (ns nativeStorage) GetCredentials(context.Context) (model.Credentials, error) {
	const op = "native: get credentials"

	var res model.Credentials

	item, err := ns.keyring.Get(appKey)
	if err != nil {
		return res, e.Wrap(op, err)
	}

	b := bytes.NewBuffer(item.Data)
	dec := gob.NewDecoder(b)
	if err := dec.Decode(&res); err != nil {
		return res, e.Wrap(op, err)
	}

	return res, nil
}
func (ns nativeStorage) DeleteCredentials(context.Context) error {
	const op = "native: delete credentials"

	if err := ns.keyring.Remove(appKey); err != nil {
		return e.Wrap(op, err)
	}

	return nil
}
