package btree

import (
	"fmt"
	"math"
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ query.Index = (*BTreeIndex)(nil)

type BTreeIndex struct {
	tx         *transaction.Transaction
	dirLayout  *record.Layout
	leafLayout *record.Layout
	leafTbl    string
	leaf       *BTreeLeaf
	rootBlock  *file.BlockId
}

func NewBTreeIndex(tx *transaction.Transaction, idxName string, leafLayout *record.Layout) (*BTreeIndex, error) {
	leafTable := idxName + "leaf"
	size := tx.Size(leafTable)
	if size == 0 {
		blk := tx.Append(leafTable)
		node, err := NewBTreePage(tx, blk, leafLayout)
		if err != nil {
			return nil, err
		}

		if err = node.Format(blk, -1); err != nil {
			return nil, err
		}
	}

	dirSchema := record.NewSchema()
	dirSchema.Add("block", leafLayout.Schema())
	dirSchema.Add("dataval", leafLayout.Schema())

	dirTable := idxName + "dir"
	dirLayout := record.NewLayoutFromSchema(dirSchema)

	rootBlk := file.NewBlockId(dirTable, 0)
	size = tx.Size(dirTable)
	if size == 0 {
		tx.Append(dirTable)

		node, err := NewBTreePage(tx, rootBlk, dirLayout)
		if err != nil {
			return nil, err
		}
		if err := node.Format(rootBlk, 0); err != nil {
			return nil, err
		}

		fldType := dirSchema.Type("dataval")
		var minVal *record.Constant
		switch fldType {
		case record.INT:
			minVal = record.NewConstantWithInt(math.MinInt32)
		case record.VARCHAR:
			minVal = record.NewConstantWithString("")
		default:
			return nil, fmt.Errorf("unexpected value type: %d", fldType)
		}

		if err := node.InsertDir(0, minVal, 0); err != nil {
			return nil, err
		}

		if err := node.Close(); err != nil {
			return nil, err
		}
	}

	return &BTreeIndex{
		tx:         tx,
		dirLayout:  dirLayout,
		leafLayout: leafLayout,
		leafTbl:    leafTable,
		rootBlock:  rootBlk,
	}, nil
}

func (bi *BTreeIndex) BeforeFirst(searchKey *record.Constant) error {
	err := bi.Close()
	if err != nil {
		return err
	}

	root, err := NewBTreeDir(bi.tx, bi.rootBlock, bi.dirLayout)
	if err != nil {
		return err
	}

	blkNum, err := root.Search(searchKey)
	if err != nil {
		return err
	}

	if err := root.Close(); err != nil {
		return err
	}

	blk := file.NewBlockId(bi.leafTbl, int64(blkNum))
	leaf, err := NewBTreeLeaf(bi.tx, blk, bi.leafLayout, searchKey)
	if err != nil {
		return err
	}

	bi.leaf = leaf
	return nil
}

func (bi *BTreeIndex) Next() (bool, error) {
	return bi.leaf.Next()
}

func (bi *BTreeIndex) GetDataRID() (*record.RID, error) {
	return bi.leaf.GetDataRID()
}

func (bi *BTreeIndex) Insert(dataVal *record.Constant, dataRID *record.RID) error {
	if err := bi.BeforeFirst(dataVal); err != nil {
		return err
	}
	entry, err := bi.leaf.Insert(dataRID)
	if err != nil {
		return err
	}

	if err := bi.leaf.Close(); err != nil {
		return err
	}

	if entry == nil {
		return nil
	}

	root, err := NewBTreeDir(bi.tx, bi.rootBlock, bi.dirLayout)
	if err != nil {
		return nil
	}

	entry2, err := root.Insert(entry)
	if err != nil {
		return err
	}

	if entry2 != nil {
		if err := root.MakeNewRoot(entry2); err != nil {
			return err
		}
	}

	if err := root.Close(); err != nil {
		return err
	}

	return nil
}

func (bi *BTreeIndex) Delete(dataVal *record.Constant, dataRID *record.RID) error {
	if err := bi.BeforeFirst(dataVal); err != nil {
		return err
	}
	if err := bi.leaf.Delete(dataRID); err != nil {
		return err
	}
	return nil
}

func (bi *BTreeIndex) Close() error {
	if bi.leaf != nil {
		return bi.leaf.Close()
	}
	return nil
}

func SearchCost(numBlocks, rpb int32) int32 {
	return 1 + int32(math.Log(float64(numBlocks))/math.Log(float64(rpb)))
}
