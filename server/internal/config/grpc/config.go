package grpc

type Config struct {
	Host         string `env:"HOST"`
	Port         string `env:"PORT,notEmpty" envDefault:"8080"`
	CertFileName string `env:"CERT_FILE_NAME"`
	KeyFileName  string `env:"KEY_FILE_NAME"`
}

func (cfg Config) Address() string {
	return cfg.Host + ":" + cfg.Port
}
