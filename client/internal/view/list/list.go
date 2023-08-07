package list

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type View struct {
	client  *client.Client
	idNames []vault.IDName
	table   table.Model
}

func New(c *client.Client) View {
	columns := []table.Column{
		{Title: "Name", Width: 70},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
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

	return View{
		client: c,
		table:  t,
	}
}

func (v *View) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlL:
			return vc.ToViewCmd(vc.Login)
		case tea.KeyCtrlN:
			return vc.ToViewCmd(vc.ChooseItemType)
		case tea.KeyEnter:
			return tea.Sequence(vc.ToViewCmd(vc.Item),
				item.GetCmd(v.client, v.idNames[v.table.Cursor()].ID))
		default:
		}

	case successfulListMsg:
		v.idNames = msg
		rows := make([]table.Row, len(msg))
		for i, idName := range msg {
			rows[i] = make(table.Row, 1)
			rows[i][0] = idName.Name
		}
		v.table.SetRows(rows)

	case successfulSyncMsg:
		return tea.Batch(ListIDNameCmd(v.client), func() tea.Msg {
			return vc.MsgMsg{
				Msg:  "Your vault has been updated.",
				Time: time.Now(),
			}
		})
	}
	v.table, cmd = v.table.Update(msg)
	return cmd
}

func (v View) View(body *strings.Builder, help *strings.Builder) {
	body.WriteString("\n\nYour vault:\n")
	body.WriteString(baseStyle.Render(v.table.View()))

	help.WriteString("ctrl+n create • tab next • shift+tab prev • ")
}
