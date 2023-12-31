package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	serr "github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

func (s *storage) ListVaultItems(ctx context.Context) ([]vault.Item, error) {
	const op = "sqlite: list vault items"

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, type, value, client_updated_at, server_updated_at, is_deleted FROM vaults`)

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res := make([]vault.Item, 0)
	for rows.Next() {
		var item vault.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Value, &item.ClientUpdatedAt, &item.ServerUpdatedAt, &item.IsDeleted)
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
func (s *storage) ListVaultItemsIDName(ctx context.Context) ([]vault.IDName, error) {
	const op = "sqlite: list vault items names"

	rows, err := s.db.QueryContext(ctx, `SELECT id, name FROM vaults WHERE is_deleted = 0;`)

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res := make([]vault.IDName, 0)
	for rows.Next() {
		var item vault.IDName
		err := rows.Scan(&item.ID, &item.Name)
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
		`SELECT id, name, type, value, client_updated_at, server_updated_at, is_deleted 
		FROM vaults 
		WHERE server_updated_at < client_updated_at;`)

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res := make([]vault.Item, 0)
	for rows.Next() {
		var item vault.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Value, &item.ClientUpdatedAt, &item.ServerUpdatedAt, &item.IsDeleted)
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

func (s *storage) GetVaultItem(ctx context.Context, id string) (vault.Item, error) {
	const op = "sqlite: get vault item"

	item := vault.Item{ID: id}
	err := s.db.QueryRowContext(ctx,
		`SELECT name, type, value, client_updated_at, server_updated_at, is_deleted 
		FROM vaults 
		WHERE id = ?`, id).
		Scan(&item.Name, &item.Type, &item.Value, &item.ClientUpdatedAt, &item.ServerUpdatedAt, &item.IsDeleted)
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
		`INSERT INTO vaults(id,name,type,value,client_updated_at,server_updated_at,is_deleted) VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) 
		DO UPDATE SET name = excluded.name, type = excluded.type, value=excluded.value, 
		client_updated_at=excluded.client_updated_at, server_updated_at=excluded.server_updated_at, is_deleted=excluded.is_deleted;`,
		item.ID, item.Name, item.Type, item.Value, item.ClientUpdatedAt, item.ServerUpdatedAt, item.IsDeleted)
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
func (s *storage) DeleteVaultItem(ctx context.Context, id string) error {
	const op = "sqlite: delete vault item"

	res, err := s.db.ExecContext(ctx, `DELETE FROM vaults WHERE id = ?;`, id)
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
func (s *storage) MoveVaultItemToConflict(ctx context.Context, id string) error {
	const op = "sqlite: move vault item to conflict"

	tx, err := s.db.Begin()
	if err != nil {
		return e.Wrap(op, err)
	}
	defer tx.Rollback() //nolint:errcheck

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO conflict_vaults (id,name,type,value,client_updated_at,server_updated_at,is_deleted)
		SELECT id,name,type,value,client_updated_at,server_updated_at,is_deleted
		FROM vaults WHERE id = ?;
		DELETE FROM vaults WHERE id = ?;`, id, id)

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
func (s *storage) GetLastServerUpdatedAt(ctx context.Context) (int64, error) {
	const op = "sqlite: get last server updated at"

	var t int64
	err := s.db.QueryRowContext(ctx, `SELECT server_updated_at FROM vaults ORDER BY server_updated_at DESC LIMIT 1`).Scan(&t)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, serr.ErrRecordNotFound
		}
	}
	return t, e.Wrap(op, err)
}
