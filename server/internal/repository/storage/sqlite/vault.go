package sqlite

import (
	"context"
	"database/sql"

	"github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/e"
	serr "github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

func (s *storage) SetVaultItem(ctx context.Context, email string, item vault.Item) error {
	const op = "sqlite: set vault item"

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO vaults(id,email,name,type,value,updated_at,is_deleted) VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id,email) 
		DO UPDATE SET name = excluded.name, type = excluded.type, value=excluded.value, updated_at=excluded.updated_at, is_deleted=excluded.is_deleted
		WHERE vaults.updated_at=?;`,
		item.ID, email, item.Name, item.Type, item.Value, item.ClientUpdatedAt, item.IsDeleted, item.ServerUpdatedAt)
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

func (s *storage) ListVaultItems(ctx context.Context, email string, since int64) ([]vault.Item, error) {
	const op = "sqlite: list vault items"

	var (
		rows *sql.Rows
		err  error
	)
	if since != 0 {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, name, type, value, updated_at, is_deleted FROM vaults WHERE email = ? AND updated_at > ?;`, email, since)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, name, type, value, updated_at, is_deleted FROM vaults WHERE email = ?`, email)
	}

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res := make([]vault.Item, 0)
	for rows.Next() {
		var item vault.Item
		err := rows.Scan(&item.ID, &item.Name, &item.Type, &item.Value, &item.ServerUpdatedAt, &item.IsDeleted)
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
