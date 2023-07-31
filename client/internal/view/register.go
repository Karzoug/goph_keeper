package view

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
)

type registerView struct {
	focusIndex int
	inputs     []textinput.Model
}

func initialRegisterView() registerView {
	m := registerView{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 64

		switch i {
		case 0:
			t.Placeholder = "Email"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = fmt.Sprintf("Password (min %d chars)", client.MinPasswordLength)
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		}

		m.inputs[i] = t
	}

	return m
}

func viewRegisterView(v registerView, b *strings.Builder) {
	b.WriteString("Enter your email and password to register:\n")

	for i := range v.inputs {
		b.WriteString(v.inputs[i].View())
		if i < len(v.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if v.focusIndex == len(v.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("Press ctrl+l to go to login tab"))
}

func initRegisterView() tea.Cmd {
	return textinput.Blink
}

func updateRegisterView(v *view, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+l":
			return toLoginView

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			if s == "enter" && v.subviews.register.focusIndex == len(v.subviews.register.inputs) {
				return v.register
			}

			if s == "up" || s == "shift+tab" {
				v.subviews.register.focusIndex--
			} else {
				v.subviews.register.focusIndex++
			}

			if v.subviews.register.focusIndex > len(v.subviews.register.inputs) {
				v.subviews.register.focusIndex = 0
			} else if v.subviews.register.focusIndex < 0 {
				v.subviews.register.focusIndex = len(v.subviews.register.inputs)
			}

			cmds := make([]tea.Cmd, len(v.subviews.register.inputs))
			for i := 0; i <= len(v.subviews.register.inputs)-1; i++ {
				if i == v.subviews.register.focusIndex {
					// Set focused state
					cmds[i] = v.subviews.register.inputs[i].Focus()
					v.subviews.register.inputs[i].PromptStyle = focusedStyle
					v.subviews.register.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				v.subviews.register.inputs[i].Blur()
				v.subviews.register.inputs[i].PromptStyle = noStyle
				v.subviews.register.inputs[i].TextStyle = noStyle
			}

			return tea.Batch(cmds...)
		}
	case successfulRegistrationMsg:
		return tea.Batch(toLoginView, func() tea.Msg {
			return msgMsg{
				msg:  "You are registered!",
				time: time.Now(),
			}
		})
	}

	// Handle character input and blinking
	cmd := v.updateRegisterViewInputs(msg)

	return cmd
}

func (v *view) updateRegisterViewInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(v.subviews.register.inputs))

	for i := range v.subviews.register.inputs {
		v.subviews.register.inputs[i], cmds[i] = v.subviews.register.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
