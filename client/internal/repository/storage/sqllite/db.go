package sqlite

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"

	"github.com/Karzoug/goph_keeper/client/migrations"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

const (
	duplicateKeyErrorCode = "1555"
	dbFilename            = "vault.db"
)

type storage struct {
	db *sql.DB
}

func New() (*storage, error) {
	op := "create sqlite storage"

	db, err := sql.Open("sqlite", dbFilename)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, e.Wrap(op, err)
	}

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	d, err := iofs.New(migrations.FS, "sql")
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	m, err := migrate.NewWithInstance("iofs", d, "sqlite", driver)
	if err != nil {
		return nil, e.Wrap(op, err)
	}
	_ = m.Up()

	return &storage{
		db: db,
	}, nil
}

func (s *storage) Close() error {
	const op = "sqlite: close"

	return e.Wrap(op, s.db.Close())
}
