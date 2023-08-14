package password

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
)

type View struct {
	Frame *tview.Frame
	form  *tview.Form

	item  vault.Item
	value vault.Password

	client      *client.Client
	msgCh       chan<- any
	appUpdateFn func(func()) *tview.Application
}

func New(c *client.Client, msgCh chan<- any, appUpdateFn func(func()) *tview.Application) View {
	v := View{
		client:      c,
		msgCh:       msgCh,
		appUpdateFn: appUpdateFn,
	}

	frame := tview.NewFrame(nil).
		AddText("Save password:", true, tview.AlignLeft, tcell.ColorWhite)

	v.Frame = frame

	return v
}

func (v *View) Init() (common.KeyHandlerFnc, common.Help) {
	form := tview.NewForm()
	form.SetBorderPadding(1, 1, 0, 1)
	form.AddInputField("Name", v.item.Name, 40, nil, func(name string) {
		v.item.Name = name
	})
	form.AddInputField("Login", v.value.Login, 40, nil, func(login string) {
		v.value.Login = login
	})
	form.AddInputField("Password", v.value.Password, 40, nil, func(password string) {
		v.value.Password = password
	})
	form.AddButton("Save", func() {
		go v.saveCmd()
	})
	v.form = form
	v.Frame.SetPrimitive(form)

	return v.keyHandler, "tab next • esc back • "
}

func (v *View) Update(vitem vault.Item, value any) error {
	v.item = vitem

	if value == nil {
		return nil
	}
	psw, ok := value.(vault.Password)
	if !ok {
		return common.NewErrMsg(item.ErrWrongItemType)
	}
	v.value = psw

	return nil
}

func (v *View) saveCmd() {
	err := item.Set(v.client, v.item, v.value)
	if err != nil {
		v.msgCh <- common.NewErrMsg(err)
		return
	}

	// clear before go to list items
	v.value = vault.Password{}
	v.appUpdateFn(func() {
		v.Frame.SetPrimitive(nil)
		v.form = nil
	})

	v.msgCh <- common.NewMsg("Password saved!")
	v.msgCh <- common.ToViewMsg{
		ViewType: common.ListItems,
	}
}

func (v *View) keyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		v.value = vault.Password{}
		v.Frame.SetPrimitive(nil)
		v.form = nil
		go func() {
			v.msgCh <- common.ToViewMsg{
				ViewType: common.ListItems,
			}
		}()
	}

	return event
}
