package storage

import (
	"errors"

	"github.com/99designs/keyring"
)

var ErrUnknownCredentialsStorageType = errors.New("unknown credentials storage type")

type Type string

const (
	Database      Type = Type("database")
	Keychain      Type = Type(keyring.KeychainBackend)
	SecretService Type = Type(keyring.SecretServiceBackend)
	WinCred       Type = Type(keyring.WinCredBackend)
	Pass          Type = Type(keyring.PassBackend)
)

func (t *Type) UnmarshalText(text []byte) error {
	switch tt := Type(text); tt {
	case Database, Keychain, SecretService, WinCred, Pass:
		*t = tt
		return nil
	default:
		return ErrUnknownCredentialsStorageType
	}
}
