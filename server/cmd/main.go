package main

import (
	"log"

	"golang.org/x/exp/slog"

	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
	"github.com/Karzoug/goph_keeper/server/internal/service"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

func main() {
	cfg, err := buildConfig()
	if err != nil {
		log.Fatal("parse config error: ", err)
	}

	logger := buildLogger(cfg.Env)

	logger.Info(
		"starting goph-keeper server",
		slog.String("env", cfg.Env.String()),
		slog.String("build version", buildVersion),
		slog.String("build date", buildDate),
	)

	_, err = service.New(cfg.Service, logger)
	if err != nil {
		logger.Error("main", sl.Error(err))
	}
}
