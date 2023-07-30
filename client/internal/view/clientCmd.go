package view

import (
	"context"
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
)

type (
	successfulListItemsNamesMsg    []table.Row
	successfulUpdateListItemsMsg   struct{}
	successfulVerificationEmailMsg struct{}
	successfulLoginMsg             struct{}
	successfulRegistrationMsg      struct{}
	errorAndSwitchMsg              errMsg
	loginNeedEmailVerificationMsg  struct{}
)

func (v view) register() tea.Msg {
	err := v.client.Register(context.TODO(),
		v.subviews.register.inputs[0].Value(),
		[]byte(v.subviews.register.inputs[1].Value()))
	if err != nil {
		return errMsg{
			time: time.Now(),
			err:  err.Error(),
		}
	}

	return successfulRegistrationMsg{}
}

func (v view) login() tea.Msg {
	err := v.client.Login(context.TODO(),
		v.subviews.login.inputs[0].Value(),
		[]byte(v.subviews.login.inputs[1].Value()))
	if err != nil {
		if errors.Is(err, client.ErrUserEmailNotVerified) {
			return loginNeedEmailVerificationMsg{}
		}

		return errMsg{
			time: time.Now(),
			err:  err.Error(),
		}
	}

	return successfulLoginMsg{}
}

func (v view) verifyEmail() tea.Msg {
	err := v.client.VerifyEmail(context.TODO(), v.subviews.emailVerification.textInput.Value())
	if err != nil {
		if errors.Is(err, client.ErrInvalidEmailVerificationCode) {
			return errMsg{
				time: time.Now(),
				err:  client.ErrInvalidEmailVerificationCode.Error(),
			}
		}

		return errorAndSwitchMsg{
			time: time.Now(),
			err:  err.Error(),
		}
	}

	return successfulVerificationEmailMsg{}
}

func (v view) listItemsNames() tea.Msg {
	names, err := v.client.ListVaultItemsNames(context.TODO())
	if err != nil {
		return errMsg{
			time: time.Now(),
			err:  err.Error(),
		}
	}

	res := make([]table.Row, len(names))
	for i, name := range names {
		res[i] = make(table.Row, 1)
		res[i][0] = name
	}

	return successfulListItemsNamesMsg(res)
}

func (v view) getItem(name string) tea.Cmd {
	return func() tea.Msg {
		// TODO: implement
		return nil
	}
}

func (v view) updateListItems() tea.Msg {
	err := v.client.UpdateVaultItems(context.TODO())
	if err != nil {
		return errMsg{
			time: time.Now(),
			err:  err.Error(),
		}
	}

	return successfulUpdateListItemsMsg{}
}
