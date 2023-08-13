package card

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
)

type View struct {
	client     *client.Client
	item       vault.Item
	isNewItem  bool
	focusIndex int
	inputs     []textinput.Model
}

func New(c *client.Client, item vault.Item, card vault.Card, isNewItem bool) View {
	v := View{
		client:    c,
		item:      item,
		inputs:    make([]textinput.Model, 5),
		isNewItem: isNewItem,
	}

	var t textinput.Model
	for i := range v.inputs {
		t = textinput.New()
		t.Cursor.Style = vc.CursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			if !isNewItem {
				t.SetValue(item.Name)
			}
			t.Placeholder = "Name"
			t.Focus()
			t.PromptStyle = vc.FocusedStyle
			t.TextStyle = vc.FocusedStyle
			t.CharLimit = 128
		case 1:
			if !isNewItem {
				t.SetValue(card.Number)
			}
			t.Placeholder = "Number"
			t.Validate = validateOnlyNumbersFn
			t.PromptStyle = vc.NoStyle
			t.TextStyle = vc.NoStyle
			// For the most popular card types
			// the maximum card number length is up to 19 digits.
			t.CharLimit = 19
		case 2:
			if !isNewItem {
				t.SetValue(card.Holder)
			}
			t.Placeholder = "CardHolder"
			t.Validate = validateOnlyLettersAndSpacesFn
			t.PromptStyle = vc.NoStyle
			t.TextStyle = vc.NoStyle
			// Since 2014, VISA and MasterCard are shortens maximum length
			// for Cardholder names to 21 characters and 22 characters respectively.
			// Set here 40, because there might be older cards with longer names.
			t.CharLimit = 40
		case 3:
			if !isNewItem {
				t.SetValue(card.Expired)
			}
			t.Placeholder = "Expired"
			t.Validate = validateExpiredDateFn
			t.PromptStyle = vc.NoStyle
			t.TextStyle = vc.NoStyle
			t.CharLimit = 5
		case 4:
			if !isNewItem {
				t.SetValue(card.CSC)
			}
			t.Placeholder = "CVV/CVC"
			t.Validate = validateOnlyNumbersFn
			t.PromptStyle = vc.NoStyle
			t.TextStyle = vc.NoStyle
			// 4-digit for AMEX, 3 for all other.
			t.CharLimit = 4
		}
		v.inputs[i] = t
	}
	return v
}

func (v *View) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s := msg.Type; s {
		case tea.KeyEsc:
			for i := 0; i < len(v.inputs); i++ {
				v.inputs[i].Reset()
			}
			return vc.ToViewCmd(vc.ListItems)

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			// Did the user press enter while the submit button was focused?
			if s == tea.KeyEnter && v.focusIndex == len(v.inputs) {
				return v.cmd()
			}

			if s == tea.KeyUp || s == tea.KeyShiftTab {
				v.focusIndex--
			} else {
				v.focusIndex++
			}

			if v.focusIndex > len(v.inputs) {
				v.focusIndex = 0
			} else if v.focusIndex < 0 {
				v.focusIndex = len(v.inputs)
			}

			cmds := make([]tea.Cmd, len(v.inputs))
			for i := 0; i <= len(v.inputs)-1; i++ {
				if i == v.focusIndex {
					// Set focused state
					cmds[i] = v.inputs[i].Focus()
					v.inputs[i].PromptStyle = vc.FocusedStyle
					v.inputs[i].TextStyle = vc.FocusedStyle
					continue
				}
				// Remove focused state
				v.inputs[i].Blur()
				v.inputs[i].PromptStyle = vc.NoStyle
				v.inputs[i].TextStyle = vc.NoStyle
			}

			return tea.Batch(cmds...)
		default:
		}

	case item.SuccessfulSetItemMsg:
		return tea.Batch(vc.ToViewCmd(vc.ListItems),
			vc.ShowMsgCmd("Saved!"))
	case item.ConflictVersionSetItemMsg:
		return tea.Batch(vc.ToViewCmd(vc.ListItems),
			vc.ShowMsgCmd("Saved!"),
			vc.ShowErrCmd(client.ErrConflictVersion.Error()))
	}

	// Handle character input and blinking
	cmd := v.updateCardViewInputs(msg)

	return cmd
}

func (v *View) updateCardViewInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(v.inputs))

	for i := range v.inputs {
		v.inputs[i], cmds[i] = v.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (v View) View(body *strings.Builder, help *strings.Builder) {
	if v.isNewItem {
		body.WriteString("\n\nAdd new card:\n")
	} else {
		body.WriteString("\n\nEdit card:\n")
	}

	for i := range v.inputs {
		body.WriteString(v.inputs[i].View())
		if i < len(v.inputs)-1 {
			body.WriteRune('\n')
		}
	}

	button := &vc.BlurredButton
	if v.focusIndex == len(v.inputs) {
		button = &vc.FocusedButton
	}
	fmt.Fprintf(body, "\n\n%s", *button)

	help.WriteString("tab next • shift+tab prev • esc back • ")
}

var validateOnlyNumbersFn textinput.ValidateFunc = func(s string) error {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return fmt.Errorf("contains not number")
		}
	}
	return nil
}

var validateOnlyLettersAndSpacesFn textinput.ValidateFunc = func(s string) error {
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsSpace(r)) {
			return fmt.Errorf("contains not letters")
		}
	}
	return nil
}

var validateExpiredDateFn textinput.ValidateFunc = func(s string) error {
	for i := 0; i < len(s); i++ {
		if (s[i] < '0' || s[i] > '9') && s[i] != '/' {
			return fmt.Errorf("invalid expired date")
		}
	}
	return nil
}

func (v View) cmd() tea.Cmd {
	v.item.Name = v.inputs[0].Value()

	return item.SetCmd(v.client,
		v.item,
		vault.Card{
			Holder:  v.inputs[2].Value(),
			Expired: v.inputs[3].Value(),
			Number:  v.inputs[1].Value(),
			CSC:     v.inputs[4].Value(),
		})
}
