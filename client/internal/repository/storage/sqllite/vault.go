package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	serr "github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

func (s *storage) ListVaultItems(ctx context.Context) ([]vault.Item, error) {
	const op = "sqlite: list vault items"

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, value, client_updated_at, server_updated_at FROM vaults`)

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res := make([]vault.Item, 0)
	for rows.Next() {
		var item vault.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Value, &item.ClientUpdatedAt, &item.ServerUpdatedAt)
		if err != nil {
			return nil, e.Wrap(op, err)
		}
		res = append(res, item)
	}

	if err = rows.Err(); err != nil {
		return nil, e.Wrap(op, err)
	}

	return res, nil
}
func (s *storage) ListVaultItemsNames(ctx context.Context) ([]string, error) {
	const op = "sqlite: list vault items names"

	rows, err := s.db.QueryContext(ctx, `SELECT name FROM vaults`)

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res := make([]string, 0)
	for rows.Next() {
		var item string
		err := rows.Scan(&item)
		if err != nil {
			return nil, e.Wrap(op, err)
		}
		res = append(res, item)
	}

	if err = rows.Err(); err != nil {
		return nil, e.Wrap(op, err)
	}

	return res, nil
}

func (s *storage) ListModifiedVaultItems(ctx context.Context) ([]vault.Item, error) {
	const op = "sqlite: list modified vault items"

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, value, client_updated_at, server_updated_at 
		FROM vaults 
		WHERE server_updated_at < client_updated_at;`)

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res := make([]vault.Item, 0)
	for rows.Next() {
		var item vault.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Value, &item.ClientUpdatedAt, &item.ServerUpdatedAt)
		if err != nil {
			return nil, e.Wrap(op, err)
		}
		res = append(res, item)
	}

	if err = rows.Err(); err != nil {
		return nil, e.Wrap(op, err)
	}

	return res, nil
}

func (s *storage) GetVaultItem(ctx context.Context, name string) (vault.Item, error) {
	const op = "sqlite: get vault item"

	var item vault.Item
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, type, value, client_updated_at, server_updated_at 
		FROM vaults 
		WHERE name = ?`, name).
		Scan(&item.ID, &item.Name, &item.Type, &item.Value, &item.ClientUpdatedAt, &item.ServerUpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, serr.ErrRecordNotFound
		}
	}
	return item, e.Wrap(op, err)
}
func (s *storage) SetVaultItem(ctx context.Context, item vault.Item) error {
	const op = "sqlite: set vault item"

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO vaults(id,name,type,value,client_updated_at,server_updated_at) VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) 
		DO UPDATE SET id = excluded.id, type = excluded.type, value=excluded.value, client_updated_at=excluded.client_updated_at, server_updated_at=excluded.server_updated_at;`,
		item.ID, item.Name, item.Type, item.Value, item.ClientUpdatedAt, item.ServerUpdatedAt)
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
func (s *storage) MoveVaultItemToConflict(ctx context.Context, name string) error {
	const op = "sqlite: move vault item to conflict"

	tx, err := s.db.Begin()
	if err != nil {
		return e.Wrap(op, err)
	}
	defer tx.Rollback()

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO conflict_vaults (id,name,type,value,client_updated_at,server_updated_at)
		SELECT id,name,type,value,client_updated_at,server_updated_at
		FROM vaults WHERE name = ?;
		DELETE FROM vaults WHERE name = ?;`)

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
func (s *storage) GetLastServerUpdatedAt(ctx context.Context) (time.Time, error) {
	const op = "sqlite: get last server updated at"

	var item time.Time
	err := s.db.QueryRowContext(ctx, `SELECT MAX(server_updated_at) FROM vaults`).Scan(&item)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return item, serr.ErrRecordNotFound
		}
	}
	return item, e.Wrap(op, err)
}
