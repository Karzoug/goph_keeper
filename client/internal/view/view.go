package view

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	c "github.com/Karzoug/goph_keeper/client/internal/client"
)

const (
	notificationTickerInterval    = 5 * time.Second
	notificationVisibilityTimeout = 15 * time.Second
)

const (
	nothing credentialsState = iota
	standalone
	online
)

const (
	login viewType = iota
	register
	emailVerification
	listItems
	item
)

type (
	credentialsState int8
	viewType         int8
)

type (
	view struct {
		client            *c.Client
		currentCredsState credentialsState
		currentViewType   viewType
		subviews          subviews
		err               errMsg
		msg               msgMsg
	}
	subviews struct {
		listItems         listItemsView
		login             loginView
		register          registerView
		emailVerification emailVerificationView
		item              itemView
	}
)

type (
	msgMsg struct {
		time time.Time
		msg  string
	}
	errMsg struct {
		time time.Time
		err  string
	}
	tickMsg struct{}
)

func New(client *c.Client) view {
	return view{client: client}
}

func (v view) Init() tea.Cmd {
	if v.client.HasLocalCredintials() {
		return tea.Batch(toListItemsView, tick())
	}

	return tea.Batch(toLoginView, tick())
}

// Return the updated view to the Bubble Tea runtime for processing and
// a command if necessary.
func (v view) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// make sure these key always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "ctrl+c" {
			return v, tea.Quit
		}
	}

	if cmd := updateCredsState(&v); cmd != nil {
		return v, cmd
	}

	if _, ok := msg.(tickMsg); ok {
		if time.Since(v.msg.time) > notificationVisibilityTimeout {
			v.msg.msg = ""
		}
		if time.Since(v.err.time) > notificationVisibilityTimeout {
			v.err.err = ""
		}
		return v, tick()
	}

	// handle error if exists
	if msg, ok := msg.(errMsg); ok {
		v.err = msg
		return v, nil
	}

	// handle event message if exists
	if msg, ok := msg.(msgMsg); ok {
		v.msg = msg
		return v, nil
	}

	// switch to another subview
	if msg, ok := msg.(toViewMsg); ok {
		return switchToAnotherSubview(&v, msg)
	}

	// handle messages according to the current subview
	switch v.currentViewType {
	case register:
		cmd := updateRegisterView(&v, msg)
		return v, cmd
	case login:
		cmd := updateLoginView(&v, msg)
		return v, cmd
	case listItems:
		cmd := updateListItemsView(&v, msg)
		return v, cmd
	case emailVerification:
		cmd := updateEmailVerificationView(&v, msg)
		return v, cmd
	case item:
		cmd := updateItemView(&v, msg)
		return v, cmd
	}

	return v, nil
}

func (v view) View() string {
	b := new(strings.Builder)

	fmt.Fprintf(b, "Goph Keeper: your password manager & vault app\nversion: %s\n\n", v.client.Version())

	switch v.currentViewType {
	case register:
		viewRegisterView(v.subviews.register, b)
	case login:
		viewLoginView(v.subviews.login, b)
	case emailVerification:
		viewEmailVerificationView(v.subviews.emailVerification, b)
	case listItems:
		viewListItemsView(v.subviews.listItems, b)
	case item:
		viewItemView(v.subviews.item, b)
	}

	printStatus(b, v.currentCredsState)

	if v.msg.msg != "" {
		fmt.Fprintf(b, "\nMesssage: %s: %s\n\n", v.msg.time.Format(time.TimeOnly), v.msg.msg)
	}

	if v.err.err != "" {
		fmt.Fprintf(b, "\nError: %s: %s\n\n", v.err.time.Format(time.TimeOnly), v.err.err)
	}

	fmt.Fprintln(b, "\n\nPress ctr + c to quit.")

	return b.String()
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
	if v.currentViewType != login {
		return toLoginView
	}

	return nil
}

func tick() tea.Cmd {
	return tea.Tick(notificationTickerInterval, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func printStatus(b *strings.Builder, s credentialsState) {
	if s == standalone {
		fmt.Fprint(b, "\nYou are offline, the data is not synced.\nPress ctrl+l to login")
	}
}
