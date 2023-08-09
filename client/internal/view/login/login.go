package login

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
)

type View struct {
	client     *client.Client
	focusIndex int
	inputs     []textinput.Model
}

type (
	successfulMsg            struct{}
	needEmailVerificationMsg struct{}
)

func New(c *client.Client) View {
	m := View{
		client: c,
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = vc.CursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Email"
			t.Focus()
			t.PromptStyle = vc.FocusedStyle
			t.TextStyle = vc.FocusedStyle
			t.CharLimit = 64
		case 1:
			t.Placeholder = "Password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func (v *View) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s := msg.Type; s {
		case tea.KeyCtrlR:
			return vc.ToViewCmd(vc.Register)

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			// Did the user press enter while the submit button was focused?
			if s == tea.KeyEnter && v.focusIndex == len(v.inputs) {
				return v.cmd
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
		default:
		}
	case successfulMsg:
		return tea.Batch(vc.ToViewCmd(vc.ListItems), vc.ShowMsgCmd("You are entered!"))
	case needEmailVerificationMsg:
		return vc.ToViewCmd(vc.EmailVerification)
	}

	// Handle character input and blinking
	cmd := v.updateLoginViewInputs(msg)

	return cmd
}

func (v *View) updateLoginViewInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(v.inputs))

	for i := range v.inputs {
		v.inputs[i], cmds[i] = v.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (v View) View(body *strings.Builder, help *strings.Builder) {
	body.WriteString("\n\nEnter your email and password to login:\n")

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

	help.WriteString("ctrl+r register • ")
}

func (v View) cmd() tea.Msg {
	email := v.inputs[0].Value()
	password := []byte(v.inputs[1].Value())

	ctx, cancel := context.WithTimeout(context.TODO(), vc.StandartTimeout)
	defer cancel()

	err := v.client.Login(ctx, email, password)
	if err != nil {
		if errors.Is(err, client.ErrUserEmailNotVerified) {
			return needEmailVerificationMsg{}
		}
		return vc.ErrMsg{
			Time: time.Now(),
			Err:  err.Error(),
		}
	}
	return successfulMsg{}
}
