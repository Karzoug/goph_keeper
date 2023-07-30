package view

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type emailVerificationView struct {
	textInput textinput.Model
}

func initialEmailVerificationView() emailVerificationView {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 6
	ti.Width = 10

	return emailVerificationView{
		textInput: ti,
	}
}

func initEmailVerificationView() tea.Cmd {
	return textinput.Blink
}

func updateEmailVerificationView(v *view, msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return v.verifyEmail
		}

	case errorAndSwitchMsg:
		return tea.Batch(toLoginView, func() tea.Msg {
			return errMsg(msg)
		})

	case successfulVerificationEmailMsg:
		return toListItemsView
	}

	v.subviews.emailVerification.textInput, cmd = v.subviews.emailVerification.textInput.Update(msg)
	return cmd
}

func viewEmailVerificationView(v view, b *strings.Builder) {
	fmt.Fprintf(b, "Enter the code from mail:\n\n%s\n\n", v.subviews.emailVerification.textInput.View())
}
