package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/server/internal/model/user"
	serr "github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

func (s *storage) AddUser(ctx context.Context, u user.User) error {
	const op = "postgres: add user"

	_, err := s.db.Exec(ctx,
		`INSERT 
		INTO users(email, is_email_verified, auth_key, created_at) 
		VALUES ($1, $2, $3, $4)`,
		u.Email, u.IsEmailVerified, []byte(u.AuthKey), u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == duplicateKeyErrorCode {
			return e.Wrap(op, serr.ErrRecordAlreadyExists)
		}
		return e.Wrap(op, err)
	}

	return nil
}

func (s *storage) GetUser(ctx context.Context, email string) (user.User, error) {
	const op = "postgres: get user"

	u := user.User{
		Email: email,
	}
	var byteKey []byte
	err := s.db.QueryRow(ctx,
		`SELECT is_email_verified, auth_key, created_at 
		FROM users 
		WHERE email = $1`, email).
		Scan(&u.IsEmailVerified, &byteKey, &u.CreatedAt)

	u.AuthKey = byteKey

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, e.Wrap(op, serr.ErrRecordNotFound)
		}
		return user.User{}, e.Wrap(op, err)
	}

	return u, nil
}

func (s *storage) UpdateUser(ctx context.Context, u user.User) error {
	const op = "postgres: update user"

	tag, err := s.db.Exec(ctx,
		`UPDATE users 
		SET is_email_verified = $1, auth_key = $2, created_at = $3 
		WHERE email = $4`,
		u.IsEmailVerified, []byte(u.AuthKey), u.CreatedAt, u.Email)
	if err != nil {
		return e.Wrap(op, err)
	}

	if tag.RowsAffected() == 0 { // driver specific
		return e.Wrap(op, serr.ErrNoRecordsAffected)
	}

	return nil
}
