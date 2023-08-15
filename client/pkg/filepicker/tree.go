package filepicker

import (
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const nameOfTree = "Tree"

type tree struct {
	*tview.TreeView
	originalRootNode *tview.TreeNode
	window           *Window
}

func newTree(window *Window) *tree {
	rootDir := SharedConfig.RootPath
	root := tview.NewTreeNode(rootDir).
		SetColor(tcell.ColorRed).
		SetReference(newNodeReference(rootDir, true, nil))
	tree := &tree{
		TreeView: tview.NewTreeView().
			SetRoot(root).
			SetCurrentNode(root),
		originalRootNode: root,
		window:           window,
	}
	tree.addNode(root, rootDir)

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		tree.expandOrAddNode(node)
	})

	return tree
}

func (tree tree) name() string {
	return nameOfTree
}

func (tree tree) view() tview.Primitive {
	return tree.TreeView
}

func (tree tree) GetCurrentPath() string {
	nodeReference := extractNodeReference(tree.GetCurrentNode())
	if nodeReference == nil {
		return ""
	}
	return nodeReference.path
}

func (tree *tree) addNode(directoryNode *tview.TreeNode, path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		node := createTreeNode(file.Name(), file.IsDir(), directoryNode)
		directoryNode.AddChild(node)
	}
}

func (tree tree) expandOrAddNode(node *tview.TreeNode) {
	reference := node.GetReference()
	if reference == nil {
		return
	}
	nodeReference := reference.(*nodeReference)
	if !nodeReference.isDir {
		return
	}

	children := node.GetChildren()
	if len(children) == 0 {
		path := nodeReference.path
		tree.addNode(node, path)
	} else {
		node.SetExpanded(!node.IsExpanded())
	}
}
