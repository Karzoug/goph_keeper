package service

import (
	"context"
	"errors"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/server/internal/model/user"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage"
)

func (s *Service) Register(ctx context.Context, email string, hash []byte) error {
	const op = "service: register user"

	u, err := user.New(email, hash)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmailFormat):
			return ErrInvalidEmailFormat
		case errors.Is(err, user.ErrInvalidHashFormat):
			return ErrInvalidHashFormat
		default:
			return e.Wrap(op, err)
		}
	}

	err = s.storage.AddUser(ctx, u)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrRecordAlreadyExists):
			return ErrUserAlreadyExists
		default:
			return e.Wrap(op, err)
		}
	}

	return nil
}

func (s *Service) Login(ctx context.Context, email string, hash []byte) (string, error) {
	const op = "service: login user"

	u, err := s.getUser(ctx, email, hash)
	if err != nil {
		return "", e.Wrap(op, err)
	}

	if !u.IsEmailVerified {
		return "", ErrUserEmailNotVerified
	}

	// TODO: generate token
	tokenString := ""

	return tokenString, nil
}

func (s *Service) AuthUser(ctx context.Context, tokenString string) (string, error) {
	const op = "service: auth user"

	panic("not implemented")
}

func (s *Service) getUser(ctx context.Context, email string, hash []byte) (user.User, error) {
	const op = "get user"

	u, err := s.storage.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return user.User{}, ErrUserNotExists
		}
		return user.User{}, e.Wrap(op, err)
	}
	if !u.AuthKey.Verify(hash) {
		return user.User{}, ErrUserInvalidHash
	}
	return u, nil
}
