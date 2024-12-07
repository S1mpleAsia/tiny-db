package btree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

/*
	Stores metadata for the B-Tree:
	-	`root`: The root node of the B-Tree
	-	`file`: The file used for storing the B-Tree nodes
	-	`blockSize`: The size of a single block in bytes
	-	`numBlocks`: The total number of blocks currently in used
	-	`degree`: The minimum children node of the B-Tree
*/
type BTree struct {
	Root *Node
	File *os.File
	BlockSize int
	NumBlocks int
	Degree int
}

/*
	# Properties
	-	minChildren = degree
	-	maxChildren = 2*degree
	-	minKeys = degree -1
	-	maxKeys = 2*degree - 1
	
	File Format (Suppose block size = 4096):

	+---------+----------+----------+	   +--------+--------+	
	| KL | CL |	   K1	 |	  K2    |  ... |   C1   |   C2   | ...
	+---------+----------+----------+      +--------+--------+
	  2    2	   4		  4					4		4

	KL: Key length
	CL: Children length

	Number of Keys = n 		   | => Children pointer start at 4096 / 2
	Number of Children = n+1   | 
*/
func NewBTree(fileName string) (*BTree, error) {
	file, err := os.OpenFile(fileName, os.O_RDWR | os.O_CREATE, 0666)

	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()

	if err != nil {
		return nil, err
	}

	blockSize := 4096 // Default block size
	capacity := blockSize /  4
	maxKeys := capacity/2 - 1
	degree := (maxKeys + 1) / 2
	numBlocks := int(stat.Size()) / blockSize

	return &BTree{
		File: file,
		BlockSize: blockSize,
		Degree: degree,
		NumBlocks: numBlocks,
	}, nil
}

// Allocate a new free block for the B-Tree.
// A node in B-Tree corresponding to a Block in Disk-based storage 
func (btree *BTree) FreeBlock() int {
	return btree.NumBlocks
}

func (btree *BTree) Seek(block int) error {
	offset := int64(btree.BlockSize * block)
	_, err := btree.File.Seek(offset, 0)
	return err
}

// Read a node at a given block number
func (btree *BTree) ReadNode(block int) (*Node, error) {
	if err := btree.Seek(block); err != nil {
		return nil, err
	}

	buffer := make([]byte, btree.BlockSize)
	_, err := btree.File.Read(buffer)

	if err != nil {
		return nil, err
	}

	node := NewNode()
	node.Block = block

	keysLen := binary.LittleEndian.Uint16(buffer[:2])
	childrenLen := binary.LittleEndian.Uint16(buffer[2:4])
	
	i := 4

	for k := 0; k < int(keysLen); k++ {
		key := int(binary.LittleEndian.Uint32(buffer[i:i+4]))
		node.Keys = append(node.Keys, key)
		i += 4
	}

	i = btree.BlockSize / 2
	for c := 0; c < int(childrenLen); c++ {
		child := int(binary.LittleEndian.Uint32(buffer[i:i+4]))
		node.Children = append(node.Children, child)
		i += 4
	}

	return node, nil
}

func (btree *BTree) WriteNode(node *Node) error {
	if err := btree.Seek(node.Block); err != nil {
		return err
	}
	
	buffer := make([]byte, btree.BlockSize)

	binary.LittleEndian.PutUint16(buffer[:2], uint16(len(node.Keys)))
	binary.LittleEndian.PutUint16(buffer[2:4], uint16(len(node.Children)))

	i := 4

	for _, key := range node.Keys {
		binary.LittleEndian.PutUint32(buffer[i:i+4], uint32(key))
		i += 4
	}

	i = btree.BlockSize / 2
	for _, child := range node.Children {
		binary.LittleEndian.PutUint32(buffer[i:i+4], uint32(child))
		i += 4
	}

	if _, err := btree.File.Write(buffer); err != nil {
		return err
	}

	if node.Block == btree.NumBlocks {
		btree.NumBlocks++
	}

	return nil
}

// Split a child node at specified index
func (btree *BTree) SplitChild(parent *Node, index int) error {
	targetNode, err := btree.ReadNode(parent.Children[index])

	if err != nil {
		return err
	}

	newNode := NewNode()
	newNode.Block = btree.FreeBlock()

	// Move keys from the child node to the new node
	// Maximum keys is 2*degree -1 => Need to split at degree
	newNode.Keys = append(newNode.Keys, targetNode.Keys[btree.Degree:]...)

	// Move the children if not leaf
	if len(targetNode.Children) > 0 {
		newNode.Children = append(newNode.Children, targetNode.Children[btree.Degree:]...)
		targetNode.Children = targetNode.Children[:btree.Degree]
	}

	// Add the median key to the parent node
	medianKey := targetNode.Keys[btree.Degree - 1]
	parent.Keys = append(parent.Keys[:index], append([]int{medianKey}, parent.Keys[index:]...)...)
	targetNode.Keys = targetNode.Keys[:btree.Degree-1]

	// Insert new node to the parent
	parent.Children = append(parent.Children[:index+1], append([]int{newNode.Block}, parent.Children[index+1:]...)...)

	if err := btree.WriteNode(targetNode); err != nil {
		return err
	}

	if err := btree.WriteNode(newNode); err != nil {
		return err
	}

	return btree.WriteNode(parent)
}

func (btree *BTree) Insert(key int) error {
	if btree.Root == nil {
		root := NewNodeWithKeys([]int{key})
		root.Block = btree.FreeBlock()
		btree.Root = root
		return btree.WriteNode(root)
	}

	root := btree.Root
	if len(root.Keys) == (2*btree.Degree - 1) {
		newRoot := NewNode()
		newRoot.Block = btree.FreeBlock()
		newRoot.Children = append(newRoot.Children, root.Block)

		if err := btree.SplitChild(newRoot, 0); err != nil {
			return err
		}

		btree.Root = newRoot
	}
	return btree.insertInto(btree.Root, key)
}

func (btree *BTree) insertInto(node *Node, key int) error {
	if len(node.Children) == 0 {
		index := 0
		for index < len(node.Keys) && key > node.Keys[index] {
			index++
		}

		node.Keys = append(node.Keys[:index], append([]int{key}, node.Keys[index:]...)...)
		return btree.WriteNode(node)
	}

	index := 0
	for index < len(node.Keys) && key > node.Keys[index] {
		index++
	}

	child, err := btree.ReadNode(node.Children[index])
	if err != nil {
		return err
	}

	if len(child.Keys) == 2 * btree.Degree - 1 {
		if err := btree.SplitChild(node, index); err != nil {
			return err
		}

		if key > node.Keys[index] {
			index++
		}

		child, err = btree.ReadNode(node.Children[index])
		if err != nil {
			return err
		}
	}

	return btree.insertInto(child, key)
}

func (btree *BTree) Json(node *Node) (string, error) {
	if node == nil {
		return "{}", nil
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf(`"block": %d, "key": %v, "children": [`, node.Block, node.Keys))

	for i, child := range node.Children {
		if i > 0 {
			buffer.WriteString(",")
		}

		childNode, err := btree.ReadNode(child)

		if err != nil {
			return "", nil
		}

		childJson, err := btree.Json(childNode)
		if err != nil {
			return "", err
		}

		buffer.WriteString(childJson)
	}

	buffer.WriteString("]}")

	return buffer.String(), nil
}
