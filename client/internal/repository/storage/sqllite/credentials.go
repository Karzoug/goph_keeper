package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Karzoug/goph_keeper/client/internal/model"
	serr "github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

const (
	emailDBKey   = "CREDS_EMAIL"
	tokenDBKey   = "CREDS_TOKEN"
	encrKeyDBKey = "CREDS_ENCRKEY"
)

func (s *storage) SetCredentials(ctx context.Context, creds model.Credentials) error {
	const op = "sqlite: set credentials"

	encrKeyBytes := creds.EncrKey.Encode()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return e.Wrap(op, err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO app(key,value) VALUES(?, ?)
	ON CONFLICT(key) 
	DO UPDATE SET value=excluded.value;`)
	if err != nil {
		return e.Wrap(op, err)
	}
	defer stmt.Close()

	if _, err = stmt.ExecContext(ctx, emailDBKey, creds.Email); err != nil {
		return err
	}
	if _, err = stmt.ExecContext(ctx, encrKeyDBKey, encrKeyBytes); err != nil {
		return err
	}
	if len(creds.Token) == 0 {
		return e.Wrap(op, tx.Commit())
	}
	if _, err = stmt.ExecContext(ctx, tokenDBKey, creds.Token); err != nil {
		return err
	}
	return e.Wrap(op, tx.Commit())
}

func (s *storage) GetCredentials(ctx context.Context) (model.Credentials, error) {
	const op = "sqlite: get credentials"

	creds := model.Credentials{}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return creds, e.Wrap(op, err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx, `SELECT value FROM app WHERE key = ?;`)
	if err != nil {
		return creds, e.Wrap(op, err)
	}
	defer stmt.Close()

	if err := stmt.QueryRowContext(ctx, emailDBKey).Scan(&creds.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return creds, e.Wrap(op, serr.ErrRecordNotFound)
		}
		return creds, e.Wrap(op, err)
	}
	if err := stmt.QueryRowContext(ctx, tokenDBKey).Scan(&creds.Token); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return creds, e.Wrap(op, err)
		}
	}
	var encrKeyBytes []byte
	if err := stmt.QueryRowContext(ctx, encrKeyDBKey).Scan(&encrKeyBytes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return creds, e.Wrap(op, serr.ErrRecordNotFound)
		}
		return creds, e.Wrap(op, err)
	}
	if err := creds.EncrKey.UnmarshalBinary(encrKeyBytes); err != nil {
		return creds, e.Wrap(op, err)
	}

	return creds, e.Wrap(op, tx.Commit())
}
func (s *storage) DeleteCredentials(ctx context.Context) error {
	const op = "sqlite: delete credentials"

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return e.Wrap(op, err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx, `DELETE FROM app WHERE key = ?;`)
	if err != nil {
		return e.Wrap(op, err)
	}
	defer stmt.Close()

	if _, err = stmt.ExecContext(ctx, encrKeyDBKey); err != nil {
		return err
	}
	if _, err = stmt.ExecContext(ctx, tokenDBKey); err != nil {
		return err
	}
	return e.Wrap(op, tx.Commit())
}
