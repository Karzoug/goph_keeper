package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Karzoug/goph_keeper/common/model/vault"
	"github.com/Karzoug/goph_keeper/pkg/e"
	serr "github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

func (s *storage) SetVaultItem(ctx context.Context, email string, item vault.Item) error {
	const op = "postgres: set vault item"

	res, err := s.db.Exec(ctx,
		`INSERT INTO vaults(id,email,name,type,value,updated_at,is_deleted) VALUES($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(id,email) 
		DO UPDATE SET name = excluded.name, type = excluded.type, value=excluded.value, updated_at=excluded.updated_at, is_deleted=excluded.is_deleted
		WHERE vaults.updated_at=$8;`,
		item.ID, email, item.Name, item.Type, item.Value, item.ClientUpdatedAt, item.IsDeleted, item.ServerUpdatedAt)
	if err != nil {
		return e.Wrap(op, err)
	}

	if res.RowsAffected() == 0 {
		return serr.ErrNoRecordsAffected
	}

	return nil
}
func (s *storage) LastUpdateVault(ctx context.Context, email string) (time.Time, error) {
	const op = "postgres: last update vault"

	var t time.Time
	err := s.db.QueryRow(ctx,
		`SELECT updated_at FROM vaults WHERE email = $1 ORDER BY updated_at DESC LIMIT 1;`, email).
		Scan(&t)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return t, e.Wrap(op, serr.ErrRecordNotFound)
		}
		return t, e.Wrap(op, err)
	}

	return t, nil
}

func (s *storage) DeleteVaultItem(ctx context.Context, email string, id string) error {
	const op = "postgres: delete vault item"

	res, err := s.db.Exec(ctx,
		`DELETE FROM vaults WHERE id = $1 AND email = $2;`, id, email)
	if err != nil {
		return e.Wrap(op, err)
	}

	if res.RowsAffected() == 0 { // driver specific
		return e.Wrap(op, serr.ErrNoRecordsAffected)
	}

	return nil
}
func (s *storage) ListVaultItems(ctx context.Context, email string, since int64) ([]vault.Item, error) {
	const op = "postgres: list vault items"

	var (
		rows pgx.Rows
		err  error
	)
	if since != 0 {
		rows, err = s.db.Query(ctx,
			`SELECT id, name, type, value, updated_at, is_deleted FROM vaults WHERE email = $1 AND updated_at > $2;`, email, since)
	} else {
		rows, err = s.db.Query(ctx,
			`SELECT id, name, type, value, updated_at, is_deleted FROM vaults WHERE email = $1`, email)
	}

	if err != nil {
		return nil, e.Wrap(op, err)
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (vault.Item, error) {
		var item vault.Item
		err = rows.Scan(&item.ID, &item.Name, &item.Type, &item.Value, &item.ServerUpdatedAt, &item.IsDeleted)
		return item, err
	})
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	err = rows.Err()
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	return res, nil
}
