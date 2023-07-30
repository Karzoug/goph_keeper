package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/exp/slog"

	"github.com/Karzoug/goph_keeper/pkg/logger/slog/sl"
	"github.com/Karzoug/goph_keeper/server/internal/app"
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
	slog.SetDefault(logger)

	logger.Info(
		"starting goph-keeper server",
		slog.String("env", cfg.Env.String()),
		slog.String("build version", buildVersion),
		slog.String("build date", buildDate),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		logger.Error("application stopped with error", sl.Error(err))
		os.Exit(1)
	}
}
