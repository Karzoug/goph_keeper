package badger

import (
	"github.com/dgraph-io/badger/v3"

	"github.com/Karzoug/goph_keeper/pkg/e"
)

const dbFilename = "gophkeeper.db"

type Storage struct {
	db *badger.DB
}

func New() (*Storage, error) {
	const op = "create badger storage"

	db, err := badger.Open(badger.DefaultOptions(dbFilename).WithLoggingLevel(badger.ERROR))
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Close() error {
	const op = "badger: close"

	return e.Wrap(op, s.db.Close())
}
