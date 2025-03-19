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

func NewBTreeLeaf(tx *transaction.Transaction, block *file.BlockId, layout *record.Layout, searchKey *record.Constant) (*BTreeLeaf, error) {
	contents, err := NewBTreePage(tx, block, layout)
	if err != nil {
		return nil, err
	}

	currentSlot, err := contents.FindSlotBefore(searchKey)
	if err != nil {
		return nil, err
	}

	return &BTreeLeaf{tx, layout, searchKey, contents, currentSlot, block.FileName()}, nil
}
func (bLeaf *BTreeLeaf) Close() error {
	return bLeaf.contents.Close()
}

func (bLeaf *BTreeLeaf) Next() (bool, error) {
	bLeaf.currentSlot++
	numRecs, err := bLeaf.contents.GetNumRecs()

	if bLeaf.currentSlot >= numRecs {
		return bLeaf.tryOverFlow()
	}

	val, err := bLeaf.contents.GetDataVal(bLeaf.currentSlot)
	if err != nil {
		return false, err
	}

	if val.Equals(bLeaf.searchKey) {
		return true, nil
	} else {
		return bLeaf.tryOverFlow()
	}
}

func (bLeaf *BTreeLeaf) GetDataRID() (*record.RID, error) {
	return bLeaf.contents.GetDataRID(bLeaf.currentSlot)
}

func (bLeaf *BTreeLeaf) Delete(dataRID *record.RID) error {
	for {
		ok, err := bLeaf.Next()
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		rid, err := bLeaf.GetDataRID()
		if err != nil {
			return err
		}
		if rid.Equals(dataRID) {
			return bLeaf.contents.Delete(bLeaf.currentSlot)
		}
	}
}

func (bLeaf *BTreeLeaf) Insert(dataRID *record.RID) (*DirEntry, error) {
	flag, err := bLeaf.contents.GetFlag()
	if err != nil {
		return nil, err
	}

	firstVal, err := bLeaf.contents.GetDataVal(0)
	if err != nil {
		return nil, err
	}

	cmp, err := firstVal.CompareTo(bLeaf.searchKey)
	if err != nil {
		return nil, err
	}

	if flag >= 0 && cmp > 0 {
		newBlock, err := bLeaf.contents.Split(0, flag)
		if err != nil {
			return nil, err
		}

		bLeaf.currentSlot = 0
		if err = bLeaf.contents.SetFlag(-1); err != nil {
			return nil, err
		}

		if err = bLeaf.contents.InsertLeaf(bLeaf.currentSlot, bLeaf.searchKey, dataRID); err != nil {
			return nil, err
		}

		return NewDirEntry(firstVal, int32(newBlock.BlockNumber())), nil
	}

	bLeaf.currentSlot++
	if err := bLeaf.contents.InsertLeaf(bLeaf.currentSlot, bLeaf.searchKey, dataRID); err != nil {
		return nil, err
	}

	if isFull, err := bLeaf.contents.IsFull(); err != nil {
		return nil, err
	} else if !isFull {
		return nil, nil
	}

	firstKey, err := bLeaf.contents.GetDataVal(0)
	if err != nil {
		return nil, err
	}

	numRecs, err := bLeaf.contents.GetNumRecs()
	if err != nil {
		return nil, err
	}

	lastKey, err := bLeaf.contents.GetDataVal(numRecs - 1)
	if err != nil {
		return nil, err
	}

	if lastKey.Equals(firstKey) {
		newBlock, err := bLeaf.contents.Split(1, flag)
		if err != nil {
			return nil, err
		}

		if err = bLeaf.contents.SetFlag(int32(newBlock.BlockNumber())); err != nil {
			return nil, err
		}

		return nil, nil
	} else {
		splitPos := numRecs / 2
		splitKey, err := bLeaf.contents.GetDataVal(splitPos)
		if err != nil {
			return nil, err
		}

		if splitKey.Equals(firstKey) {
			for {
				val, err := bLeaf.contents.GetDataVal(splitPos)
				if err != nil {
					return nil, err
				}

				if !val.Equals(splitKey) {
					break
				}
				splitPos++
			}
			splitKey, err = bLeaf.contents.GetDataVal(splitPos)
			if err != nil {
				return nil, err
			}
		} else {
			for {
				val, err := bLeaf.contents.GetDataVal(splitPos - 1)
				if err != nil {
					return nil, err
				}

				if !val.Equals(splitKey) {
					break
				}
				splitPos--
			}
		}

		newBlock, err := bLeaf.contents.Split(splitPos, -1)
		if err != nil {
			return nil, err
		}

		return NewDirEntry(splitKey, int32(newBlock.BlockNumber())), nil
	}
}

func (bLeaf *BTreeLeaf) tryOverFlow() (bool, error) {
	firstVal, err := bLeaf.contents.GetDataVal(0)
	if err != nil {
		return false, err
	}

	flag, err := bLeaf.contents.GetFlag()
	if err != nil {
		return false, err
	}

	if !firstVal.Equals(bLeaf.searchKey) || flag < 0 {
		return false, nil
	}

	if err = bLeaf.contents.Close(); err != nil {
		return false, err
	}

	nextBlock := file.NewBlockId(bLeaf.fileName, int64(flag))
	contents, err := NewBTreePage(bLeaf.tx, nextBlock, bLeaf.layout)
	if err != nil {
		return false, err
	}

	bLeaf.contents = contents
	bLeaf.currentSlot = 0

	return true, nil
}
