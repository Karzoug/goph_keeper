package view

import tea "github.com/charmbracelet/bubbletea"

type toViewMsg struct {
	viewType viewType
}

func toLoginView() tea.Msg {
	return toViewMsg{
		viewType: login,
	}
}

func toRegisterView() tea.Msg {
	return toViewMsg{
		viewType: register,
	}
}

func toEmailVerificationView() tea.Msg {
	return toViewMsg{
		viewType: emailVerification,
	}
}

func toListItemsView() tea.Msg {
	return toViewMsg{
		viewType: listItems,
	}
}

func toItemView() tea.Msg {
	return toViewMsg{
		viewType: item,
	}
}

func switchToAnotherSubview(v *view, msg toViewMsg) (tea.Model, tea.Cmd) {
	v.currentViewType = msg.viewType

	switch msg.viewType {
	case login:
		v.subviews.login = initialLoginView()
		return v, initLoginView()
	case register:
		v.subviews.register = initialRegisterView()
		return v, initRegisterView()
	case emailVerification:
		v.subviews.emailVerification = initialEmailVerificationView()
		return v, initEmailVerificationView()
	case listItems:
		v.subviews.listItems = initialListItemsView()
		switch v.currentCredsState {
		case online:
			return v, tea.Sequence(v.listItemsNames, v.updateListItems)
		case standalone:
			return v, v.listItemsNames
		default:
			return v, toLoginView
		}
	case item:
		v.subviews.item = initialItemView()
		if v.currentCredsState == nothing {
			return v, toLoginView
		}
		return v, nil // nil, because data loaded by cmd from list view
	}

	return v, nil
}
