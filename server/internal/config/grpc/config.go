package grpc

type Config struct {
	Address string `env:"ADDRESS,notEmpty" envDefault:":8080"`
}
