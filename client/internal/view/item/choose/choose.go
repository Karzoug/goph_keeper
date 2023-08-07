package choose

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
)

type SuccessfulMsg cvault.ItemType

type View struct {
	choices []string
	cursor  int
}

func New() View {
	return View{
		choices: []string{"Password", "Card", "Text", "Binary"},
		cursor:  0,
	}
}

func (v *View) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			return vc.ToViewCmd(vc.ListItems)
		case tea.KeyUp, tea.KeyShiftTab:
			if v.cursor > 0 {
				v.cursor--
			}
		case tea.KeyDown, tea.KeyTab:
			if v.cursor < len(v.choices)-1 {
				v.cursor++
			}
		case tea.KeyEnter, tea.KeySpace:
			return tea.Sequence(vc.ToViewCmd(vc.Item),
				func() tea.Msg {
					return SuccessfulMsg(v.cursor + 1) // +1 because unknown type missing
				})
		default:
		}
	}
	return nil
}

func (v View) View(body *strings.Builder, help *strings.Builder) {
	body.WriteString("\n\nChoose a type of item\n\n")
	for i, choice := range v.choices {
		if i == v.cursor {
			body.WriteString("> ")
		} else {
			body.WriteString("  ")
		}
		body.WriteString(choice)
		body.WriteByte('\n')
	}

	help.WriteString("tab next • shift+tab prev • esc back • ")
}
