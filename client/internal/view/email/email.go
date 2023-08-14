package email

import (
	"context"
	"errors"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/view/common"
)

type View struct {
	Frame *tview.Frame
	input *tview.InputField

	client      *client.Client
	msgCh       chan<- any
	appUpdateFn func(func()) *tview.Application

	code string
}

func New(c *client.Client, msgCh chan<- any, appUpdateFn func(func()) *tview.Application) View {
	v := View{
		client:      c,
		msgCh:       msgCh,
		appUpdateFn: appUpdateFn,
	}

	input := tview.NewInputField().
		SetLabel("Enter the code from mail: ").
		SetFieldWidth(6).
		SetAcceptanceFunc(tview.InputFieldInteger)

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		code := v.input.GetText()
		if len(code) < 6 {
			msgCh <- errors.New("code too short")
			return
		}
		v.code = code
		input.SetDisabled(true)
		go v.cmd()
	})

	frame := tview.NewFrame(input)

	v.Frame = frame
	v.input = input

	return v
}

func (v *View) Init() (common.KeyHandlerFnc, common.Help) {
	return v.keyHandler, "esc back • "
}

func (v *View) cmd() {
	ctx, cancel := context.WithTimeout(context.TODO(), common.StandartTimeout)
	defer cancel()

	err := v.client.VerifyEmail(ctx, v.code)
	if err != nil {
		v.msgCh <- common.NewErrMsg(client.ErrInvalidEmailVerificationCode)
		v.appUpdateFn(func() {
			v.input.SetDisabled(false)
		})
		return
	}
	v.appUpdateFn(func() {
		v.input.SetDisabled(false)
		v.input.SetText("")
	})
	v.msgCh <- common.ToViewMsg{
		ViewType: common.ListItems,
	}
}

func (v *View) keyHandler(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEsc:
		v.input.SetText("")
		go func() {
			v.msgCh <- common.ToViewMsg{
				ViewType: common.Auth,
			}
		}()
	}
	return event
}
