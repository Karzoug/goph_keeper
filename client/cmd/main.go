package main

import (
	"log"
	"os"

	"log/slog"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/config"
	"github.com/Karzoug/goph_keeper/client/internal/view"
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

	c, err := client.New(cfg, logger)
	if err != nil {
		logger.Error("client create", sl.Error(err))
		os.Exit(1)
	}
	defer c.Close()

	logger.Debug(
		"starting goph-keeper client",
		slog.String("env", envMode.String()),
		slog.String("build version", buildVersion),
		slog.String("build date", buildDate),
	)

	v, err := view.New(c)
	if err != nil {
		logger.Error("build ui", sl.Error(err))
		os.Exit(1)
	}
	if err := v.Run(); err != nil {
		logger.Error("application stopped with error", sl.Error(err))
		os.Exit(1)
	}
}
