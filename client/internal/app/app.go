package app

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/config"
	"github.com/Karzoug/goph_keeper/client/internal/view"
	"github.com/Karzoug/goph_keeper/pkg/e"
)

func Run(cfg *config.Config, logger *slog.Logger) error {
	const op = "app run"

	c, err := client.New(cfg, logger)
	if err != nil {
		return e.Wrap(op, err)
	}
	defer c.Close()

	p := tea.NewProgram(view.New(c))
	if _, err := p.Run(); err != nil {
		return e.Wrap(op, err)
	}

	return nil
}
