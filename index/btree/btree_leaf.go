package btree

import (
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

// BTreeLeaf stores the actual index array of dataval
type BTreeLeaf struct {
	tx          *transaction.Transaction
	layout      *record.Layout
	searchKey   *record.Constant
	contents    *BTreePage
	currentSlot int32
	fileName    string
}

func NewBTreeLeaf(tx *transaction.Transaction, block *file.BlockId, layout *record.Layout, searchKey *record.Constant) *BTreeLeaf {
	contents := &BTreePage{tx, block, layout}

	slot, err := contents.FindSlotBefore(searchKey)
	if err != nil {
		panic(err)
	}

	return &BTreeLeaf{
		tx:          tx,
		layout:      layout,
		searchKey:   searchKey,
		contents:    contents,
		currentSlot: slot,
		fileName:    block.FileName(),
	}
}
func (bLeaf *BTreeLeaf) Close() {
	bLeaf.contents.Close()
}

func (bLeaf *BTreeLeaf) Next() bool {
	return false
}

func (bLeaf *BTreeLeaf) GetDataRID() *record.RID {
	return nil
}

func (bLeaf *BTreeLeaf) Delete(dataRID *record.RID) {
}

func (bLeaf *BTreeLeaf) Insert(dataRID *record.RID) *DirEntry {
	return nil
}

func (bLeaf *BTreeLeaf) tryOverFlow() bool {
	return false
}
