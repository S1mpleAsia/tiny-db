package btree

// Node represents in the B-Tree
type Node[T any] struct {
	keys     []T
	children []*Node[T]
}

func NewNode[T any] (keys []T, children []*Node[T]) *Node[T] {
	return &Node[T]{
		keys: keys,
		children: children,
	}
}

func NewEmptyNode[T any]() *Node[T] {
	return &Node[T]{}
}

