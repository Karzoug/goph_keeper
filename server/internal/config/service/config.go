package service

import (
	"time"

	"github.com/Karzoug/goph_keeper/server/internal/config/storage"
	"github.com/Karzoug/goph_keeper/server/internal/model/auth/token"
)

type Config struct {
	Token struct {
		// TokenLifetime is the lifetime of the token.
		TokenLifetime time.Duration `env:"TOKEN_LIFETIME,notEmpty" envDefault:"168h"`
		// SecretKey is the secret key to sign token.
		SecretKey token.SecretKey `env:"TOKEN_SECRET_KEY,notEmpty,unset"`
	}
	Email struct {
		CodeLength   int           `env:"EMAIL_CODE_LENGTH,notEmpty" envDefault:"6"`
		CodeLifetime time.Duration `env:"EMAIL_CODE_LIFETIME,notEmpty" envDefault:"24h"`
	}
	// Storage is a configuration for storage.
	Storage storage.Config `envPrefix:"STORAGE_"`
}
