package main

import (
	"os"

	"log/slog"

	"github.com/caarlos0/env/v9"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/Karzoug/goph_keeper/client/internal/config"
)

var logFilename = "log.log"

func buildConfig() (*config.Config, error) {
	cfg := new(config.Config)

	opts := env.Options{
		Prefix: "GOPH_KEEPER_",
	}

	cfg.Env = envMode
	cfg.Version = buildVersion

	return cfg, env.ParseWithOptions(cfg, opts)
}

func buildLogger(env config.EnvType) (*slog.Logger, error) {
	var log *slog.Logger

	switch env {
	case config.EnvDevelopment:
		log = slog.New(
			slog.NewJSONHandler(&lumberjack.Logger{
				Filename:   logFilename,
				MaxSize:    5, // megabytes
				MaxBackups: 3,
				MaxAge:     28, //days
			}, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log, nil
}
