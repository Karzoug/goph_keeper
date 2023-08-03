package sqlite

import (
	"database/sql"
	"errors"

	serr "github.com/Karzoug/goph_keeper/client/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

const (
	emailDBKey   = "CREDS_EMAIL"
	tokenDBKey   = "CREDS_TOKEN"
	encrKeyDBKey = "CREDS_ENCRKEY"
)

func (s *storage) SetCredentials(email, token, encrKey string) error {
	const op = "sqlite: set credentials"

	tx, err := s.db.Begin()
	if err != nil {
		return e.Wrap(op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO app(key,value) VALUES(?, ?)
	ON CONFLICT(key) 
	DO UPDATE SET value=excluded.value;`)
	if err != nil {
		return e.Wrap(op, err)
	}
	defer stmt.Close()

	if _, err = stmt.Exec(emailDBKey, email); err != nil {
		return err
	}
	if _, err = stmt.Exec(encrKeyDBKey, encrKey); err != nil {
		return err
	}
	if len(token) == 0 {
		return e.Wrap(op, tx.Commit())
	}
	if _, err = stmt.Exec(tokenDBKey, token); err != nil {
		return err
	}
	return e.Wrap(op, tx.Commit())
}

func (s *storage) GetCredentials() (email, token, encrKey string, err error) {
	const op = "sqlite: get credentials"

	tx, err := s.db.Begin()
	if err != nil {
		return "", "", "", e.Wrap(op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`SELECT value FROM app WHERE key = ?;`)
	if err != nil {
		return "", "", "", e.Wrap(op, err)
	}
	defer stmt.Close()

	if err := stmt.QueryRow(emailDBKey).Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", e.Wrap(op, serr.ErrRecordNotFound)
		}
		return "", "", "", e.Wrap(op, err)
	}
	if err := stmt.QueryRow(tokenDBKey).Scan(&token); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", "", e.Wrap(op, err)
		}
	}
	if err := stmt.QueryRow(encrKeyDBKey).Scan(&encrKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", "", e.Wrap(op, serr.ErrRecordNotFound)
		}
		return "", "", "", e.Wrap(op, err)
	}

	err = e.Wrap(op, tx.Commit())
	return
}
func (s *storage) DeleteCredentials() error {
	const op = "sqlite: delete credentials"

	tx, err := s.db.Begin()
	if err != nil {
		return e.Wrap(op, err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`DELETE FROM app WHERE key = ?;`)
	if err != nil {
		return e.Wrap(op, err)
	}
	defer stmt.Close()

	if _, err = stmt.Exec(encrKeyDBKey); err != nil {
		return err
	}
	if _, err = stmt.Exec(tokenDBKey); err != nil {
		return err
	}
	return e.Wrap(op, tx.Commit())
}
