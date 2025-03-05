package btree

import (
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/index"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ index.Index = (*BTreeIndex)(nil)

type BTreeIndex struct {
	tx        *transaction.Transaction
	dirLayout *record.Layout
	leafTbl   string
	leaf      *BTreeLeaf
	rootBlock *file.BlockId
}

func NewBTreeIndex(tx *transaction.Transaction, idxName string, leafLayout *record.Layout) *BTreeIndex {
	return nil
}

func (B BTreeIndex) BeforeFirst(searchKey *record.Constant) error {
	//TODO implement me
	panic("implement me")
}

func (B BTreeIndex) Next() (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (B BTreeIndex) GetDataRID() (*record.RID, error) {
	//TODO implement me
	panic("implement me")
}

func (B BTreeIndex) Insert(dataval *record.Constant, datarid *record.RID) error {
	//TODO implement me
	panic("implement me")
}

func (B BTreeIndex) Delete(dataval *record.Constant, datarid *record.RID) error {
	//TODO implement me
	panic("implement me")
}

func (B BTreeIndex) Close() error {
	//TODO implement me
	panic("implement me")
}
