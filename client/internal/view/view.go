package view

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	"github.com/Karzoug/goph_keeper/client/internal/view/auth"
	"github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/email"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/binary"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/card"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/choose"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/password"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/text"
	"github.com/Karzoug/goph_keeper/client/internal/view/list"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
)

const (
	refreshNotificationInterval = 500 * time.Millisecond
	notificationLifetime        = 5 * time.Second
)

type View struct {
	app         *tview.Application
	root        *tview.Flex
	pages       *tview.Pages
	currentPage common.ViewType
	msgCh       chan any
	closeCh     chan struct{}

	client *client.Client

	subviews struct {
		auth     auth.View
		list     list.View
		email    email.View
		choose   choose.View
		password password.View
		text     text.View
		card     card.View
		binary   binary.View
	}
	footer struct {
		msgText    *tview.TextView
		msg        common.Msg
		errText    *tview.TextView
		err        common.ErrMsg
		statusText *tview.TextView
		helpText   *tview.TextView
	}
}

func New(client *client.Client) (*View, error) {
	var app = tview.NewApplication()
	var pages = tview.NewPages()

	v := &View{
		msgCh:   make(chan any),
		closeCh: make(chan struct{}),
		client:  client,
	}

	// create footer to view app info
	v.footer.msgText = tview.NewTextView().SetTextColor(tcell.ColorBlue)
	v.footer.errText = tview.NewTextView().SetTextColor(tcell.ColorRed)
	v.footer.statusText = tview.NewTextView().SetTextColor(tcell.ColorRed)
	v.footer.helpText = tview.NewTextView().SetTextColor(tcell.ColorGray)
	v.footer.helpText.SetBorderPadding(1, 0, 0, 0)

	// create subviews
	v.subviews.auth = auth.New(client, v.msgCh, app.QueueUpdateDraw)
	v.subviews.list = list.New(client, v.msgCh, app.QueueUpdateDraw)
	v.subviews.email = email.New(client, v.msgCh, app.QueueUpdateDraw)
	v.subviews.choose = choose.New(client, v.msgCh)
	v.subviews.password = password.New(client, v.msgCh, app.QueueUpdateDraw)
	v.subviews.text = text.New(client, v.msgCh, app.QueueUpdateDraw)
	v.subviews.card = card.New(client, v.msgCh, app.QueueUpdateDraw)
	var err error
	v.subviews.binary, err = binary.New(client, v.msgCh, app)
	if err != nil {
		return nil, err
	}

	// add subviews to pages
	pages.AddPage(common.Auth.String(), v.subviews.auth.Frame, true, false)
	pages.AddPage(common.ListItems.String(), v.subviews.list.Frame, true, false)
	pages.AddPage(common.EmailVerification.String(), v.subviews.email.Frame, true, false)
	pages.AddPage(common.ChooseItemType.String(), v.subviews.choose.Frame, true, false)
	pages.AddPage(common.Password.String(), v.subviews.password.Frame, true, false)
	pages.AddPage(common.Text.String(), v.subviews.text.Frame, true, false)
	pages.AddPage(common.Card.String(), v.subviews.card.Frame, true, false)
	pages.AddPage(common.Binary.String(), v.subviews.binary.Frame, true, false)
	pages.SetChangedFunc(v.initSubview)

	// create header to view app info
	header := fmt.Sprintf("Goph Keeper: your password manager & vault app\nversion: %s", client.Version())
	headerTextView := tview.NewTextView().SetText(header)

	// build main view
	var flex = tview.NewFlex()
	flex.SetDirection(tview.FlexRow).
		AddItem(headerTextView, 2, 0, false).
		AddItem(tview.NewFlex().
			AddItem(pages, 0, 1, false), 0, 1, true).
		AddItem(v.footer.msgText, 1, 0, false).
		AddItem(v.footer.errText, 1, 0, false).
		AddItem(v.footer.statusText, 1, 0, false).
		AddItem(v.footer.helpText, 2, 0, false)

	v.pages = pages
	v.root = flex
	v.app = app.SetRoot(flex, true).EnableMouse(true)

	return v, nil
}

func (v *View) handleMsgs() {
	for {
		select {
		case <-v.closeCh:
			return
		case msg := <-v.msgCh:
			switch msg := msg.(type) {
			case common.ErrMsg:
				v.footer.err = msg
				v.app.QueueUpdateDraw(func() {
					v.footer.errText.SetText("Error: " + msg.Error())
				})
			case common.Msg:
				v.footer.msg = msg
				v.app.QueueUpdateDraw(func() {
					v.footer.msgText.SetText(msg.String())
				})
			case common.ToViewMsg:
				v.currentPage = msg.ViewType

				switch msg.ViewType { //nolint:exhaustive
				case common.ListItems:
					if err := v.subviews.list.Update(); err != nil {
						err = common.NewErrMsg(err)
						v.app.QueueUpdateDraw(func() {
							v.footer.errText.SetText("Error: " + err.Error())
						})
						return
					}
				case common.Item:
					v.toItem(msg.Value)
				}

				v.pages.SwitchToPage(v.currentPage.String())
			}
		}
	}
}

