package auth

import (
	"errors"

	"github.com/Karzoug/goph_keeper/pkg/e"

	"github.com/matthewhartstonge/argon2"
)

var ErrEmptyHash = errors.New("empty hash")

// Key is a user auth data (hash) to store on server,
// it's an Argon2 hash based on client auth hash and random salt.
type Key []byte

// NewKey returns a new auth key to store on server.
func NewKey(hash []byte) (Key, error) {
	const op = "model: create auth key"

	if len(hash) == 0 {
		return nil, ErrEmptyHash
	}
	argon := argon2.DefaultConfig()

	encoded, err := argon.HashEncoded(hash)
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	return Key(encoded), nil
}

// Verify returns true if hash matches the key and otherwise false.
func (k Key) Verify(hash []byte) bool {
	ok, err := argon2.VerifyEncoded(hash, k)
	if err != nil || !ok {
		return false
	}

	return true
}
