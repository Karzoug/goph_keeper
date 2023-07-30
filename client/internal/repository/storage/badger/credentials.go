package badger

import (
	"errors"

	"github.com/dgraph-io/badger/v3"
)

var (
	emailDBKey   = []byte("GOPH_KEEPER_CREDS_EMAIL")
	tokenDBKey   = []byte("GOPH_KEEPER_CREDS_TOKEN")
	encrKeyDBKey = []byte("GOPH_KEEPER_CREDS_ENCRKEY")
)

func (s *Storage) SetCredentials(email, token string, encrKey []byte) error {
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
	return err
}
func (s *Storage) GetCredentials() (email, token string, encrKey []byte, err error) {
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
		return "", "", nil, err
	}

	return
}
func (s *Storage) DeleteCredentials() error {
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
	return err
}
