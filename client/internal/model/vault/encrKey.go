package vault

import (
	"errors"

	"github.com/matthewhartstonge/argon2"

	"github.com/Karzoug/goph_keeper/pkg/e"
)

var (
	ErrEmptyEmail    = errors.New("empty email")
	ErrEmptyPassword = errors.New("empty password")
)

type EncryptionKey []byte

func NewEncryptionKey(email, password []byte) (EncryptionKey, error) {
	if len(password) == 0 {
		return nil, ErrEmptyPassword
	}
	if len(email) == 0 {
		return nil, ErrEmptyEmail
	}
	argon := argon2.DefaultConfig()

	encoded, err := argon.Hash(password, email)
	if err != nil {
		return nil, e.Wrap("model: create encryption key", err)
	}

	return EncryptionKey(encoded.Encode()), nil
}
