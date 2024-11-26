package btree

import (
	"bytes"
)

/*
	Declare Item and Node in B-Tree data structure
*/

type Item struct {
	key []byte
	val []byte
}

type Node struct {
	items        [maxItems]*Item
	children     [maxChildren]*Node
	numItems     int
	numChildrens int
}

func (n *Node) isLeaf() bool {
	return n.numChildrens == 0
}

func (n *Node) search(key []byte) (int, bool) {
	low, high := 0, n.numItems

	var mid int
	for low < high {
		mid = (low + high) / 2
		cmp := bytes.Compare(key, n.items[mid].key)

		switch {
		case cmp > 0:
			low = mid + 1
		case cmp < 0:
			high = mid
		case cmp == 0:
			return mid, true
		}
	}

	return low, false
}

func (n *Node) insertItemAt(pos int, i *Item) {
	if pos < n.numItems {
		copy(n.items[pos+1:n.numItems+1], n.items[pos:n.numItems])
	}

	n.items[pos] = i
	n.numItems++
}

func (n *Node) insertChildAt(pos int, c *Node) {
	if pos < n.numChildrens {
		copy(n.children[pos+1:n.numChildrens+1], n.children[pos:n.numChildrens])
	}

	n.children[pos] = c
	n.numChildrens++;
}

// Return the middle Item and new Node to move to the parent node
func (n *Node) split() (*Item, *Node) {
	// Node with full capacity = 2*N - 1 -> N is mid item
	mid := minItems
	midItem := n.items[mid]

	newNode := &Node{}
	copy(newNode.items[:], n.items[mid+1:])
	newNode.numItems = minItems

	if !n.isLeaf() {
		copy(newNode.children[:], n.children[mid+1:])
		newNode.numChildrens = minItems + 1
	}

	for i, l := mid, n.numItems; i < l; i++ {
		n.items[i] = nil
		n.numItems--

		if !n.isLeaf() {
			n.children[i+1] = nil
			n.numChildrens--
		}
	}

	return midItem, newNode
}

func (n *Node) insert(item* Item) bool {
	pos, found := n.search(item.key)

	if found {
		n.items[pos] = item
		return false
	}

	if n.isLeaf() {
		n.insertItemAt(pos, item)
		return true
	}

	if n.children[pos].numItems >= maxItems {
		midItem, newNode := n.children[pos].split()
		n.insertItemAt(pos, midItem)
		n.insertChildAt(pos, newNode)

		switch cmp := bytes.Compare(item.key, n.items[pos].key); {
		case cmp < 0:
			// The key looking for still smaller than the key of the middle -> Take the direction
		case cmp > 0:
			pos++
		default:
			n.items[pos] = item
			return true
		}
	}

	return n.children[pos].insert(item)
}