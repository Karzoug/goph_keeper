package filepicker

import (
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func extractNodeReference(node *tview.TreeNode) *nodeReference {
	return node.GetReference().(*nodeReference)
}

func createTreeNode(fileName string, isDir bool, parent *tview.TreeNode) *tview.TreeNode {
	var parentPath string

	if parent == nil {
		parentPath = SharedConfig.RootPath
	} else {
		reference, ok := parent.GetReference().(*nodeReference)
		if !ok {
			parentPath = SharedConfig.RootPath
		} else {
			parentPath = reference.path
		}
	}

	var color tcell.Color
	if isDir {
		color = tcell.ColorGreen
	} else {
		color = tview.Styles.PrimaryTextColor
	}

	return tview.NewTreeNode(fileName).
		SetReference(
			newNodeReference(
				filepath.Join(parentPath, fileName),
				isDir,
				parent,
			),
		).
		SetSelectable(true).
		SetColor(color)
}
