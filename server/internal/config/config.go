package config

import "github.com/Karzoug/goph_keeper/server/internal/service"

// Config is a configuration for GophKeeper server.
type Config struct {
	// Env is a environment type (production or development).
	Env EnvType `env:"ENV" envDefault:"production"`
	// Service is a configuration for service.
	Service service.Config `envPrefix:"SERVICE_"`
}
