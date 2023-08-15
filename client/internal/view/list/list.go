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

const syncCmdTimeout = 5 * time.Second

type View struct {
	Frame *tview.Frame
	list  *tview.List

	client      *client.Client
	msgCh       chan<- any
	appUpdateFn func(func()) *tview.Application

	idNames []vault.IDName
}

func New(c *client.Client, msgCh chan<- any, appUpdateFn func(func()) *tview.Application) View {
	frame := tview.NewFrame(nil).
		AddText("Your vault:", true, tview.AlignLeft, tcell.ColorWhite)
	v := View{
		client:      c,
		msgCh:       msgCh,
		Frame:       frame,
		appUpdateFn: appUpdateFn,
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

	return v.keyHandler, "ctrl+n create • tab next • ctrl+u sync • "
}

func (v *View) Update() error {
	ctx, cancel := context.WithTimeout(context.TODO(), common.StandartTimeout)
	defer cancel()

	in, err := v.client.ListVaultItemsIDName(ctx)
	if err != nil {
		return err
	}
	v.idNames = in
	return nil
}

func (v *View) sync() {
	ctx, cancel := context.WithTimeout(context.TODO(), syncCmdTimeout)
	defer cancel()

	err := v.client.SyncVaultItems(ctx)
	if err != nil {
		if errors.Is(err, client.ErrUserNeedAuthentication) {
			return
		}
		v.msgCh <- common.NewErrMsg(err)
		return
	}

	if err := v.Update(); err != nil {
		v.msgCh <- common.NewErrMsg(err)
		return
	}

	v.appUpdateFn(func() {
		v.Init()
	})

	v.msgCh <- common.NewMsg("Vault synced!")
}

func (v *View) keyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() { // nolint:exhaustive
	case tcell.KeyCtrlN:
		go func() {
			v.msgCh <- common.ToViewMsg{
				ViewType: common.ChooseItemType,
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
	case tcell.KeyCtrlU:
		go v.sync()
	}
	return event
}
