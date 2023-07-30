package config

// Config is a configuration for goph-keeper client.
type Config struct {
	// Env is a environment type (production or development).
	Env     EnvType
	Version string
	Address string `env:"ADDRESS,notEmpty" envDefault:":8080"`
}
