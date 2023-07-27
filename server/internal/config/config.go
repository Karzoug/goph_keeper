package config

import (
	"github.com/Karzoug/goph_keeper/server/internal/delivery/grpc"
	"github.com/Karzoug/goph_keeper/server/internal/repository/storage"
	"github.com/Karzoug/goph_keeper/server/internal/service"
)

// Config is a configuration for GophKeeper server.
type Config struct {
	// Env is a environment type (production or development).
	Env EnvType `env:"ENV" envDefault:"production"`
	// GRPC is a configuration for gRPC.
	GRPC grpc.Config `envPrefix:"GRPC_"`
	// Service is a configuration for service.
	Service service.Config `envPrefix:"SERVICE_"`
	// Storage is a configuration for storage.
	Storage storage.Config `envPrefix:"STORAGE_"`
}
