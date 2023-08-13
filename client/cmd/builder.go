package main

import (
	"log/slog"

	"github.com/ilyakaznacheev/cleanenv"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/Karzoug/goph_keeper/client/internal/config"
)

var (
	logFilename    = "log.log"
	configFilename = "config.yaml"
)

func buildConfig() (*config.Config, error) {
	cfg := new(config.Config)

	err := cleanenv.ReadConfig(configFilename, cfg)
	if err != nil {
		return nil, err
	}

	cfg.Env = envMode
	cfg.Version = buildVersion

	return cfg, nil
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
