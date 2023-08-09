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

const syncCmdTimeout = 10 * time.Second

type (
	successfulListMsg []vault.IDName
	successfulSyncMsg struct{}
)

func ListIDNameCmd(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.TODO(), vc.StandartTimeout)
		defer cancel()

		in, err := c.ListVaultItemsIDName(ctx)
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
		ctx, cancel := context.WithTimeout(context.TODO(), syncCmdTimeout)
		defer cancel()

		err := c.SyncVaultItems(ctx)
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
