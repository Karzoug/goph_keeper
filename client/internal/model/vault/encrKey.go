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

type EncryptionKey struct {
	argon2.Raw
}

func NewEncryptionKey(email, password []byte) (EncryptionKey, error) {
	const op = "create encryption key"

	if len(password) == 0 {
		return EncryptionKey{}, e.Wrap(op, ErrEmptyPassword)
	}
	if len(email) == 0 {
		return EncryptionKey{}, e.Wrap(op, ErrEmptyEmail)
	}
	argon := argon2.DefaultConfig()

	encoded, err := argon.Hash(password, email)
	if err != nil {
		return EncryptionKey{}, e.Wrap(op, err)
	}

	return EncryptionKey{Raw: encoded}, nil
}

func EncryptionKeyFromString(s string) (EncryptionKey, error) {
	const op = "encryption key from string"

	raw, err := argon2.Decode([]byte(s))
	if err != nil {
		return EncryptionKey{}, e.Wrap(op, err)
	}
	return EncryptionKey{Raw: raw}, nil
}
