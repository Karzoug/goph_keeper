package common

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

type KeyHandlerFnc func(event *tcell.EventKey) *tcell.EventKey

type Help string

type Msg struct {
	Time time.Time
	msg  string
}

func NewMsg(msg string) Msg {
	return Msg{
		Time: time.Now(),
		msg:  msg,
	}
}

func (msg Msg) String() string {
	return fmt.Sprintf("%s %s", msg.Time.Format(time.TimeOnly), msg.msg)
}

type ViewType string

func (vt ViewType) String() string {
	return string(vt)
}

type ToViewMsg struct {
	ViewType ViewType
	Value    any
}

type ErrMsg struct {
	Time time.Time
	error
}

func NewErrMsg(err error) ErrMsg {
	return ErrMsg{
		Time:  time.Now(),
		error: err,
	}
}

func (msg ErrMsg) Error() string {
	return fmt.Sprintf("%s %s", msg.Time.Format(time.TimeOnly), msg.error)
}

const (
	Auth              ViewType = "Auth"
	EmailVerification ViewType = "EmailVerification"
	ListItems         ViewType = "ListItems"
	Item              ViewType = "Item" // transitional view type, only to switch to another view
	ChooseItemType    ViewType = "ChooseItemType"
	Password          ViewType = "Password"
	Card              ViewType = "Card"
	Text              ViewType = "Text"
	Binary            ViewType = "Binary"
)

const StandartTimeout = 3 * time.Second
