package filepicker

import "github.com/rivo/tview"

type nodeReference struct {
	path       string
	isDir      bool
	parentNode *tview.TreeNode
}

func newNodeReference(path string, isDir bool, parentNode *tview.TreeNode) *nodeReference {
	return &nodeReference{
		path:       path,
		isDir:      isDir,
		parentNode: parentNode,
	}
}
