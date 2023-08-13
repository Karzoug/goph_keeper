package model

import (
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
)

type Credentials struct {
	Email   string
	Token   string
	EncrKey vault.EncryptionKey
}
