package main

import (
	"os"
	"reflect"

	"log/slog"

	"github.com/caarlos0/env/v9"

	"github.com/Karzoug/goph_keeper/pkg/logger/slog/pretty"
	"github.com/Karzoug/goph_keeper/server/internal/config"
)

func buildConfig() (*config.Config, error) {
	cfg := new(config.Config)

	opts := env.Options{
		Prefix: "GOPHKEEPER_",
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(cfg.Env): config.EnvTypeParserFunc},
	}

	return cfg, env.ParseWithOptions(cfg, opts)
}

func buildLogger(env config.EnvType) *slog.Logger {
	var log *slog.Logger

	switch env {
	case config.EnvDevelopment:
		opts := pretty.HandlerOptions{
			SlogOpts: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}

		handler := opts.NewPrettyHandler(os.Stdout)
		return slog.New(handler)
	case config.EnvProduction:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
