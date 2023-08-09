package sqlite

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"

	"github.com/Karzoug/goph_keeper/pkg/e"
	sconfig "github.com/Karzoug/goph_keeper/server/internal/config/storage"
)

const (
	// SQLite uses the "file:" URI syntax to identify database files,
	// see: https://www.sqlite.org/uri.html
	URIPreffix            = "file:"
	duplicateKeyErrorCode = "1555"
)

type storage struct {
	db *sql.DB
}

func New(ctx context.Context, cfg sconfig.Config) (*storage, error) {
	op := "create sqlite storage"

	db, err := sql.Open("sqlite", cfg.URI)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, e.Wrap(op, err)
	}

	return &storage{
		db: db,
	}, nil
}

func (s *storage) Close() error {
	const op = "sqlite: close"

	return e.Wrap(op, s.db.Close())
}
