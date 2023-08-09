package email

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
	client    *client.Client
	textInput textinput.Model
}

type successfulMsg struct{}

func New(c *client.Client) View {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 6
	ti.Width = 10

	return View{
		client:    c,
		textInput: ti,
	}
}

func (v *View) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape:
			return vc.ToViewCmd(vc.Login)
		case tea.KeyEnter:
			return v.cmd
		default:
		}

	case successfulMsg:
		return vc.ToViewCmd(vc.ListItems)
	}

	v.textInput, cmd = v.textInput.Update(msg)
	return cmd
}

func (v View) View(body *strings.Builder, help *strings.Builder) {
	fmt.Fprintf(body, "\n\nEnter the code from mail:\n\n%s", v.textInput.View())

	help.WriteString("esc back • ")
}

func (v View) cmd() tea.Msg {
	ctx, cancel := context.WithTimeout(context.TODO(), vc.StandartTimeout)
	defer cancel()

	err := v.client.VerifyEmail(ctx, v.textInput.Value())
	if err != nil {
		if errors.Is(err, client.ErrInvalidEmailVerificationCode) {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  client.ErrInvalidEmailVerificationCode.Error(),
			}
		}
		return vc.ErrMsg{
			Time: time.Now(),
			Err:  err.Error(),
		}
	}
	return successfulMsg{}
}
