package view

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type listItemsView struct {
	table table.Model
}

func initialListItemsView() listItemsView {
	columns := []table.Column{
		{Title: "Name", Width: 80},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return listItemsView{
		table: t,
	}
}

func viewListItemsView(v view, b *strings.Builder) {
	fmt.Fprint(b, "Your vault:\n")
	b.WriteString(baseStyle.Render(v.subviews.listItems.table.View()) + "\n")
}

func updateListItems(v *view, msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if v.subviews.listItems.table.Focused() {
				v.subviews.listItems.table.Blur()
			} else {
				v.subviews.listItems.table.Focus()
			}
		case "ctrl+l":
			return toLoginView
		case "enter":
			return tea.Sequence(toItemView, v.getItem(v.subviews.listItems.table.SelectedRow()[0]))
		}

	case errorAndSwitchMsg:
		return tea.Batch(toLoginView, func() tea.Msg {
			return errMsg(msg)
		})

	case successfulListItemsNamesMsg:
		v.subviews.listItems.table.SetRows(msg)

	case successfulUpdateListItemsMsg:
		return tea.Batch(v.listItemsNames, func() tea.Msg {
			return msgMsg{
				msg:  "Your vault has been updated.",
				time: time.Now(),
			}
		})
	}
	v.subviews.listItems.table, cmd = v.subviews.listItems.table.Update(msg)
	return cmd
}
