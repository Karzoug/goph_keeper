package auth

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
	form  *tview.Form

	baseContext context.Context
	client      *client.Client
	msgCh       chan<- any
	appUpdateFn func(func()) *tview.Application

	email    string
	password string
}

func New(c *client.Client, msgCh chan<- any, appUpdateFn func(func()) *tview.Application) View {
	v := View{
		client:      c,
		msgCh:       msgCh,
		appUpdateFn: appUpdateFn,
	}
	frame := tview.NewFrame(nil).
		AddText("Enter email and password to login/register", true, tview.AlignLeft, tcell.ColorWhite)
	v.Frame = frame
	return v
}

func (v *View) Update(ctx context.Context) {
	v.baseContext = ctx
}

func (v *View) Init() (common.KeyHandlerFnc, common.Help) {
	form := tview.NewForm()
	form.SetBorderPadding(1, 1, 0, 1)
	form.AddInputField("Email", "", 35, nil, func(email string) {
		v.email = email
	})
	form.AddPasswordField("Password", "", 35, '*', func(lastName string) {
		v.password = lastName
	})
	form.AddButton("Login", func() {
		go v.loginCmd()
	})
	form.AddButton("Register", func() {
		go v.registerCmd()
	})
	v.form = form
	v.Frame.SetPrimitive(form)

	return nil, ""
}

func (v *View) loginCmd() {
	ctx, cancel := context.WithTimeout(v.baseContext, common.StandartTimeout)
	defer cancel()

	err := v.client.Login(ctx, v.email, []byte(v.password))
	if err != nil {
		if errors.Is(err, client.ErrUserEmailNotVerified) {
			// clear before go to list items
			v.email = ""
			v.password = ""
			v.appUpdateFn(func() {
				v.Frame.SetPrimitive(nil)
				v.form = nil
			})
			v.msgCh <- common.ToViewMsg{
				ViewType: common.EmailVerification,
			}
			return
		}
		v.msgCh <- common.NewErrMsg(err)
		return
	}

	// clear before go to list items
	v.email = ""
	v.password = ""
	v.appUpdateFn(func() {
		v.Frame.SetPrimitive(nil)
		v.form = nil
	})

	v.msgCh <- common.NewMsg("You are entered!")
	v.msgCh <- common.ToViewMsg{
		ViewType: common.ListItems,
	}
}

func (v *View) registerCmd() {
	ctx, cancel := context.WithTimeout(v.baseContext, common.StandartTimeout)
	defer cancel()

	err := v.client.Register(ctx, v.email, []byte(v.password))
	if err != nil {
		v.msgCh <- common.NewErrMsg(err)
		return
	}

	v.email = ""
	v.password = ""
	v.appUpdateFn(func() {
		v.Init()
	})

	v.msgCh <- common.NewMsg("You are registered!")
}
