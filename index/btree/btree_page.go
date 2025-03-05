package btree

import (
	"fmt"
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

// BTreePage
/*
	- Holding sorted list of record, act as a list-like
	- First 4-bytes stores the flag of the current block
*/
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

// Find the smallest slot x that searchkey <= dataval(x)
func (bp *BTreePage) FindSlotBefore(searchKey *record.Constant) (int32, error) {
	var slot int32 = 0

	for {
		numRecs, err := bp.GetNumRecs()
		if err != nil {
			return 0, nil
		}

		if slot == numRecs {
			break
		}

		val, err := bp.GetDataVal(slot)
		if err != nil {
			return 0, nil
		}

		cmp, err := val.CompareTo(searchKey)
		if err != nil {
			return 0, nil
		}

		if cmp >= 0 {
			break
		}
		slot++
	}

	return slot - 1, nil
}

func (bp *BTreePage) Close() error {
	emptyBlock := &file.BlockId{}
	if bp.currentBlock == emptyBlock {
		return nil
	}

	bp.tx.Unpin(bp.currentBlock)
	bp.currentBlock = emptyBlock
	return nil
}

func (bp *BTreePage) IsFull() (bool, error) {
	numRecs, err := bp.GetNumRecs()
	if err != nil {
		return false, err
	}

	return bp.slotPos(numRecs+1) >= int32(bp.tx.BlockSize()), nil
}

func (bp *BTreePage) Split(splitPos int32, flag int32) (*file.BlockId, error) {
	newBlock, err := bp.appendNew(flag)
	if err != nil {
		return &file.BlockId{}, err
	}
	newPage, err := NewBTreePage(bp.tx, newBlock, bp.layout)
	if err != nil {
		return &file.BlockId{}, err
	}
	defer newPage.Close()

	if err = bp.transferRecs(splitPos, newPage); err != nil {
		return &file.BlockId{}, err
	}

	if err = newPage.SetFlag(flag); err != nil {
		return &file.BlockId{}, err
	}
	return newBlock, nil
}

func (bp *BTreePage) GetNumRecs() (int32, error) {
	return bp.tx.GetInt(bp.currentBlock, file.INT_32_BITS)
}

func (bp *BTreePage) GetDataVal(slot int32) (*record.Constant, error) {
	return bp.getVal(slot, "dataval")
}

func (bp *BTreePage) GetFlag() (int32, error) {
	return bp.tx.GetInt(bp.currentBlock, 0)
}

func (bp *BTreePage) SetFlag(val int32) error {
	return bp.tx.SetInt(bp.currentBlock, 0, val, true)
}

func (bp *BTreePage) getVal(slot int32, fldName string) (*record.Constant, error) {
	dataType := bp.layout.Schema().Type(fldName)

	switch dataType {
	case record.INT:
		value, err := bp.getInt(slot, fldName)
		if err != nil {
			return nil, err
		}

		return record.NewConstantWithInt(value), nil
	case record.VARCHAR:
		value, err := bp.tx.GetString(bp.currentBlock, int(bp.fldPos(slot, fldName)))
		if err != nil {
			return nil, err
		}

		return record.NewConstantWithString(value), nil
	default:
		return nil, fmt.Errorf("unsupported data type: %d", dataType)
	}
}

func (bp *BTreePage) setInt(slot int32, fldName string, val int32) error {
	pos := bp.fldPos(slot, fldName)
	return bp.tx.SetInt(bp.currentBlock, pos, val, true)
}

func (bp *BTreePage) setString(slot int32, fldName string, val string) error {
	pos := bp.fldPos(slot, fldName)
	return bp.tx.SetString(bp.currentBlock, pos, val, true)
}

func (bp *BTreePage) setVal(slot int32, fldName string, val *record.Constant) error {
	dataType := bp.layout.Schema().Type(fldName)

	switch dataType {
	case record.INT:
		intVal, err := val.AsInt()
		if err != nil {
			return err
		}

		return bp.setInt(slot, fldName, intVal)
	case record.VARCHAR:
		strVal, err := val.AsString()
		if err != nil {
			return err
		}

		return bp.setString(slot, fldName, strVal)
	default:
		return fmt.Errorf("unexpected data type %d", dataType)
	}
}

func (bp *BTreePage) setNumRecs(n int32) error {
	return bp.tx.SetInt(bp.currentBlock, file.INT_32_BITS, n, true)
}

func (bp *BTreePage) insert(slot int32) error {
	numRecs, err := bp.GetNumRecs()
	if err != nil {
		return err
	}

	for i := numRecs; i > slot; i-- {
		if err := bp.copyRecords(i-1, i); err != nil {
			return err
		}
	}

	return bp.setNumRecs(numRecs + 1)
}

func (bp *BTreePage) copyRecords(from int32, to int32) error {
	sch := bp.layout.Schema()

	for _, fldName := range sch.Fields() {
		val, err := bp.getVal(from, fldName)
		if err != nil {
			return err
		}

		if err := bp.setVal(to, fldName, val); err != nil {
			return err
		}
	}
	return nil
}

func (bp *BTreePage) fldPos(slot int32, fldName string) int32 {
	offset := bp.layout.Offset(fldName)
	return bp.slotPos(slot) + offset
}

func (bp *BTreePage) slotPos(slot int32) int32 {
	slotSize := bp.layout.SlotSize()
	return file.INT_32_BITS + file.INT_32_BITS*(slot*slotSize)
}

func (bp *BTreePage) appendNew(flag int32) (*file.BlockId, error) {
	blk := bp.tx.Append(bp.currentBlock.FileName())

	if err := bp.tx.Pin(blk); err != nil {
		return &file.BlockId{}, err
	}

	if err := bp.Format(blk, flag); err != nil {
		return &file.BlockId{}, err
	}
	return blk, nil
}

func (bp *BTreePage) Format(block *file.BlockId, flag int32) error {
	if err := bp.tx.SetInt(block, 0, flag, true); err != nil {
		return err
	}

	if err := bp.tx.SetInt(block, file.INT_32_BITS, 0, false); err != nil {
		return err
	}

	for pos := int32(2 * file.INT_32_BITS); pos+bp.layout.SlotSize() < int32(bp.tx.BlockSize()); pos += bp.layout.SlotSize() {
		if err := bp.makeDefaultRecord(block, pos); err != nil {
			return err
		}
	}
	return nil
}

func (bp *BTreePage) makeDefaultRecord(blk *file.BlockId, pos int32) error {
	for _, fldName := range bp.layout.Schema().Fields() {
		offset := bp.layout.Offset(fldName)

		switch bp.layout.Schema().Type(fldName) {
		case record.INT:
			if err := bp.tx.SetInt(blk, pos+offset, 0, false); err != nil {
				return err
			}
		case record.VARCHAR:
			if err := bp.tx.SetString(blk, pos+offset, "", false); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected value type: %d", bp.layout.Schema().Type(fldName))
		}
	}
	return nil
}

func (bp *BTreePage) transferRecs(slot int32, dest *BTreePage) error {
	destSlot := int32(0)

	for {
		numRecs, err := bp.GetNumRecs()
		if err != nil {
			return err
		}

		if slot >= numRecs {
			break
		}

		if err := dest.insert(destSlot); err != nil {
			return err
		}

		for _, fldName := range dest.layout.Schema().Fields() {
			val, err := bp.getVal(slot, fldName)
			if err != nil {
				return err
			}

			if err := dest.setVal(destSlot, fldName, val); err != nil {
				return err
			}
		}

		if err := bp.Delete(slot); err != nil {
			return err
		}
		destSlot++
	}
	return nil
}

func (bp *BTreePage) Delete(slot int32) error {
	numRecs, err := bp.GetNumRecs()
	if err != nil {
		return err
	}

	for i := slot + 1; i < numRecs; i++ {
		if err := bp.copyRecords(i, i-1); err != nil {
			return err
		}
	}

	if err := bp.setNumRecs(numRecs - 1); err != nil {
		return err
	}
	return nil
}

// Method call only by BTreeDir
func (bp *BTreePage) GetChildNum(slot int32) (int32, error) {
	return bp.getInt(slot, "block")
}

func (bp *BTreePage) InsertDir(slot int32, val *record.Constant, blockNum int32) error {
	if err := bp.insert(slot); err != nil {
		return err
	}

	if err := bp.setVal(slot, "dataval", val); err != nil {
		return err
	}

	if err := bp.setInt(slot, "block", blockNum); err != nil {
		return err
	}
	return nil
}

// Method call only by BTreeLeaf
func (bp *BTreePage) GetDataRID(slot int32) (*record.RID, error) {
	blockNum, err := bp.getInt(slot, "block")
	if err != nil {
		return nil, err
	}
	id, err := bp.getInt(slot, "id")
	if err != nil {
		return nil, err
	}

	return record.NewRID(blockNum, id), nil
}

func (bp *BTreePage) InsertLeaf(slot int32, val *record.Constant, rid *record.RID) error {
	if err := bp.insert(slot); err != nil {
		return err
	}

	if err := bp.setVal(slot, "dataval", val); err != nil {
		return err
	}

	if err := bp.setInt(slot, "block", rid.BlockNumber()); err != nil {
		return err
	}

	if err := bp.setInt(slot, "id", rid.Slot()); err != nil {
		return err
	}
	return nil
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
