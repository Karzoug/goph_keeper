package binary

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/Karzoug/goph_keeper/client/internal/client"
	"github.com/Karzoug/goph_keeper/client/internal/model/vault"
	vc "github.com/Karzoug/goph_keeper/client/internal/view/common"
	"github.com/Karzoug/goph_keeper/client/internal/view/item"
)

var (
	decryptFocusedButton = vc.FocusedStyle.Copy().Render("[ Decrypt ]")
	decryptBlurredButton = fmt.Sprintf("[ %s ]", vc.BlurredStyle.Render("Decrypt"))

	encryptFocusedButton = vc.FocusedStyle.Copy().Render("[ Encrypt ]")
	encryptBlurredButton = fmt.Sprintf("[ %s ]", vc.BlurredStyle.Render("Encrypt"))
)

type View struct {
	client               *client.Client
	item                 vault.Item
	value                vault.Binary
	isNewItem            bool
	focusIndex           int
	nameInput            textinput.Model
	saveFilenameInput    textinput.Model
	filepicker           filepicker.Model
	selectedInFilepicker string
}

func New(c *client.Client, item vault.Item, b vault.Binary, isNewItem bool) View {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = isNewItem
	fp.Height = 10
	fp.CurrentDirectory, _ = os.UserHomeDir()

	t := textinput.New()
	t.Cursor.Style = vc.CursorStyle
	t.CharLimit = 128
	t.Placeholder = "Name"
	t.Focus()
	t.PromptStyle = vc.FocusedStyle
	t.TextStyle = vc.FocusedStyle

	if !isNewItem {
		t.SetValue(item.Name)
	}

	v := View{
		client:     c,
		item:       item,
		isNewItem:  isNewItem,
		filepicker: fp,
		nameInput:  t,
		value:      b,
	}

	if !isNewItem {
		v.saveFilenameInput = textinput.New()
		v.saveFilenameInput.Cursor.Style = vc.CursorStyle
		v.saveFilenameInput.CharLimit = 128
		v.saveFilenameInput.Placeholder = "Filename"
		v.saveFilenameInput.PromptStyle = vc.FocusedStyle
		v.saveFilenameInput.TextStyle = vc.FocusedStyle
	}

	_ = b

	return v
}

func (v View) Init() tea.Cmd {
	return v.filepicker.Init()
}

