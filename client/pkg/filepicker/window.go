package filepicker

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type transition interface {
	tview.Primitive
	AddAndSwitchToPage(name string, item tview.Primitive, resize bool) *tview.Pages
	RemovePage(name string) *tview.Pages
}

type Window struct {
	*root
	transition
	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
}

func NewWindow(width, height int) *Window {
	window := &Window{}
	window.root = newRoot(window)
	window.transition = newPages(
		window.root,
	)
	window.SetRect(0, 0, width, height)
	return window
}

// Confirm for tview.Primitive
func (window Window) Draw(screen tcell.Screen) {
	window.transition.Draw(screen)
}
func (window Window) GetRect() (x int, y int, width int, height int) {
	return window.transition.GetRect()
}
func (window Window) SetRect(x, y, width, height int) {
	window.transition.SetRect(x, y, width, height)
}
func (window *Window) GetFieldHeight() int {
	_, _, _, height := window.transition.GetRect()
	return height
}
func (window *Window) GetFieldWidth() int {
	_, _, width, _ := window.transition.GetRect()
	return width
}
func (window *Window) GetLabel() string {
	return ""
}
func (window *Window) SetDisabled(disabled bool) tview.FormItem {
	return window
}
func (window *Window) SetFinishedFunc(handler func(key tcell.Key)) tview.FormItem {
	window.finished = handler
	return window
}
func (window *Window) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) tview.FormItem {
	return window
}
func (window Window) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if key := event.Key(); key == tcell.KeyTab && window.finished != nil {
			window.finished(key)
			return
		}
		h := window.transition.InputHandler()
		h(event, setFocus)
	}
}
func (window Window) Focus(delegate func(p tview.Primitive)) {
	window.transition.Focus(delegate)
}
func (window Window) Blur() {
	window.transition.Blur()
}
