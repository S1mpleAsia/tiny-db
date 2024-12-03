package main

import "s1mpleasia.com/tinydb/btree"

func main() {
	// app := cli.NewCLI()

	// app.Start()

	tree := btree.NewBTree[int](2)
	tree.Insert(1)
	tree.Insert(2)
	tree.Insert(3)
	tree.Insert(4)
	tree.Insert(5)
	tree.Insert(6)
	tree.Insert(7)
	tree.Insert(8)
	tree.Insert(9)

	btree.PrintTree(tree.Root(), 0)
}