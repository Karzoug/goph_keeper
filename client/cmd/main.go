package main

import (
	"log"
	"os"

	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

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
		logger.Error("application cannot start", sl.Error(err))
		os.Exit(1)
	}
	defer c.Close()

	logger.Debug(
		"starting goph-keeper client",
		slog.String("env", envMode.String()),
		slog.String("build version", buildVersion),
		slog.String("build date", buildDate),
	)

	// os signals registered in bubble tea program
	p := tea.NewProgram(view.New(c))
	if _, err := p.Run(); err != nil {
		logger.Debug("application stopped with error", sl.Error(err))
		os.Exit(1)
	}
}
