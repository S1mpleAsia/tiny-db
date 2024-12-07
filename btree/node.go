package btree

/*
	Node represents in the B-Tree. Each node stores:
	-	`block`: The block number where this node is stored in the file
	-	`keys`: The list of keys stored in this node
	-	`children`: The list of child block pointers for this node
*/
type Node struct {
	Block int
	Keys     []int
	Children []int
}

func NewNode() *Node {
	return &Node{
		Keys: []int{},
		Children: []int{},
		Block: 0,
	}
}

func NewNodeWithKeys(keys []int) *Node {
	return &Node{
		Keys: keys,
		Children: []int{},
		Block: 0,
	}
}

// func NewNode[T any] (keys []T, children []*Node[T]) *Node[T] {
// 	return &Node[T]{
// 		keys: keys,
// 		children: children,
// 	}
// }

// func NewEmptyNode[T any]() *Node[T] {
// 	return &Node[T]{}
// }

