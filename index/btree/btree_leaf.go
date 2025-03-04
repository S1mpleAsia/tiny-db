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

	return &BTreeLeaf{
		tx:          tx,
		layout:      layout,
		searchKey:   searchKey,
		contents:    contents,
		currentSlot: contents.FindSlotBefore(searchKey),
		fileName:    block.FileName(),
	}
}
func (bLeaf *BTreeLeaf) Close() {
	bLeaf.contents.Close()
}

func (bLeaf *BTreeLeaf) Next() bool {
	bLeaf.currentSlot++
	numRecs, err := bLeaf.contents.GetNumRecs()
	if err != nil {
		panic(err)
	}

	if bLeaf.currentSlot >= numRecs {
		// tryOverflow()
		return true
	} else if bLeaf.contents.GetDataVal(bLeaf.currentSlot).Equals(bLeaf.searchKey) {
		return true
	} else {
		return true
	}
}
