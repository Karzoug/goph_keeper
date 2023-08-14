package choose

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/common/model/vault"
)

type View struct {
	Frame *tview.Frame
	list  *tview.List

	client *client.Client
	msgCh  chan<- any

	choices []string
}

func New(c *client.Client, msgCh chan<- any) View {
	v := View{
		client:  c,
		msgCh:   msgCh,
		choices: []string{"Password", "Card", "Text", "Binary"},
	}

	list := tview.NewList().ShowSecondaryText(false)
	for i := 0; i < len(v.choices); i++ {
		list = list.AddItem(v.choices[i], "", 0, nil)
	}

	frame := tview.NewFrame(list).
		AddText("Choose a type of item:", true, tview.AlignLeft, tcell.ColorWhite)
	v.list = list
	v.Frame = frame

	return v
}

func (v *View) Init() (common.KeyHandlerFnc, common.Help) {
	return v.keyHandler, "tab next • esc back • "
}

func (v *View) keyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		go func() {
			v.msgCh <- common.ToViewMsg{
				ViewType: common.ListItems,
			}
		}()
	case tcell.KeyEnter:
		var value vault.ItemType
		switch v.list.GetCurrentItem() {
		case 0:
			value = vault.Password
		case 1:
			value = vault.Card
		case 2:
			value = vault.Text
		case 3:
			value = vault.Binary
		default:
			return event
		}
		go func() {
			v.msgCh <- common.ToViewMsg{
				ViewType: common.Item,
				Value:    value,
			}
		}()
	}
	return event
}
