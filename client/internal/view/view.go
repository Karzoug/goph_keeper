package view

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	c "github.com/Karzoug/goph_keeper/client/internal/client"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/email"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/binary"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/card"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/choose"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/password"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/text"
	"github.com/Karzoug/goph_keeper/client/internal/view/list"
	"github.com/Karzoug/goph_keeper/client/internal/view/login"
	"github.com/Karzoug/goph_keeper/client/internal/view/register"
)

const (
	notificationTickerInterval    = 5 * time.Second
	notificationVisibilityTimeout = 10 * time.Second
)

const (
	nothing credentialsState = iota
	standalone
	online
)

type (
	view struct {
		client            *c.Client
		currentCredsState credentialsState
		currentViewType   vc.ViewType
		subviews          subviews
		err               vc.ErrMsg
		msg               vc.MsgMsg
	}
	subviews struct {
		listItems         list.View
		login             login.View
		register          register.View
		emailVerification email.View
		chooseItemType    choose.View
		password          password.View
		card              card.View
		text              text.View
		binary            binary.View
	}
	tickMsg          struct{}
	credentialsState int8
)

func New(client *c.Client) view {
	return view{client: client}
}

func (v view) Init() tea.Cmd {
	if v.client.HasLocalCredintials() {
		return tea.Batch(vc.ToViewCmd(vc.ListItems), tick())
	}

	return tea.Batch(vc.ToViewCmd(vc.Login), tick())
}

// Return the updated view to the Bubble Tea runtime for processing and
// a command if necessary.
func (v view) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// make sure these key always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch k := msg.Type; k { //nolint:exhaustive
		case tea.KeyCtrlC:
			return v, tea.Quit
		case tea.KeyCtrlX:
			return v, v.logoutCmd
		}
	}

	if _, ok := msg.(tickMsg); ok {
		if time.Since(v.msg.Time) > notificationVisibilityTimeout {
			v.msg.Msg = ""
		}
		if time.Since(v.err.Time) > notificationVisibilityTimeout {
			v.err.Err = ""
		}
		return v, tick()
	}

	// handle error if exists
	if msg, ok := msg.(vc.ErrMsg); ok {
		v.err = msg
		return v, nil
	}

	// handle event message if exists
	if msg, ok := msg.(vc.MsgMsg); ok {
		v.msg = msg
		return v, nil
	}

	if cmd := updateCredsState(&v); cmd != nil {
		return v, cmd
	}

	// switch to another subview
	if msg, ok := msg.(vc.ToViewMsg); ok {
		return switchToAnotherSubview(&v, msg)
	}

	// handle messages according to the current subview
	var cmd tea.Cmd
	switch v.currentViewType {
	case vc.Register:
		cmd = v.subviews.register.Update(msg)
	case vc.Login:
		cmd = v.subviews.login.Update(msg)
	case vc.ListItems:
		cmd = v.subviews.listItems.Update(msg)
	case vc.EmailVerification:
		cmd = v.subviews.emailVerification.Update(msg)
	case vc.Item: // switch to add/edit item by type here
		cmd = v.updateItemView(msg)
	case vc.ChooseItemType:
		cmd = v.subviews.chooseItemType.Update(msg)
	case vc.Password:
		cmd = v.subviews.password.Update(msg)
	case vc.Card:
		cmd = v.subviews.card.Update(msg)
	case vc.Text:
		cmd = v.subviews.text.Update(msg)
	case vc.Binary:
		cmd = v.subviews.binary.Update(msg)
	}
	return v, cmd
}

func (v view) View() string {
	body := new(strings.Builder)
	help := new(strings.Builder)

	fmt.Fprintf(body, "Goph Keeper: your password manager & vault app\nversion: %s", v.client.Version())

	switch v.currentViewType { //nolint:exhaustive // missing view.item is transitional view type, uses only to switch to another view
	case vc.Register:
		v.subviews.register.View(body, help)
	case vc.Login:
		v.subviews.login.View(body, help)
	case vc.EmailVerification:
		v.subviews.emailVerification.View(body, help)
	case vc.ListItems:
		v.subviews.listItems.View(body, help)
	case vc.ChooseItemType:
		v.subviews.chooseItemType.View(body, help)
	case vc.Password:
		v.subviews.password.View(body, help)
	case vc.Text:
		v.subviews.text.View(body, help)
	case vc.Card:
		v.subviews.card.View(body, help)
	case vc.Binary:
		v.subviews.binary.View(body, help)
	}

	if v.currentCredsState == standalone &&
		v.currentViewType > vc.EmailVerification {
		body.WriteString("\n\nYou are not logged in to the server, the data is not synced.")
		if v.currentViewType != vc.Register {
			help.WriteString("ctrl+l login • ")
		}
	}

	if v.msg.Msg != "" {
		fmt.Fprintf(body, "\n\nMesssage: %s: %s", v.msg.Time.Format(time.TimeOnly), v.msg.Msg)
	}

	if v.err.Err != "" {
		fmt.Fprintf(body, "\n\nError: %s: %s", v.err.Time.Format(time.TimeOnly), v.err.Err)
	}

	body.WriteString("\n\n")

	if v.currentCredsState != nothing {
		help.WriteString("ctrl+x logout • ")
	}

	help.WriteString("ctrl+c quit")
	body.WriteString(vc.HelpStyle.Render(help.String()))
	return body.String()
}

func updateCredsState(v *view) tea.Cmd {
	if v.client.HasToken() && v.client.HasLocalCredintials() {
		v.currentCredsState = online
		return nil
	}
	if v.client.HasLocalCredintials() {
		v.currentCredsState = standalone
		return nil
	}
	v.currentCredsState = nothing
	if v.currentViewType != vc.Login && v.currentViewType != vc.Register {
		v.currentViewType = vc.Login
		return vc.ToViewCmd(vc.Login)
	}

	return nil
}

func tick() tea.Cmd {
	return tea.Tick(notificationTickerInterval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (v view) logoutCmd() tea.Msg {
	ctx, cancel := context.WithTimeout(context.TODO(), vc.StandartTimeout)
	defer cancel()

	err := v.client.Logout(ctx)
	if err != nil {
		return vc.ErrMsg{
			Time: time.Now(),
			Err:  err.Error(),
		}
	}
	return vc.MsgMsg{
		Time: time.Now(),
		Msg:  "You are logged out!",
	}
}
