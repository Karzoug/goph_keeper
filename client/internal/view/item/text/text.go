package text

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
	value vault.Text

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
		AddText("Save text:", true, tview.AlignLeft, tcell.ColorWhite)

	v.Frame = frame

	return v
}

func (v *View) Init() (common.KeyHandlerFnc, common.Help) {
	form := tview.NewForm()
	form.SetBorderPadding(1, 1, 0, 1)
	form.AddInputField("Name", v.item.Name, 60, nil, func(name string) {
		v.item.Name = name
	})
	form.AddTextArea("Text", v.value.Text, 60, 10, 0, func(text string) {
		v.value.Text = text
	})
	form.AddButton("Save", func() {
		form.ClearButtons()
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
	txt, ok := value.(vault.Text)
	if !ok {
		return item.ErrWrongItemType
	}
	v.value = txt

	return nil
}

func (v *View) saveCmd() {
	err := item.Set(v.client, v.item, v.value)
	if err != nil {
		v.msgCh <- common.NewErrMsg(err)
		v.appUpdateFn(func() {
			v.form.AddButton("Save", func() {
				v.form.ClearButtons()
				go v.saveCmd()
			})
		})
		return
	}

	// clear before go to list items
	v.value = vault.Text{}
	v.appUpdateFn(func() {
		v.Frame.SetPrimitive(nil)
		v.form = nil
	})

	v.msgCh <- common.NewMsg("Text saved!")
	v.msgCh <- common.ToViewMsg{
		ViewType: common.ListItems,
	}
}

func (v *View) keyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		v.value = vault.Text{}
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
