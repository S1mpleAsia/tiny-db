package btree

import (
	"s1mpleasia.com/tinydb/record"
)

type DirEntry struct {
	dataval *record.Constant
	block   int32
}

func NewDirEntry(dataval *record.Constant, block int32) *DirEntry {
	return &DirEntry{
		dataval: dataval,
		block:   block,
	}
}
