package filepicker

import (
	"github.com/rivo/tview"
)

type root struct {
	*tview.Grid
	*tree
}

func newRoot(window *Window) *root {
	tree := newTree(window)
	view := &root{
		Grid: tview.NewGrid().
			SetRows(0).
			SetColumns(30, 0),
	}
	view.tree = tree
	view.addTree()
	return view
}

func (view *root) addTree() {
	view.
		AddItem(view.tree.TreeView, 0, 0, 1, 1, 0, 0, true)
}

func (view *root) removeTree() {
	view.RemoveItem(view.tree.TreeView)
}
