package btree

import (
	"fmt"
	"strings"
)

/*
	minChildren = degree
	maxChildren = 2*degree
	minKeys = degree -1
	maxKeys = 2*degree - 1
*/

// Represents a B-Tree with a given degree
type BTree[T any] struct {
	root *Node[T]
	degree int
}

func NewBTree[T any](degree int) *BTree[T] {
	return &BTree[T]{
		degree: degree,
	}
}

func (tree *BTree[T]) Root() *Node[T] {
	return tree.root
}

// Split a child node at specified index
func (tree *BTree[T]) SplitChild(parent *Node[T], index int) {
	child := parent.children[index]
	newNode := NewEmptyNode[T]()

	// Move keys from the child node to the new node
	// Maximum keys is 2*degree -1 => Need to split at degree
	newNode.keys = append(newNode.keys, child.keys[tree.degree:]...)

	// Move the children if not leaf
	if len(child.children) > 0 {
		newNode.children = append(newNode.children, child.children[tree.degree:]...)
		child.children = child.children[:tree.degree]
	}

	// Add the median key to the parent node
	parent.keys = append(parent.keys[:index], append([]T{child.keys[tree.degree - 1]}, parent.keys[index:]...)...)
	child.keys = child.keys[:tree.degree-1]

	// Insert new node to the parent
	parent.children = append(parent.children[:index+1], append([]*Node[T]{newNode}, parent.children[index+1:]...)...)
}

func (tree *BTree[T]) Insert(key T) {
	if tree.root == nil {
		tree.root = NewNode([]T{key}, []*Node[T]{})
		return
	}

	if len(tree.root.keys) == (2*tree.degree - 1) {
		oldRoot := tree.root
		newRoot := NewNode([]T{}, []*Node[T]{oldRoot})
		tree.SplitChild(newRoot, 0)
		tree.root = newRoot
		tree.insertNonFull(tree.root, key)
	} else {
		tree.insertNonFull(tree.root, key)
	}
}

func (tree *BTree[T]) insertNonFull(node *Node[T], key T) {
	i := len(node.keys) - 1

	if len(node.children) == 0 {
		node.keys = append(node.keys, key)

		for i >= 0 && less(key, node.keys[i]) {
			node.keys[i+1] = node.keys[i]
			i--
		}

		node.keys[i+1] = key
	} else {
		for i >= 0 && less(key, node.keys[i]) {
			i--
		}

		i++

		if len(node.children[i].keys) == 2*tree.degree - 1 {
			tree.SplitChild(node, i)

			if less(node.keys[i], key) {
				i++
			}
		}
		tree.insertNonFull(node.children[i], key)
	}
}

func PrintTree[T any](node *Node[T], level int) {
	if node == nil {
		return
	}

	fmt.Printf("%s%v\n", strings.Repeat("  ", level), node.keys)

	for _, child := range node.children {
		PrintTree(child, level + 1)
	}
}

// Helper function to compare 2 generic values
func less[T any](a, b T) bool {
	return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
}