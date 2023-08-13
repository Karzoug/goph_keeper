package config

// Config is a configuration for goph-keeper client.
type Config struct {
	// Env is a environment type (production or development).
	Env          EnvType
	Version      string
	Host         string `yaml:"host" env:"GOPH_KEEPER_HOST" env-default:"localhost"`
	Port         string `yaml:"port" env:"GOPH_KEEPER_PORT" env-default:"8080"`
	CertFilename string `yaml:"cert_filename" env:"GOPH_KEEPER_CERT_FILENAME"`
}
