package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"log/slog"

	"golang.org/x/sync/errgroup"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	c, err := client.New(ctx, cfg, logger)
	if err != nil {
		logger.Error("client create", sl.Error(err))
		os.Exit(1)
	}

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

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return c.Run(ctx)
	})
	eg.Go(func() error {
		return v.Run(ctx)
	})

	if err := eg.Wait(); err != nil {
		logger.Error("application stopped with error", sl.Error(err))
		os.Exit(1)
	}
}
