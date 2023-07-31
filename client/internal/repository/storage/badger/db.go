package badger

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v3"

	"github.com/Karzoug/goph_keeper/pkg/e"
)

const dbFilename = "vault.db"

var (
	appDataDBKeyPreffix = []byte("GOPH_KEEPER_")
	ownerDBKey          = append(appDataDBKeyPreffix, []byte("OWNER")...)
)

type Storage struct {
	db *badger.DB
}

func New() (*Storage, error) {
	const op = "create badger storage"

	db, err := badger.Open(badger.DefaultOptions(dbFilename).WithLoggingLevel(badger.ERROR))
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	s := &Storage{
		db:      db,
	}

	return s, nil
	}

func (s *Storage) VerifyOwnerOrClearData(_ context.Context, email string) error {
	const op = "badger: verify owner or clear data"

	err := s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(ownerDBKey)
		if err != nil {
			return err
		}

		// if email equals email in db then return
		if err := item.Value(func(val []byte) error {
			if bytes.Equal(val, []byte(email)) {
				return nil
			}
			return errors.New("not the same owner")
		}); err == nil {
			return nil
		}

		// otherwise: delete all data except creds
		err = txn.Delete(ownerDBKey)
		if err != nil {
			return err
		}

		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			key := it.Item().KeyCopy(nil)
			if bytes.HasPrefix(key, appDataDBKeyPreffix) {
				continue
			}
			err := txn.Delete(key)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return e.Wrap(op, err)
}

func (s *Storage) Close() error {
	const op = "badger: close"

	return e.Wrap(op, s.db.Close())
}
