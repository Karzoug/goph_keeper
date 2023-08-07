package list

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
	successfulListMsg []vault.IDName
	successfulSyncMsg struct{}
)

func ListIDNameCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		in, err := c.ListVaultItemsIDName(context.TODO())
		if err != nil {
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  err.Error(),
			}
		}
		return successfulListMsg(in)
	}
}

func SyncCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		err := c.SyncVaultItems(context.TODO())
		if err != nil {
			if errors.Is(err, client.ErrUserNeedAuthentication) {
				return nil
			}
			return vc.ErrMsg{
				Time: time.Now(),
				Err:  err.Error(),
			}
		}
		return successfulSyncMsg{}
	}
}
