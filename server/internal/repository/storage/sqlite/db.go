package sqlite

import (
	"database/sql"

	_ "modernc.org/sqlite"

	"github.com/Karzoug/goph_keeper/pkg/e"
)

const duplicateKeyErrorCode = "1555"

type storage struct {
	db *sql.DB
}

func New(dsnURI string) (*storage, error) {
	op := "create sqlite storage"

	db, err := sql.Open("sqlite", dsnURI)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, e.Wrap(op, err)
	}

	return &storage{
		db: db,
	}, nil
}
