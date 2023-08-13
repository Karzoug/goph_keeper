package config

import "github.com/Karzoug/goph_keeper/client/internal/repository/storage"

// Config is a configuration for goph-keeper client.
type Config struct {
	// Env is a environment type (production or development).
	Env                    EnvType
	CredentialsStorageType storage.Type `yaml:"credentials_storage_type" env:"GOPH_KEEPER_CREDENTIALS_STORAGE_TYPE" env-default:"database"`
	Version                string
	Host                   string `yaml:"host" env:"GOPH_KEEPER_HOST" env-default:"localhost"`
	Port                   string `yaml:"port" env:"GOPH_KEEPER_PORT" env-default:"8080"`
	CertFilename           string `yaml:"cert_filename" env:"GOPH_KEEPER_CERT_FILENAME"`
}
