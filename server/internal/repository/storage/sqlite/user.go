package sqlite

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/server/internal/model/user"
	serr "github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

func (s *storage) AddUser(ctx context.Context, u user.User) error {
	const op = "sqlite: add user"

	_, err := s.db.ExecContext(ctx,
		`INSERT 
		INTO users(email, is_email_verified, auth_key, created_at) 
		VALUES (?, ?, ?, ?)`,
		u.Email, u.IsEmailVerified, u.AuthKey, u.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), duplicateKeyErrorCode) {
			return e.Wrap(op, serr.ErrRecordAlreadyExists)
		}
		return e.Wrap(op, err)
	}

	return nil
}

func (s *storage) GetUser(ctx context.Context, email string) (user.User, error) {
	const op = "sqlite: get user"

	u := user.User{
		Email: email,
	}
	err := s.db.QueryRowContext(ctx,
		`SELECT is_email_verified, auth_key, created_at 
		FROM users 
		WHERE email = ?`, email).
		Scan(&u.IsEmailVerified, &u.AuthKey, &u.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return user.User{}, e.Wrap(op, serr.ErrRecordNotFound)
		}
		return user.User{}, e.Wrap(op, err)
	}

	return u, nil
}

func (s *storage) UpdateUser(ctx context.Context, u user.User) error {
	const op = "sqlite: update user"

	res, err := s.db.ExecContext(ctx,
		`UPDATE users 
		SET is_email_verified = ?, auth_key = ?, created_at = ? 
		WHERE email = ?`,
		u.IsEmailVerified, u.AuthKey, u.CreatedAt, u.Email)
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
