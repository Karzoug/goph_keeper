package badger

import (
	"errors"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/dgraph-io/badger/v3"
)

var (
	emailDBKey   = append(appDataDBKeyPreffix, []byte("CREDS_EMAIL")...)
	tokenDBKey   = append(appDataDBKeyPreffix, []byte("CREDS_TOKEN")...)
	encrKeyDBKey = append(appDataDBKeyPreffix, []byte("CREDS_ENCRKEY")...)
)

func (s *Storage) SetCredentials(email, token string, encrKey []byte) error {
	const op = "badger: set credentials"

	err := s.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(emailDBKey, []byte(email))
		if err != nil {
			return err
		}
		err = txn.Set(tokenDBKey, []byte(token))
		if err != nil {
			return err
		}
		err = txn.Set(encrKeyDBKey, encrKey)
		if err != nil {
			return err
		}
		return nil
	})
	return e.Wrap(op, err)
}

func (s *Storage) GetCredentials() (email, token string, encrKey []byte, err error) {
	const op = "badger: get credentials"

	err = s.db.View(func(txn *badger.Txn) error {
		emailItem, err := txn.Get(emailDBKey)
		if err != nil {
			return err
		}
		emailBytes, err := emailItem.ValueCopy(nil)
		if err != nil {
			return err
		}
		email = string(emailBytes)

		encrKeyItem, err := txn.Get(encrKeyDBKey)
		if err != nil {
			return err
		}
		encrKey, err = encrKeyItem.ValueCopy(nil)
		if err != nil {
			return err
		}

		tokenItem, err := txn.Get(tokenDBKey)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			return err
		}
		tokenBytes, err := tokenItem.ValueCopy(nil)
		if err != nil {
			return err
		}
		token = string(tokenBytes)
		return nil
	})

	if err != nil {
		return "", "", nil, e.Wrap(op, err)
	}

	return
}
func (s *Storage) DeleteCredentials() error {
	const op = "badger: delete credentials"

	err := s.db.View(func(txn *badger.Txn) error {
		err := txn.Delete(emailDBKey)
		if err != nil {
			return err
		}
		err = txn.Delete(tokenDBKey)
		if err != nil {
			return err
		}
		err = txn.Delete(encrKeyDBKey)
		if err != nil {
			return err
		}
		return nil
	})
	return e.Wrap(op, err)
}
