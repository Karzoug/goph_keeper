package view

import (
	tea "github.com/charmbracelet/bubbletea"

	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/email"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/choose"
	"github.com/Karzoug/goph_keeper/client/internal/view/list"
	"github.com/Karzoug/goph_keeper/client/internal/view/login"
	"github.com/Karzoug/goph_keeper/client/internal/view/register"
)

func switchToAnotherSubview(v *view, msg vc.ToViewMsg) (tea.Model, tea.Cmd) {
	v.currentViewType = msg.ViewType

	switch msg.ViewType { //nolint:exhaustive // missing cases in item view logic
	case vc.Login:
		v.subviews.login = login.New(v.client)
		return v, nil
	case vc.Register:
		v.subviews.register = register.New(v.client)
		return v, nil
	case vc.EmailVerification:
		v.subviews.emailVerification = email.New(v.client)
		return v, nil
	case vc.ListItems:
		v.subviews.listItems = list.New(v.client)
		switch v.currentCredsState {
		case online:
			return v, tea.Sequence(list.ListIDNameCmd(v.client), list.SyncCmd(v.client))
		case standalone:
			return v, list.ListIDNameCmd(v.client)
		default:
			return v, vc.ToViewCmd(vc.Login)
		}
	case vc.Item:
		if v.currentCredsState == nothing {
			return v, vc.ToViewCmd(vc.Login)
		}
	case vc.ChooseItemType:
		v.subviews.chooseItemType = choose.New()
		if v.currentCredsState == nothing {
			return v, vc.ToViewCmd(vc.Login)
		}
		return v, nil
	}

	return v, nil
}
