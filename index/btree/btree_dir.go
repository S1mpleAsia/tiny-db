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

func (bd *BTreeDir) Close() error {
	return bd.contents.Close()
}

func (bd *BTreeDir) Search(searchKey *record.Constant) (int32, error) {
	childBlk, err := bd.findChildBlock(searchKey)
	if err != nil {
		return 0, err
	}

	for {
		flag, err := bd.contents.GetFlag()
		if err != nil {
			return 0, err
		}

		if flag <= 0 {
			break
		}

		if err := bd.contents.Close(); err != nil {
			return 0, err
		}

		bp, err := NewBTreePage(bd.tx, childBlk, bd.layout)
		if err != nil {
			return 0, err
		}

		bd.contents = bp
		childBlk, err = bd.findChildBlock(searchKey)
		if err != nil {
			return 0, err
		}
	}

	return int32(childBlk.BlockNumber()), nil
}

func (bd *BTreeDir) MakeNewRoot(e *DirEntry) error {
	firstVal, err := bd.contents.GetDataVal(0)
	if err != nil {
		return err
	}

	flag, err := bd.contents.GetFlag()
	if err != nil {
		return err
	}

	newBlock, err := bd.contents.Split(0, flag)
	if err != nil {
		return err
	}

	oldRoot := NewDirEntry(firstVal, int32(newBlock.BlockNumber()))
	if _, err := bd.InsertEntry(oldRoot); err != nil {
		return err
	}

	if _, err := bd.InsertEntry(e); err != nil {
		return err
	}

	if err = bd.contents.SetFlag(flag + 1); err != nil {
		return err
	}

	return nil
}

func (bd *BTreeDir) Insert(e *DirEntry) (*DirEntry, error) {
	flag, err := bd.contents.GetFlag()
	if err != nil {
		return nil, err
	}

	if flag == 0 {
		return bd.InsertEntry(e)
	}
	childBlock, err := bd.findChildBlock(e.dataval)
	if err != nil {
		return nil, err
	}

	child, err := NewBTreeDir(bd.tx, childBlock, bd.layout)
	if err != nil {
		return nil, err
	}

	myEntry, err := child.Insert(e)
	if err != nil {
		return nil, err
	}

	if err = child.Close(); err != nil {
		return nil, err
	}

	if myEntry == nil {
		return nil, nil
	}

	return bd.InsertEntry(myEntry)
}

func (bd *BTreeDir) InsertEntry(e *DirEntry) (*DirEntry, error) {
	slot, err := bd.contents.FindSlotBefore(e.dataval)
	if err != nil {
		return nil, err
	}

	newSlot := 1 + slot
	if err := bd.contents.InsertDir(newSlot, e.dataval, e.block); err != nil {
		return nil, err
	}

	if isFull, err := bd.contents.IsFull(); err != nil {
		return nil, err
	} else if !isFull {
		return nil, nil
	}

	flag, err := bd.contents.GetFlag()
	if err != nil {
		return nil, err
	}

	numRecs, err := bd.contents.GetNumRecs()
	if err != nil {
		return nil, err
	}

	splitPos := numRecs / 2
	splitVal, err := bd.contents.GetDataVal(splitPos)
	if err != nil {
		return nil, err
	}

	newBlock, err := bd.contents.Split(splitPos, flag)
	if err != nil {
		return nil, err
	}

	return NewDirEntry(splitVal, int32(newBlock.BlockNumber())), nil
}

func (bd *BTreeDir) findChildBlock(searchKey *record.Constant) (*file.BlockId, error) {
	slot, err := bd.contents.FindSlotBefore(searchKey)
	if err != nil {
		return &file.BlockId{}, err
	}

	val, err := bd.contents.GetDataVal(slot + 1)
	if err != nil {
		return &file.BlockId{}, err
	}

	if val.Equals(searchKey) {
		slot++
	}

	blkNum, err := bd.contents.GetChildNum(slot)
	if err != nil {
		return &file.BlockId{}, err
	}

	return file.NewBlockId(bd.fileName, int64(blkNum)), nil
}
