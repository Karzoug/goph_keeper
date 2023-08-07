package view

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/binary"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/card"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/choose"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/password"
	"github.com/Karzoug/goph_keeper/client/internal/view/item/text"
	cvault "github.com/Karzoug/goph_keeper/common/model/vault"
)

func (v *view) updateItemView(msg tea.Msg) tea.Cmd {
	var (
		vitem     vault.Item
		decrValue any
		isNewItem bool
	)

	switch msg := msg.(type) {
	case choose.SuccessfulMsg:
		vitem.Type = cvault.ItemType(msg)
		isNewItem = true
	case item.SuccessfulGetItemMsg:
		vitem = msg.Item
		decrValue = msg.DecryptedValue
	default:
		return nil
	}

	switch vitem.Type {
	case cvault.Password:
		v.currentViewType = vc.Password
		dv, _ := decrValue.(vault.Password)
		v.subviews.password = password.New(v.client, vitem, dv, isNewItem)
		return nil
	case cvault.Card:
		v.currentViewType = vc.Card
		dv, _ := decrValue.(vault.Card)
		v.subviews.card = card.New(v.client, vitem, dv, isNewItem)
		return nil
	case cvault.Text:
		v.currentViewType = vc.Text
		dv, _ := decrValue.(vault.Text)
		v.subviews.text = text.New(v.client, vitem, dv, isNewItem)
		return nil
	case cvault.Binary:
		v.currentViewType = vc.Binary
		dv, _ := decrValue.(vault.Binary)
		v.subviews.binary = binary.New(v.client, vitem, dv, isNewItem)
		return v.subviews.binary.Init()
	case cvault.BinaryLarge:
		// TODO: implement me
		panic("implement me")
	default:
		return tea.Batch(vc.ToViewCmd(vc.ListItems), func() tea.Msg {
			return vc.ErrMsg{
				Err:  client.ErrAppInternal.Error(),
				Time: time.Now(),
			}
		})
	}
}