func (v *View) Update(msg tea.Msg) tea.Cmd { //nolint:gocyclo // TODO: change UI library
	updateFilePicker := func() tea.Cmd {
		var cmd tea.Cmd
		v.filepicker, cmd = v.filepicker.Update(msg)

		// Did the user select a file?
		if didSelect, path := v.filepicker.DidSelectFile(msg); didSelect {
			// Get the path of the selected file.
			v.selectedInFilepicker = path
		}
		// Did the user select a disabled file?
		if didSelect, path := v.filepicker.DidSelectDisabledFile(msg); didSelect {
			v.selectedInFilepicker = ""
			return tea.Batch(cmd, vc.ShowErrCmd(path+" is not valid."))
		}

		return cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s := msg.Type; s {
		case tea.KeyEscape:
			if v.focusIndex != 1 {
				return vc.ToViewCmd(vc.ListItems)
			}
		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			if v.isNewItem {
				if s == tea.KeyEnter && v.focusIndex == 2 {
					return v.cmd()
				}

				if s == tea.KeyShiftTab {
					v.focusIndex--
				}
				if s == tea.KeyTab {
					v.focusIndex++
				}

				if v.focusIndex == 1 {
					if s == tea.KeyUp || s == tea.KeyDown || s == tea.KeyEnter {
						return updateFilePicker()
					}
				} else {
					if s == tea.KeyUp {
						v.focusIndex--
					}
					if s == tea.KeyDown || s == tea.KeyEnter {
						v.focusIndex++
					}
				}

				if v.focusIndex > 2 {
					v.focusIndex = 0
				} else if v.focusIndex < 0 {
					v.focusIndex = 2
				}

				// name input
				if v.focusIndex == 0 {
					v.nameInput.PromptStyle = vc.FocusedStyle
					v.nameInput.TextStyle = vc.FocusedStyle
					return v.nameInput.Focus()
				} else {
					v.nameInput.Blur()
					v.nameInput.PromptStyle = vc.NoStyle
					v.nameInput.TextStyle = vc.NoStyle
				}
			} else {
				if s == tea.KeyEnter && v.focusIndex == 3 {
					return v.cmd()
				}

				if s == tea.KeyShiftTab {
					v.focusIndex--
				}
				if s == tea.KeyTab {
					v.focusIndex++
				}

				if v.focusIndex == 1 {
					if s == tea.KeyUp || s == tea.KeyDown || s == tea.KeyEnter {
						return updateFilePicker()
					}
				} else {
					if s == tea.KeyUp {
						v.focusIndex--
					}
					if s == tea.KeyDown || s == tea.KeyEnter {
						v.focusIndex++
					}
				}

				if v.focusIndex > 3 {
					v.focusIndex = 0
				} else if v.focusIndex < 0 {
					v.focusIndex = 3
				}

				cmds := make([]tea.Cmd, 2)

				// name input
				if v.focusIndex == 0 {
					v.nameInput.PromptStyle = vc.FocusedStyle
					v.nameInput.TextStyle = vc.FocusedStyle
					cmds[0] = v.nameInput.Focus()
				} else {
					v.nameInput.Blur()
					v.nameInput.PromptStyle = vc.NoStyle
					v.nameInput.TextStyle = vc.NoStyle
				}
				// filename input
				if v.focusIndex == 2 {
					v.saveFilenameInput.PromptStyle = vc.FocusedStyle
					v.saveFilenameInput.TextStyle = vc.FocusedStyle
					cmds[1] = v.saveFilenameInput.Focus()
				} else {
					v.saveFilenameInput.Blur()
					v.saveFilenameInput.PromptStyle = vc.NoStyle
					v.saveFilenameInput.TextStyle = vc.NoStyle
				}

				return tea.Batch(cmds...)
			}
		default:
		}
	case item.SuccessfulSetItemMsg:
		return tea.Batch(vc.ToViewCmd(vc.ListItems),
			vc.ShowMsgCmd("Saved!"))
	}

	cmds := make([]tea.Cmd, 3)
	v.nameInput, cmds[0] = v.nameInput.Update(msg)
	v.saveFilenameInput, cmds[1] = v.saveFilenameInput.Update(msg)
	cmds[2] = updateFilePicker()

	return tea.Batch(cmds...)
}

func (v View) View(body *strings.Builder, help *strings.Builder) {
	if v.isNewItem {
		body.WriteString("\n\nAdd new binary data:\n\n")
	} else {
		body.WriteString("\n\nEdit binary data:\n\n")
	}

	body.WriteString(v.nameInput.View())
	body.WriteString("\n\n")

	if v.isNewItem {
		body.WriteString("\nPick a file to encrypt and save:")
	} else {
		body.WriteString("\nPick a folder, enter a filename to decrypt and save the data:")
	}
	body.WriteString("\n\n" + v.filepicker.View() + "\n")

	if !v.isNewItem {
		body.WriteString("\n\n" + v.saveFilenameInput.View() + "\n")
		button := &decryptBlurredButton
		if v.focusIndex == 3 {
			button = &decryptFocusedButton
		}
		fmt.Fprintf(body, "\n\n%s", *button)
	} else {
		button := &encryptBlurredButton
		if v.focusIndex == 2 {
			button = &encryptFocusedButton
		}
		fmt.Fprintf(body, "\n\n%s", *button)
	}

	help.WriteString("tab next • shift+tab prev • esc back • ")
}

func (v View) cmd() tea.Cmd {
	v.item.Name = v.nameInput.Value()
	v.item.ClientUpdatedAt = time.Now().Unix()

	if v.isNewItem {
		return createCmd(v.client, v.item, v.selectedInFilepicker)
	} else {
		return saveOnDiskCmd(v.client, v.item, v.value, v.filepicker.CurrentDirectory, v.saveFilenameInput.Value())
	}
}
