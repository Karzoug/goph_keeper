package service

import (
	"time"

	"github.com/Karzoug/goph_keeper/server/internal/model/auth/token"
)

type Config struct {
	Token struct {
		// TokenLifetime is the lifetime of the token.
		TokenLifetime time.Duration `env:"TOKEN_LIFETIME,notEmpty" envDefault:"168h"`
		// SecretKey is the secret key to sign token.
		SecretKey token.SecretKey `env:"TOKEN_SECRET_KEY,notEmpty,unset"`
	}
}
