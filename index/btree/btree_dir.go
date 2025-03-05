package btree

import (
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

type BTreeDir struct {
	tx       *transaction.Transaction
	layout   *record.Layout
	contents *BTreePage
	fileName string
}

func NewBTreeDir(tx *transaction.Transaction, block *file.BlockId, layout *record.Layout) (*BTreeDir, error) {
	bp, err := NewBTreePage(tx, block, layout)
	if err != nil {
		return nil, err
	}

	return &BTreeDir{
		tx:       tx,
		layout:   layout,
		contents: bp,
		fileName: block.FileName(),
	}, nil
}

func (bd *BTreeDir) Close() {}

func (bd *BTreeDir) Search(searchKey *record.Constant) int32 {
	return 0
}

func (bd *BTreeDir) MakeNewRoot(e *DirEntry) {

}

func (bd *BTreeDir) Insert(e *DirEntry) *DirEntry {
	return nil
}

func (bd *BTreeDir) InsertEntry(e *DirEntry) *DirEntry {
	return nil
}
func (bd *BTreeDir) findChildBlock(searchKey *record.Constant) *file.BlockId {
	return nil
}