func (v *View) Run() error {
	var err error
	// run main loop to handle tview events
	go func() {
		err = v.app.Run()
		close(v.closeCh)
	}()

	// run loop to handle msgs
	go v.handleMsgs()
	// run loop to update notifications in footer
	go v.updateNotifications()

	// go to start subview
	if v.client.HasLocalCredintials() {
		v.msgCh <- common.ToViewMsg{
			ViewType: common.ListItems,
		}
	} else {
		v.msgCh <- common.ToViewMsg{
			ViewType: common.Auth,
		}
	}

	<-v.closeCh
	return err
}

func (v *View) updateNotifications() {
	ticker := time.NewTicker(refreshNotificationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Since(v.footer.err.Time) > notificationLifetime {
				v.app.QueueUpdateDraw(func() {
					v.footer.errText.Clear()
				})
			}
			if time.Since(v.footer.msg.Time) > notificationLifetime {
				v.app.QueueUpdateDraw(func() {
					v.footer.msgText.Clear()
				})
			}

			if v.client.HasToken() {
				v.app.QueueUpdateDraw(func() {
					v.footer.statusText.Clear()
				})
			} else {
				v.app.QueueUpdateDraw(func() {
					v.footer.statusText.SetText("You are not logged in to the server, the data is not synced.")
				})
			}
		case <-v.closeCh:
			return
		}
	}
}

func (v *View) initSubview() {
	var (
		kh  common.KeyHandlerFnc
		hlp common.Help
	)
	switch v.currentPage { //nolint:exhaustive
	case common.Auth:
		kh, hlp = v.subviews.auth.Init()
		v.app.SetFocus(v.subviews.auth.Frame)
	case common.EmailVerification:
		kh, hlp = v.subviews.email.Init()
		v.app.SetFocus(v.subviews.email.Frame)
	case common.ListItems:
		kh, hlp = v.subviews.list.Init()
		v.app.SetFocus(v.subviews.list.Frame)
	case common.ChooseItemType:
		kh, hlp = v.subviews.choose.Init()
		v.app.SetFocus(v.subviews.choose.Frame)
	case common.Password:
		kh, hlp = v.subviews.password.Init()
		v.app.SetFocus(v.subviews.password.Frame)
	case common.Text:
		kh, hlp = v.subviews.text.Init()
		v.app.SetFocus(v.subviews.text.Frame)
	case common.Card:
		kh, hlp = v.subviews.card.Init()
		v.app.SetFocus(v.subviews.card.Frame)
	case common.Binary:
		kh, hlp = v.subviews.binary.Init()
		v.app.SetFocus(v.subviews.binary.Frame)
	}
	v.root.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlX {
			go func() {
				ctx, cancel := context.WithTimeout(context.TODO(), common.StandartTimeout)
				defer cancel()

				if err := v.client.Logout(ctx); err != nil {
					v.msgCh <- common.NewErrMsg(err)
				}

				v.msgCh <- common.ToViewMsg{
					ViewType: common.Auth,
				}
			}()
			return event
		}
		if kh == nil {
			return event
		}
		return kh(event)
	})

	sb := strings.Builder{}
	sb.WriteString(string(hlp))
	if v.client.HasLocalCredintials() {
		sb.WriteString("ctrl+x logout • ")
	}
	sb.WriteString("ctrl+c quit")

	v.footer.helpText.SetText(sb.String())
}

func (v *View) toItem(value any) {
	var (
		vitem vault.Item
		dv    any
	)

	switch value := value.(type) {
	case cvault.ItemType:
		vitem.Type = value
	case string:
		var err error
		if vitem, dv, err = item.Get(v.client, value); err != nil {
			v.app.QueueUpdateDraw(func() {
				v.footer.errText.SetText("Error: " + err.Error())
			})
			return
		}
	default:
		v.currentPage = common.ListItems
		return
	}

	switch vitem.Type { // nolint:exhaustive
	case cvault.Password:
		v.currentPage = common.Password
		if err := v.subviews.password.Update(vitem, dv); err != nil {
			err = common.NewErrMsg(err)
			v.app.QueueUpdateDraw(func() {
				v.footer.errText.SetText("Error: " + err.Error())
			})
		}
	case cvault.Card:
		v.currentPage = common.Card
		if err := v.subviews.card.Update(vitem, dv); err != nil {
			err = common.NewErrMsg(err)
			v.app.QueueUpdateDraw(func() {
				v.footer.errText.SetText("Error: " + err.Error())
			})
		}
	case cvault.Text:
		v.currentPage = common.Text
		if err := v.subviews.text.Update(vitem, dv); err != nil {
			err = common.NewErrMsg(err)
			v.app.QueueUpdateDraw(func() {
				v.footer.errText.SetText("Error: " + err.Error())
			})
		}
	case cvault.Binary:
		v.currentPage = common.Binary
		if err := v.subviews.binary.Update(vitem, dv); err != nil {
			err = common.NewErrMsg(err)
			v.app.QueueUpdateDraw(func() {
				v.footer.errText.SetText("Error: " + err.Error())
			})
		}
	}
}
