package btree

import (
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

/*	BTree Page contains common codes for directory and leaf blocks (e.g insert entries, split array of entries) */
type BTreePage struct {
	tx           *transaction.Transaction
	currentBlock *file.BlockId
	layout       *record.Layout
}

func NewBTreePage(tx *transaction.Transaction, currentBlock *file.BlockId, layout *record.Layout) (*BTreePage, error) {
	btreePage := &BTreePage{
		tx:           tx,
		currentBlock: currentBlock,
		layout:       layout,
	}

	if err := btreePage.tx.Pin(currentBlock); err != nil {
		return nil, err
	}

	return btreePage, nil
}

func (bp *BTreePage) FindSlotBefore(searchKey *record.Constant) int {
	slot := 0
	numRecs, err := bp.GetNumRecs()

	if err != nil {
		panic(err)
	}

	compare, err := bp.GetDataVal(int32(slot)).CompareTo(searchKey)
	if err != nil {
		panic(err)
	}

	for slot < int(numRecs) && compare < 0 {
		slot++
	}

	return slot - 1
}

func (bp *BTreePage) Close() {
	if bp.currentBlock != nil {
		bp.tx.Unpin(bp.currentBlock)
	}

	bp.currentBlock = nil
}

func (bp *BTreePage) IsFull() bool {
	numRecs, err := bp.GetNumRecs()
	if err != nil {
		panic(err)
	}
	return bp.slotPos(numRecs+1) >= int32(bp.tx.BlockSize())
}

func (bp *BTreePage) Split(splitPos int32, flag int32) (*file.BlockId, error) {
	newBlock := bp.AppendNew(flag)
	newPage, err := NewBTreePage(bp.tx, newBlock, bp.layout)
	if err != nil {
		return nil, err
	}

	bp.transferRecs(splitPos, newPage)
	newPage.SetFlag(flag)
	// newPage.clo
	return newBlock, nil
}

func (bp *BTreePage) GetDataVal(slot int32) *record.Constant {
	val, err := bp.getVal(slot, "dataval")
	if err != nil {
		panic(err)
	}

	return val
}

func (bp *BTreePage) GetFlag() int32 {
	flag, err := bp.tx.GetInt(bp.currentBlock, 0)
	if err != nil {
		panic(err)
	}

	return flag
}

func (bp *BTreePage) SetFlag(val int32) {
	bp.tx.SetInt(bp.currentBlock, 0, val, true)
}

func (bp *BTreePage) AppendNew(flag int32) *file.BlockId {
	blk := bp.tx.Append(bp.currentBlock.FileName())
	if err := bp.tx.Pin(blk); err != nil {
		panic(err)
	}

	bp.Format(blk, flag)
	return blk
}

func (bp *BTreePage) Format(block *file.BlockId, flag int32) {
	bp.tx.SetInt(block, 0, flag, true)
	bp.tx.SetInt(block, file.INT_32_BITS, 0, false)

	recSize := bp.layout.SlotSize()
	for pos := 2 * file.INT_32_BITS; pos+int(recSize) < bp.tx.BlockSize(); pos += int(recSize) {
		bp.makeDefaultRecord(block, int32(pos))
	}
}

func (bp *BTreePage) makeDefaultRecord(blk *file.BlockId, pos int32) {
	for _, fldName := range bp.layout.Schema().Fields() {
		offset := bp.layout.Offset(fldName)
		if bp.layout.Schema().Type(fldName) == record.INT {
			bp.tx.SetInt(blk, pos+offset, 0, false)
		} else {
			bp.tx.SetString(blk, pos+offset, "", false)
		}
	}
}

// Method call only by BTreeDir
func (bp *BTreePage) GetChildNum(slot int32) (int32, error) {
	return bp.getInt(slot, "block")
}

func (bp *BTreePage) InsertDir(slot int32, val *record.Constant, blockNum int32) {
	bp.insert(slot)
	bp.setVal(slot, "dataval", val)
	bp.setInt(slot, "block", blockNum)
}

// Method call only by BTreeLeaf
func (bp *BTreePage) GetDataRID(slot int32) (*record.RID, error) {
	blockNum, err := bp.getInt(slot, "block")
	if err != nil {
		return nil, err
	}

	return record.NewRID(blockNum, file.INT_32_BITS), nil
}

func (bp *BTreePage) InsertLeaf(slot int32, val *record.Constant, rid *record.RID) {
	bp.insert(slot)
	bp.setVal(slot, "dataval", val)
	bp.setInt(slot, "block", rid.BlockNumber())
	bp.setInt(slot, "id", rid.Slot())
}

func (bp *BTreePage) Delete(slot int32) {
	numRecs, err := bp.GetNumRecs()
	if err != nil {
		panic(err)
	}

	for i := slot + 1; i < numRecs; i++ {
		bp.copyRecords(i, i-1)
	}

	bp.setNumRecs(numRecs - 1)
}

func (bp *BTreePage) GetNumRecs() (int32, error) {
	return bp.tx.GetInt(bp.currentBlock, file.INT_32_BITS)
}

// Private methods
func (bp *BTreePage) getInt(slot int32, fldName string) (int32, error) {
	pos := bp.fldPos(slot, fldName)
	return bp.tx.GetInt(bp.currentBlock, int(pos))
}

func (bp *BTreePage) getString(slot int32, fldName string) (string, error) {
	pos := bp.fldPos(slot, fldName)
	return bp.tx.GetString(bp.currentBlock, int(pos))
}

func (bp *BTreePage) getVal(slot int32, fldName string) (*record.Constant, error) {
	dataType := bp.layout.Schema().Type(fldName)

	if dataType == record.INT {
		value, err := bp.getInt(slot, fldName)
		if err != nil {
			return nil, err
		}

		return record.NewConstantWithInt(value), nil
	} else {
		value, err := bp.getString(slot, fldName)
		if err != nil {
			return nil, err
		}

		return record.NewConstantWithString(value), nil
	}
}

func (bp *BTreePage) setInt(slot int32, fldName string, val int32) {
	pos := bp.fldPos(slot, fldName)
	bp.tx.SetInt(bp.currentBlock, pos, val, true)
}

func (bp *BTreePage) setString(slot int32, fldName string, val string) {
	pos := bp.fldPos(slot, fldName)
	bp.tx.SetString(bp.currentBlock, pos, val, true)
}

func (bp *BTreePage) setVal(slot int32, fldName string, val *record.Constant) {
	dataType := bp.layout.Schema().Type(fldName)

	if dataType == record.INT {
		intVal, err := val.AsInt()
		if err != nil {
			panic(err)
		}

		bp.setInt(slot, fldName, intVal)
	} else {
		strVal, err := val.AsString()
		if err != nil {
			panic(err)
		}

		bp.setString(slot, fldName, strVal)
	}
}

func (bp *BTreePage) setNumRecs(n int32) {
	bp.tx.SetInt(bp.currentBlock, file.INT_32_BITS, n, true)
}

func (bp *BTreePage) insert(slot int32) {
	numRecs, err := bp.GetNumRecs()
	if err != nil {
		panic(err)
	}

	for i := numRecs; i > slot; i-- {
		bp.copyRecords(i-1, i)
	}

	bp.setNumRecs(numRecs + 1)
}

func (bp *BTreePage) copyRecords(from int32, to int32) {
	sch := bp.layout.Schema()

	for _, fldName := range sch.Fields() {
		val, err := bp.getVal(from, fldName)
		if err != nil {
			panic(err)
		}

		bp.setVal(to, fldName, val)
	}
}

func (bp *BTreePage) transferRecs(slot int32, dest *BTreePage) {
	destSlot := 0
	numRecs, err := bp.GetNumRecs()

	if err != nil {
		panic(err)
	}

	for slot < numRecs {
		dest.insert(int32(destSlot))
		sch := bp.layout.Schema()

		for _, fldName := range sch.Fields() {
			val, err := bp.getVal(slot, fldName)
			if err != nil {
				panic(err)
			}
			dest.setVal(int32(destSlot), fldName, val)
		}

		bp.Delete(slot)
		destSlot++
	}
}

func (bp *BTreePage) fldPos(slot int32, fldName string) int32 {
	offset := bp.layout.Offset(fldName)
	return bp.slotPos(slot) + offset
}

func (bp *BTreePage) slotPos(slot int32) int32 {
	slotSize := bp.layout.SlotSize()
	return file.INT_32_BITS + file.INT_32_BITS*(slot*slotSize)
}
