package view

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type itemView struct {
}

func initialItemView() itemView {
	return itemView{}
}

func viewItemView(v view, b *strings.Builder) {
	// TODO: implement
}

func updateItem(v *view, msg tea.Msg) tea.Cmd {
	// TODO: implement
	return func() tea.Msg {
		return nil
	}
}
