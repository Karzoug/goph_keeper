package auth

import (
	"errors"

	"github.com/matthewhartstonge/argon2"

	"github.com/Karzoug/goph_keeper/pkg/e"
)

var (
	ErrEmptyEmail    = errors.New("empty email")
	ErrEmptyPassword = errors.New("empty password")
)

type Hash []byte

func NewHash(email, password []byte) (Hash, error) {
	if len(password) == 0 {
		return nil, ErrEmptyPassword
	}
	if len(email) == 0 {
		return nil, ErrEmptyEmail
	}
	argon := argon2.DefaultConfig()
	argon.TimeCost++ // Hash differs from EncryptionKey with an additional encryption step

	encoded, err := argon.Hash(password, email)
	if err != nil {
		return nil, e.Wrap("model: create hash", err)
	}

	return Hash(encoded.Encode()), nil
}
