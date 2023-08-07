package common

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	CursorStyle  = FocusedStyle.Copy()
	NoStyle      = lipgloss.NewStyle()
	HelpStyle    = BlurredStyle.Copy()

	FocusedButton = FocusedStyle.Copy().Render("[ Submit ]")
	BlurredButton = fmt.Sprintf("[ %s ]", BlurredStyle.Render("Submit"))
)

const (
	Login ViewType = iota
	Register
	EmailVerification
	ListItems
	Item // transitional view type, only to switch to another view
	ChooseItemType
	Password
	Card
	Text
	Binary
)

type (
	ViewType int8
	MsgMsg   struct {
		Time time.Time
		Msg  string
	}
	ErrMsg struct {
		Time time.Time
		Err  string
	}
	ToViewMsg struct {
		ViewType ViewType
	}
)

func ToViewCmd(t ViewType) tea.Cmd {
	return func() tea.Msg {
		return ToViewMsg{
			ViewType: t,
		}
	}
}

func ShowMsgCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return MsgMsg{
			Msg:  msg,
			Time: time.Now(),
		}
	}
}

func ShowErrCmd(err string) tea.Cmd {
	return func() tea.Msg {
		return ErrMsg{
			Err:  err,
			Time: time.Now(),
		}
	}
}
