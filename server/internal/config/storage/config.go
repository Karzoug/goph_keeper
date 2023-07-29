package storage

type Config struct {
	// URI is a database identifier.
	// URI consists of a scheme, an authority, a path, a query string, and a fragment
	URI string `env:"URI"`
}
