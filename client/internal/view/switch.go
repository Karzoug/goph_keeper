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
		v.currentCredsState = nothing
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
		if v.client.HasToken() && v.client.HasLocalCredintials() {
			v.currentCredsState = online
			return v, tea.Sequence(v.listItemsNames, v.updateListItems)
		}
		if v.client.HasLocalCredintials() {
			v.currentCredsState = standalone
			return v, v.listItemsNames
		}
	case item:
		v.subviews.item = initialItemView()
		if v.client.HasLocalCredintials() {
			return v, nil // nil, because data loaded by cmd from list view
		}
	}

	return v, nil
}
