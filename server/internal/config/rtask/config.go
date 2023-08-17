package rtask

import "github.com/Karzoug/goph_keeper/server/internal/config/storage"

type Config struct {
	// Maximum number of concurrent processing of tasks.
	//
	// If set to a zero or negative value, will be overwrited by the value
	// to the number of CPUs usable by the current process.
	Concurrency int `env:"CONCURRENCY" envDefault:"0"`
	// Storage is a configuration for storage (redis).
	Storage storage.Config `envPrefix:"STORAGE_"`
}
