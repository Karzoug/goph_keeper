package view

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle.Copy()
	noStyle      = lipgloss.NewStyle()
	helpStyle    = blurredStyle.Copy()

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type loginView struct {
	focusIndex int
	inputs     []textinput.Model
}

func initialLoginView() loginView {
	m := loginView{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Email"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
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

func viewLoginView(v view, b *strings.Builder) {
	b.WriteString("Enter your email and password to login:\n")

	for i := range v.subviews.login.inputs {
		b.WriteString(v.subviews.login.inputs[i].View())
		if i < len(v.subviews.login.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if v.subviews.login.focusIndex == len(v.subviews.login.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(b, "\n\n%s\n\n", *button)

	b.WriteString(helpStyle.Render("Press ctrl+r to go to registration tab"))
}

func initLoginView() tea.Cmd {
	return textinput.Blink
}

func updateLoginView(v *view, msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+r":
			return toRegisterView

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			if s == "enter" && v.subviews.login.focusIndex == len(v.subviews.login.inputs) {
				return v.login
			}

			if s == "up" || s == "shift+tab" {
				v.subviews.login.focusIndex--
			} else {
				v.subviews.login.focusIndex++
			}

			if v.subviews.login.focusIndex > len(v.subviews.login.inputs) {
				v.subviews.login.focusIndex = 0
			} else if v.subviews.login.focusIndex < 0 {
				v.subviews.login.focusIndex = len(v.subviews.login.inputs)
			}

			cmds := make([]tea.Cmd, len(v.subviews.login.inputs))
			for i := 0; i <= len(v.subviews.login.inputs)-1; i++ {
				if i == v.subviews.login.focusIndex {
					// Set focused state
					cmds[i] = v.subviews.login.inputs[i].Focus()
					v.subviews.login.inputs[i].PromptStyle = focusedStyle
					v.subviews.login.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				v.subviews.login.inputs[i].Blur()
				v.subviews.login.inputs[i].PromptStyle = noStyle
				v.subviews.login.inputs[i].TextStyle = noStyle
			}

			return tea.Batch(cmds...)
		}
	case successfulLoginMsg:
		return tea.Batch(toListItemsView, func() tea.Msg {
			return msgMsg{
				msg:  "You are logged in!",
				time: time.Now(),
			}
		})
	case loginNeedEmailVerificationMsg:
		return toEmailVerificationView
	}

	// Handle character input and blinking
	cmd := v.updateLoginViewInputs(msg)

	return cmd
}

func (v *view) updateLoginViewInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(v.subviews.login.inputs))

	for i := range v.subviews.login.inputs {
		v.subviews.login.inputs[i], cmds[i] = v.subviews.login.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}
