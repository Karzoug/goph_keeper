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

func (k *EncryptionKey) UnmarshalBinary(data []byte) error {
	const op = "encryption key from bytes"

	raw, err := argon2.Decode(data)
	if err != nil {
		return e.Wrap(op, err)
	}
	k.Raw = raw
	return nil
}

func (k EncryptionKey) MarshalBinary() (data []byte, err error) {
	const op = "encryption key to bytes"

	return k.Raw.Encode(), nil
}
