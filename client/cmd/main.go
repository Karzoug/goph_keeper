package main

import (
	"log"
	"os"

	"log/slog"

	"github.com/Karzoug/goph_keeper/client/internal/app"
	"github.com/Karzoug/goph_keeper/client/internal/config"
	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"

	envMode = config.EnvProduction
)

func main() {
	cfg, err := buildConfig()
	if err != nil {
		log.Fatal("parse config error:\n", err)
	}

	logger, err := buildLogger(envMode)
	if err != nil {
		log.Fatal("build logger error:\n", err)
	}

	logger.Debug(
		"starting goph-keeper client",
		slog.String("env", envMode.String()),
		slog.String("build version", buildVersion),
		slog.String("build date", buildDate),
	)

	// os signals registered in bubble tea program
	if err := app.Run(cfg, logger); err != nil {
		logger.Debug("application stopped with error", sl.Error(err))
		os.Exit(1)
	}
}
