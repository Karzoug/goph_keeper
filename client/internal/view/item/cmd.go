package item

import (
	"context"
	"errors"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
)

type (
	SuccessfulSetItemMsg struct{}
	SuccessfulGetItemMsg struct {
		Item           vault.Item
		DecryptedValue any
	}
	ConflictVersionSetItemMsg struct{}
)

func GetCmd(c *client.Client, id string) tea.Cmd {
	item, value, err := c.DecryptAndGetVaultItem(context.TODO(), id)
	if err != nil {
		return tea.Batch(vc.ShowErrCmd(err.Error()), vc.ToViewCmd(vc.ListItems))
	}
	return func() tea.Msg {
		return SuccessfulGetItemMsg{
			Item:           item,
			DecryptedValue: value,
		}
	}
}

func SetCmd(c *client.Client, item vault.Item, value any) tea.Cmd {
	return func() tea.Msg {
		err := c.EncryptAndSetVaultItem(context.TODO(), item, value)
		if err != nil {
			switch {
			case errors.Is(err, client.ErrConflictVersion):
				return ConflictVersionSetItemMsg{}
			case errors.Is(err, client.ErrAppInternal) || errors.Is(err, client.ErrUserNeedAuthentication):
				return vc.ErrMsg{
					Time: time.Now(),
					Err:  err.Error(),
				}
			}
		}
		return SuccessfulSetItemMsg{}
	}
}
