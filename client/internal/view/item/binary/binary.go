package binary

import (
	"errors"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
	"github.com/Karzoug/goph_keeper/client/pkg/filepicker"
)

type View struct {
	Frame      *tview.Frame
	form       *tview.Form
	filepicker *filepicker.Window

	item  vault.Item
	value vault.Binary

	path     string
	filename string

	client *client.Client
	msgCh  chan<- any
	app    *tview.Application
}

func New(c *client.Client, msgCh chan<- any, app *tview.Application) (View, error) {
	v := View{
		client: c,
		msgCh:  msgCh,
		app:    app,
	}

	filepicker.SharedConfig.Application = v.app
	path := c.RootPath()
	if path == "" {
		path, _ = os.UserHomeDir()
	}
	if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
		path, err = os.UserHomeDir()
		if err != nil {
			return v, err
		}
	}
	filepicker.SharedConfig.RootPath = path

	frame := tview.NewFrame(nil).
		AddText("Save binary:", true, tview.AlignLeft, tcell.ColorWhite)

	v.Frame = frame

	return v, nil
}

func (v *View) Init() (common.KeyHandlerFnc, common.Help) {
	nameInput := tview.NewInputField().SetLabel("Name: ").SetFieldWidth(40)
	nameInput.SetBorderPadding(0, 1, 0, 0)

	filepicker := filepicker.NewWindow(60, 15)

	f := tview.NewForm()

	if v.item.ID != "" {
		modal := tview.NewModal().
			SetText("Are you sure?").
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonIndex == 0 {
					go v.delete()
				}
				v.Frame.SetPrimitive(f)
			})
		f.AddTextView("Name", v.item.Name, 40, 1, false, false).
			AddInputField("Filename", v.filename, 40, nil, func(filename string) {
				v.filename = filename
			}).
			AddFormItem(filepicker).
			AddButton("Decrypt and save", func() {
				v.path = filepicker.GetCurrentPath()
				go v.save()
			}).
			AddButton("Delete", func() {
				v.Frame.SetPrimitive(modal)
			})
	} else {
		f.AddInputField("Name", v.item.Name, 40, nil, func(name string) {
			v.item.Name = name
		}).
			AddFormItem(filepicker).
			AddButton("Save", func() {
				v.path = filepicker.GetCurrentPath()
				go v.save()
			})
	}

	f.SetBorderPadding(1, 1, 0, 1)
	v.Frame.SetPrimitive(f)

	return v.keyHandler, "tab next • esc back • "
}

func (v *View) Update(vitem vault.Item, value any) error {
	v.item = vitem

	if value == nil {
		return nil
	}
	b, ok := value.(vault.Binary)
	if !ok {
		return item.ErrWrongItemType
	}
	v.value = b

	return nil
}

func (v *View) save() {
	var err error
	if v.item.ID == "" {
		err = v.createCmd(v.path)
	} else {
		err = v.saveOnDiskCmd(v.path, v.filename)
	}
	if err != nil {
		v.msgCh <- common.NewErrMsg(err)
		if errors.Is(err, client.ErrAppInternal) {
			return
		}
	}

	// clear before go to list items
	v.value = vault.Binary{}
	v.app.QueueUpdateDraw(func() {
		v.Frame.SetPrimitive(nil)
		v.form = nil
	})

	v.msgCh <- common.NewMsg("Binary saved!")
	v.msgCh <- common.ToViewMsg{
		ViewType: common.ListItems,
	}
}

func (v *View) delete() {
	if err := item.Delete(v.client, v.item.ID); err != nil {
		v.msgCh <- common.NewErrMsg(err)
		if errors.Is(err, client.ErrAppInternal) {
			return
		}
	}

	// clear before go to list items
	v.value = vault.Binary{}
	v.app.QueueUpdateDraw(func() {
		v.Frame.SetPrimitive(nil)
		v.form = nil
	})

	v.msgCh <- common.NewMsg("Item deleted!")
	v.msgCh <- common.ToViewMsg{
		ViewType: common.ListItems,
	}
}

func (v *View) keyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		v.value = vault.Binary{}
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
