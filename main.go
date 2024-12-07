package main

import (
	"os"
)

func main() {
	// app := cli.NewCLI()

	// app.Start()

	// tree := btree.NewBTree[int](2)
	// tree.Insert(1)
	// tree.Insert(2)
	// tree.Insert(3)
	// tree.Insert(4)
	// tree.Insert(5)
	// btree.PrintTree(tree.Root(), 0)

	dbDir := "./test"
	// tablePath := path.Join(dbDir, "filetest")
	os.MkdirAll(dbDir, 0777)
}