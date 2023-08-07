package password

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
)

type View struct {
	client     *client.Client
	item       vault.Item
	isNewItem  bool
	focusIndex int
	inputs     []textinput.Model
}

func New(c *client.Client, item vault.Item, p vault.Password, isNewItem bool) View {
	m := View{
		client:    c,
		inputs:    make([]textinput.Model, 3),
		item:      item,
		isNewItem: isNewItem,
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = vc.CursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			if !isNewItem {
				t.SetValue(item.Name)
			}
			t.Placeholder = "Name"
			t.Focus()
			t.PromptStyle = vc.FocusedStyle
			t.TextStyle = vc.FocusedStyle
			t.CharLimit = 128
		case 1:
			if !isNewItem {
				t.SetValue(p.Login)
			}
			t.Placeholder = "Login"
			t.PromptStyle = vc.NoStyle
			t.TextStyle = vc.NoStyle
			t.CharLimit = 128
		case 2:
			if !isNewItem {
				t.SetValue(p.Password)
			}
			t.Placeholder = "Password"
			t.PromptStyle = vc.NoStyle
			t.TextStyle = vc.NoStyle
			t.CharLimit = 128
		}

		m.inputs[i] = t
	}

	return m
}

func (v *View) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s := msg.Type; s { //nolint:exhaustive
		case tea.KeyEsc:
			for i := 0; i < len(v.inputs); i++ {
				v.inputs[i].Reset()
			}
			return vc.ToViewCmd(vc.ListItems)
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			// Did the user press enter while the submit button was focused?
			if s == tea.KeyEnter && v.focusIndex == len(v.inputs) {
				return v.cmd()
			}

			if s == tea.KeyUp || s == tea.KeyShiftTab {
				v.focusIndex--
			} else {
				v.focusIndex++
			}

			if v.focusIndex > len(v.inputs) {
				v.focusIndex = 0
			} else if v.focusIndex < 0 {
				v.focusIndex = len(v.inputs)
			}

			cmds := make([]tea.Cmd, len(v.inputs))
			for i := 0; i <= len(v.inputs)-1; i++ {
				if i == v.focusIndex {
					// Set focused state
					cmds[i] = v.inputs[i].Focus()
					v.inputs[i].PromptStyle = vc.FocusedStyle
					v.inputs[i].TextStyle = vc.FocusedStyle
					continue
				}
				// Remove focused state
				v.inputs[i].Blur()
				v.inputs[i].PromptStyle = vc.NoStyle
				v.inputs[i].TextStyle = vc.NoStyle
			}

			return tea.Batch(cmds...)
		}
	case item.SuccessfulSetItemMsg:
		return tea.Batch(vc.ToViewCmd(vc.ListItems),
			vc.ShowMsgCmd("Saved!"))
	case item.ConflictVersionSetItemMsg:
		return tea.Batch(vc.ToViewCmd(vc.ListItems),
			vc.ShowMsgCmd("Saved!"),
			vc.ShowErrCmd(client.ErrConflictVersion.Error()))
	}

	// Handle character input and blinking
	cmd := v.updatePasswordViewInputs(msg)

	return cmd
}

func (v *View) updatePasswordViewInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(v.inputs))

	for i := range v.inputs {
		v.inputs[i], cmds[i] = v.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (v View) View(body *strings.Builder, help *strings.Builder) {
	if v.isNewItem {
		body.WriteString("\n\nAdd new password:\n")
	} else {
		body.WriteString("\n\nEdit password:\n")
	}

	for i := range v.inputs {
		body.WriteString(v.inputs[i].View())
		if i < len(v.inputs)-1 {
			body.WriteRune('\n')
		}
	}

	button := &vc.BlurredButton
	if v.focusIndex == len(v.inputs) {
		button = &vc.FocusedButton
	}
	fmt.Fprintf(body, "\n\n%s", *button)

	help.WriteString("tab next • shift+tab prev • esc back • ")
}

func (v View) cmd() tea.Cmd {
	v.item.Name = v.inputs[0].Value()
	v.item.ClientUpdatedAt = time.Now().Unix()

	return item.SetCmd(v.client,
		v.item, vault.Password{
			Login:    v.inputs[1].Value(),
			Password: v.inputs[2].Value(),
		})
}
