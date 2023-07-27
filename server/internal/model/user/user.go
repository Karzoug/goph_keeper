package user

import (
	"errors"
	"net/mail"
	"time"

	"github.com/Karzoug/goph_keeper/server/internal/model/auth"
)

var (
	ErrInvalidHashFormat  = errors.New("invalid hash format")
	ErrInvalidEmailFormat = errors.New("invalid email format")
)

// User is a service user (client).
type User struct {
	Email           string
	IsEmailVerified bool
	AuthKey         auth.Key
	CreatedAt       time.Time
}

// New returns a new user.
func New(email string, authHash []byte) (User, error) {
	if len(email) == 0 {
		return User{}, ErrInvalidEmailFormat
	}
	e, err := mail.ParseAddress(email)
	if err != nil {
		return User{}, ErrInvalidEmailFormat
	}
	email = e.Address

	authKey, err := auth.NewKey(authHash)
	if err != nil {
		return User{}, ErrInvalidHashFormat
	}

	return User{
		Email:           email,
		IsEmailVerified: false,
		AuthKey:         authKey,
		CreatedAt:       time.Now(),
	}, nil
}
