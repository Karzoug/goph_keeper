package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/e"
	serr "github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

func (s *storage) SetVaultItem(ctx context.Context, email string, item vault.Item) error {
	const op = "sqlite: set vault item"

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO vaults(email,name,value,updated_at) VALUES(?, ?, ?, ?)
		ON CONFLICT(email,name) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at;`,
		email, item.Name, item.Value, item.UpdatedAt)
	if err != nil {
		return e.Wrap(op, err)
	}

	return nil
}
func (s *storage) LastUpdateVault(ctx context.Context, email string) (time.Time, error) {
	const op = "sqlite: last update vault"

	var t time.Time
	err := s.db.QueryRowContext(ctx,
		`SELECT MAX(updated_at) FROM vaults WHERE email = ?;`, email).
		Scan(&t)

	if err != nil {
		if err == sql.ErrNoRows {
			return t, e.Wrap(op, serr.ErrRecordNotFound)
		}
		return t, e.Wrap(op, err)
	}

	return t, nil
}

func (s *storage) DeleteVaultItem(ctx context.Context, email string, name string) error {
	const op = "sqlite: delete vault item"

	res, err := s.db.ExecContext(ctx,
		`DELETE FROM vaults WHERE email = ? AND name = ?;`,
		email, name)
	if err != nil {
		return e.Wrap(op, err)
	}

	count, err := res.RowsAffected() // driver specific
	if err != nil {
		return e.Wrap(op, err)
	}
	if count == 0 {
		return e.Wrap(op, serr.ErrNoRecordsAffected)
	}

	return nil
}
func (s *storage) ListVaultItems(ctx context.Context, email string, since *time.Time) ([]vault.Item, error) {
	const op = "sqlite: list vault items"

	var (
		rows *sql.Rows
		err  error
	)
	if since != nil {
		rows, err = s.db.QueryContext(ctx,
			`SELECT name, value, updated_at FROM vaults WHERE email = ? AND updated_at > ?;`, email, since)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT name, value, updated_at FROM vaults WHERE email = ?`, email)
	}

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res := make([]vault.Item, 0)
	for rows.Next() {
		var item vault.Item
		err := rows.Scan(&item.Name, &item.Value, &item.UpdatedAt)
		if err != nil {
			return nil, e.Wrap(op, err)
		}
		res = append(res, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	return res, nil
}
