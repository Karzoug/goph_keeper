package sqlite

import (
	"context"
	"database/sql"
	"errors"

	serr "github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

const ownerDBKey = "OWNER"

func (s *storage) GetOwner(ctx context.Context) (string, error) {
	const op = "sqlite: get owner"

	var dbOwner string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM app WHERE key = ?;`, ownerDBKey).Scan(&dbOwner)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", e.Wrap(op, err)
		}
		return "", serr.ErrRecordNotFound
	}
	return dbOwner, nil
}

func (s *storage) SetOwner(ctx context.Context, email string) error {
	const op = "sqlite: set owner"

	res, err := s.db.ExecContext(ctx, `INSERT INTO app(key,value) VALUES(?, ?) 
	ON CONFLICT(key) 
	DO UPDATE SET value = excluded.value;`, ownerDBKey, email)
	if err != nil {
		return e.Wrap(op, err)
	}

	count, err := res.RowsAffected()
	if err != nil {
		return e.Wrap(op, err)
	}
	if count == 0 {
		return serr.ErrNoRecordsAffected
	}

	return nil
}

func (s *storage) ClearVault(ctx context.Context) error {
	const op = "sqlite: clear vault"

	_, err := s.db.ExecContext(ctx, `DELETE FROM vaults; DELETE FROM conflict_vaults;`)
	if err != nil {
		return e.Wrap(op, err)
	}

	return nil
}
