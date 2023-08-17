package config

import (
	"github.com/Karzoug/goph_keeper/server/internal/config/grpc"
	"github.com/Karzoug/goph_keeper/server/internal/config/rtask"
	"github.com/Karzoug/goph_keeper/server/internal/config/service"
	"github.com/Karzoug/goph_keeper/server/internal/config/smtp"
)

// Config is a configuration for GophKeeper server.
type Config struct {
	// Env is a environment type (production or development).
	Env EnvType `env:"ENV" envDefault:"production"`
	// GRPC is a configuration for gRPC server.
	GRPC grpc.Config `envPrefix:"GRPC_"`
	// Service is a configuration for service layer.
	Service service.Config `envPrefix:"SERVICE_"`
	// RTask is a configuration for redis task manager.
	RTask rtask.Config `envPrefix:"RTASK_"`
	// SMTP is a configuration for SMTP server.
	SMTP smtp.Config `envPrefix:"SMTP_"`
}
