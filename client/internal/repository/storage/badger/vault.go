package badger

import (
	"bytes"

	"github.com/dgraph-io/badger/v3"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

func (s *Storage) ListVaultItems() ([]vault.Item, error) {
	const op = "badger: list vault items"

	res := make([]vault.Item, 0)

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			dbItem := it.Item()
			k := dbItem.Key()
			if bytes.HasPrefix(k, emailDBKey) || bytes.HasPrefix(k, tokenDBKey) || bytes.HasPrefix(k, encrKeyDBKey) {
				continue
			}
			var item vault.Item
			err := dbItem.Value(item.UnmarshalBinary)
			if err != nil {
				return err
			}
			res = append(res, item)
		}
		return nil
	})
	return res, e.Wrap(op, err)
}

func (s *Storage) ListVaultItemsNames() ([]string, error) {
	const op = "badger: list vault items names"

	res := make([]string, 0)

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			if bytes.HasPrefix(k, emailDBKey) || bytes.HasPrefix(k, tokenDBKey) || bytes.HasPrefix(k, encrKeyDBKey) {
				continue
			}
			res = append(res, string(k)) // []byte -> string, it's copy
		}
		return nil
	})
	return res, e.Wrap(op, err)
}

func (s *Storage) SetVaultItem(item vault.Item) error {
	const op = "badger: set vault item"

	err := s.db.Update(func(txn *badger.Txn) error {
		b, err := item.MarshalBinary()
		if err != nil {
			return err
		}
		return txn.Set([]byte(item.Name), b)
	})
	return e.Wrap(op, err)
}

func (s *Storage) SetVaultItems(items []vault.Item) error {
	const op = "badger: set vault items"

	err := s.db.Update(func(txn *badger.Txn) error {
		for i := 0; i < len(items); i++ {
			b, err := items[i].MarshalBinary()
			if err != nil {
				return err
			}
			err = txn.Set([]byte(items[i].Name), b)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return e.Wrap(op, err)
}
