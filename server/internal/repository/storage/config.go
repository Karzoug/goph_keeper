package storage

type Config struct {
	Type Type   `env:"TYPE,notEmpty"`
	URI  string `env:"URI,notEmpty"`
}
