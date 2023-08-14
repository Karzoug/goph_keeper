package list

import (
	"context"
	"errors"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/view/common"
)

const syncCmdTimeout = 10 * time.Second

type View struct {
	Frame *tview.Frame
	list  *tview.List

	client *client.Client
	msgCh  chan<- any

	idNames []vault.IDName
}

func New(c *client.Client, msgCh chan<- any) View {
	frame := tview.NewFrame(nil).
		AddText("Your vault:", true, tview.AlignLeft, tcell.ColorWhite)
	v := View{
		client: c,
		msgCh:  msgCh,
		Frame:  frame,
	}
	return v
}

func (v *View) Init() (common.KeyHandlerFnc, common.Help) {
	list := tview.NewList().ShowSecondaryText(false)

	for r := 0; r < len(v.idNames); r++ {
		list = list.AddItem(v.idNames[r].Name, "", 0, nil)
	}

	v.list = list
	v.Frame.SetPrimitive(list)

	return v.keyHandler, "ctrl+n create • tab next • "
}

func (v *View) Update() error {
	ctx, cancel := context.WithTimeout(context.TODO(), common.StandartTimeout)
	defer cancel()

	in, err := v.client.ListVaultItemsIDName(ctx)
	if err != nil {
		return common.NewErrMsg(err)
	}
	v.idNames = in
	return nil
}

func (v *View) Sync(c *client.Client) error {
	ctx, cancel := context.WithTimeout(context.TODO(), syncCmdTimeout)
	defer cancel()

	err := c.SyncVaultItems(ctx)
	if err != nil {
		if errors.Is(err, client.ErrUserNeedAuthentication) {
			return nil
		}
		return common.NewErrMsg(err)
	}
	return nil
}

func (v *View) keyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlN:
		go func() {
			v.msgCh <- common.ToViewMsg{
				ViewType: common.ChooseItemType,
			}
		}()
	case tcell.KeyCtrlL:
		go func() {
			v.msgCh <- common.ToViewMsg{
				ViewType: common.Auth,
			}
		}()
	case tcell.KeyEnter:
		curr := v.list.GetCurrentItem()
		go func() {
			v.msgCh <- common.ToViewMsg{
				ViewType: common.Item,
				Value:    v.idNames[curr].ID,
			}
		}()
	}
	return event
}
