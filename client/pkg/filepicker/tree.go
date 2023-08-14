package filepicker

import (
	"io/ioutil"
	"path/filepath"

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
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		node := createTreeNode(file.Name(), file.IsDir(), directoryNode)
		directoryNode.AddChild(node)
	}
}

func (tree tree) addNodeAll(node *tview.TreeNode) {
	if !extractNodeReference(node).isDir {
		return
	}
	children := node.GetChildren()
	for _, child := range children {
		if extractNodeReference(child).isDir {
			tree.addNodeAll(child)
		}
	}
	tree.expandOrAddNode(node)
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

func (tree tree) expandAll(node *tview.TreeNode) {
	if !extractNodeReference(node).isDir {
		return
	}

	for _, child := range node.GetChildren() {
		if extractNodeReference(child).isDir {
			tree.expandAll(child)
		}
	}

	tree.expandOrAddNode(node)
}

func (tree tree) setAllDisplayTextToPath(node *tview.TreeNode) {
	for _, child := range node.GetChildren() {
		tree.setAllDisplayTextToPath(child)
	}

	nodeReference := extractNodeReference(node)
	node.SetText(nodeReference.path)
}

func (tree tree) setAllDisplayTextToBasename(node *tview.TreeNode) {
	for _, child := range node.GetChildren() {
		tree.setAllDisplayTextToBasename(child)
	}

	nodeReference := extractNodeReference(node)
	path := filepath.Base(nodeReference.path)
	node.SetText(path)
}

func lastNodes(node *tview.TreeNode) []*tview.TreeNode {
	nodes := []*tview.TreeNode{}
	children := node.GetChildren()

	if len(children) > 0 {
		for _, child := range children {
			nodes = append(nodes, lastNodes(child)...)
		}
		return nodes
	}

	nodes = append(nodes, node)
	return nodes
}
