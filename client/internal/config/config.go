package config

// Config is a configuration for goph-keeper client.
type Config struct {
	// Env is a environment type (production or development).
	Env          EnvType
	Version      string
	Host         string `env:"HOST" envDefault:"localhost"`
	Port         string `env:"PORT,notEmpty" envDefault:"8080"`
	CertFilename string `env:"CERT_FILENAME"`
}
