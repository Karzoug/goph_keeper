package service

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"golang.org/x/exp/slog"

	"github.com/Karzoug/goph_keeper/pkg/e"
	"github.com/Karzoug/goph_keeper/server/internal/model/auth/token"
	"github.com/Karzoug/goph_keeper/server/internal/model/user"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/server/internal/service/task"
)

const emailSendingTimeout = 3 * time.Second

// Register registers a new user.
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

	code, err := generateNumericCode(s.cfg.Email.CodeLength)
	if err != nil {
		return e.Wrap(op, err)
	}

	err = s.caches.mail.Set(ctx, u.Email, code, s.cfg.Email.CodeLifetime)
	if err != nil {
		return e.Wrap(op, err)
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

	tsk, err := task.NewWelcomeVerificationEmailTask(u.Email, code)
	if err != nil {
		return e.Wrap(op, err)
	}
	err = s.rtaskClient.Enqueue(tsk, emailSendingTimeout)
	if err != nil {
		return e.Wrap(op, err)
	}

	s.logger.Debug("user successfully added to storage", slog.String("email", u.Email))

	return nil
}

// Login logs in a user.
func (s *Service) Login(ctx context.Context, email string, hash []byte) (string, error) {
	const op = "service: login user"

	u, err := s.getUser(ctx, email, hash)
	if err != nil {
		return "", e.Wrap(op, err)
	}

	if !u.IsEmailVerified {
		return "", ErrUserEmailNotVerified
	}

	tokenString, err := s.setUserToAuthCache(ctx, u)
	if err != nil {
		return "", e.Wrap(op, err)
	}

	return tokenString, nil
}

// LoginWithEmailCode logs in a user if user needs verification.
func (s *Service) LoginWithEmailCode(ctx context.Context, email string, hash []byte, code string) (string, error) {
	const op = "service: login user with email code"

	u, err := s.getUser(ctx, email, hash)
	if err != nil {
		return "", e.Wrap(op, err)
	}

	ccode, err := s.caches.mail.Get(ctx, email)
	if err != nil {
		return "", e.Wrap(op, err)
	}

	if ccode != code {
		return "", ErrUserEmailNotVerified
	}

	u.IsEmailVerified = true

	err = s.storage.UpdateUser(ctx, u)
	if err != nil {
		return "", e.Wrap(op, err)
	}

	_ = s.caches.mail.Delete(ctx, email)

	tokenString, err := s.setUserToAuthCache(ctx, u)
	if err != nil {
		return "", e.Wrap(op, err)
	}
	return tokenString, nil
}

// AuthUser verifies user's token and returns the email if success.
func (s *Service) AuthUser(ctx context.Context, tokenString string) (string, error) {
	const op = "service: auth user"

	token, err := token.FromString(tokenString, s.cfg.Token.SecretKey)
	if err != nil {
		return "", e.Wrap(op, ErrUserInvalidToken)
	}
	if token.IsExpired() {
		return "", e.Wrap(op, ErrUserNeedAuthentication)
	}

	email, err := s.caches.auth.Get(ctx, token.ID())
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return "", e.Wrap(op, ErrUserNeedAuthentication)
		}
		return "", e.Wrap(op, err)
	}

	return email, nil
}

// getUser returns user by email and auth hash.
func (s *Service) getUser(ctx context.Context, email string, authHash []byte) (user.User, error) {
	const op = "get user"

	u, err := s.storage.GetUser(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return user.User{}, ErrUserNotExists
		}
		return user.User{}, e.Wrap(op, err)
	}
	if !u.AuthKey.Verify(authHash) {
		return user.User{}, ErrUserInvalidHash
	}
	return u, nil
}

// setUserToAuthCache adds user email to auth cache and returns token.
func (s *Service) setUserToAuthCache(ctx context.Context, u user.User) (string, error) {
	eTime := time.Now().Add(s.cfg.Token.TokenLifetime)

	t := token.New(eTime, s.cfg.Token.SecretKey)
	err := s.caches.auth.Set(ctx, t.ID(), u.Email, time.Until(eTime))
	if err != nil {
		return "", err
	}

	return t.String(), nil
}

func generateNumericCode(n int) (string, error) {
	const op = "generate numeric code"

	rnd := make([]byte, n)
	_, err := rand.Read(rnd)
	if err != nil {
		return "", e.Wrap(op, err)
	}
	for i := range rnd {
		rnd[i] = '0' + rnd[i]%10
	}
	return string(rnd), nil
}
