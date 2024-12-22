package record

import (
	"errors"

	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/transaction"
)

var ErrUnknownFieldType = errors.New("unknown field type")

type TableScan struct {
	tx          *transaction.Transaction
	layout      *Layout
	recordPage  *RecordPage
	fileName    string
	currentSlot int32
}

func NewTableScan(tx *transaction.Transaction, tableName string, layout *Layout) (*TableScan, error) {
	fileName := tableName + ".tbl"
	ts := &TableScan{
		tx:       tx,
		fileName: fileName,
		layout:   layout,
	}

	if ts.tx.Size(fileName) == 0 {
		err := ts.moveToNewBlock()
		if err != nil {
			return nil, err
		}
	} else {
		err := ts.moveToBlock(0)
		if err != nil {
			return nil, err
		}
	}

	return ts, nil
}

func (ts *TableScan) Close() {
	if ts.recordPage != nil {
		ts.tx.Unpin(ts.recordPage.block)
	}
}

func (ts *TableScan) BeforeFirst() error {
	if err := ts.moveToBlock(0); err != nil {
		return err
	}

	return nil
}

// This function finds the next slot in the file that contains content.
// It will return false if it reaches at EOF without finding any slot
func (ts *TableScan) Next() bool {
	nextSlot, err := ts.recordPage.NextAfter(ts.currentSlot)
	if err != nil {
		return false
	}
	ts.currentSlot = nextSlot
	for ts.currentSlot < 0 {
		if ts.atLastBlock() {
			return false
		}

		err = ts.moveToBlock(ts.recordPage.block.BlockNumber() + 1)
		if err != nil {
			return false
		}

		ts.currentSlot, err = ts.recordPage.NextAfter(ts.currentSlot)
		if err != nil {
			return false
		}
	}

	return true
}

func (ts *TableScan) GetInt(fieldName string) (int32, error) {
	return ts.recordPage.GetInt(ts.currentSlot, fieldName)
}

func (ts *TableScan) GetString(fieldName string) (string, error) {
	return ts.recordPage.GetString(ts.currentSlot, fieldName)
}

func (ts *TableScan) GetVal(fieldName string) (*Constant, error) {
	switch ts.layout.schema.Type(fieldName) {
	case INT:
		val, err := ts.GetInt(fieldName)
		if err != nil {
			return nil, err
		}

		return NewConstantWithInt(val), nil
	case VARCHAR:
		val, err := ts.GetString(fieldName)
		if err != nil {
			return nil, err
		}

		return NewConstantWithString(val), nil
	default:
		return nil, ErrUnknownFieldType
	}
}

func (ts *TableScan) HasField(fieldName string) bool {
	return ts.layout.schema.HasField(fieldName)
}

func (ts *TableScan) SetInt(fieldName string, val int32) error {
	return ts.recordPage.SetInt(ts.currentSlot, fieldName, val)
}

func (ts *TableScan) SetString(fieldName string, val string) error {
	return ts.recordPage.SetString(ts.currentSlot, fieldName, val)
}

func (ts *TableScan) SetVal(fieldName string, val *Constant) error {
	switch ts.layout.schema.Type(fieldName) {
	case INT:
		ival, err := val.AsInt()
		if err != nil {
			return err
		}

		if err = ts.SetInt(fieldName, ival); err != nil {
			return err
		}
	case VARCHAR:
		sval, err := val.AsString()
		if err != nil {
			return err
		}

		if err = ts.SetString(fieldName, sval); err != nil {
			return err
		}

	default:
		return ErrUnknownFieldType
	}

	return nil
}

// It finds the next slot with empty flag.
// If it's EOF and still not found any slot -> Append new block to the file
func (ts *TableScan) Insert() error {
	nextSlot, err := ts.recordPage.InsertAfter(ts.currentSlot)

	if err != nil {
		return err
	}

	ts.currentSlot = nextSlot
	for ts.currentSlot < 0 {
		if ts.atLastBlock() {
			err = ts.moveToNewBlock()
		} else {
			err = ts.moveToBlock(ts.recordPage.block.BlockNumber() + 1)
		}

		if err != nil {
			return err
		}

		nextSlot, err := ts.recordPage.InsertAfter(ts.currentSlot)
		if err != nil {
			return err
		}

		ts.currentSlot = nextSlot
	}

	return nil
}

func (ts *TableScan) Delete() error {
	return ts.recordPage.Delete(ts.currentSlot)
}

func (ts *TableScan) MoveToRID(rid *RID) {
	ts.Close()
	block := file.NewBlockId(ts.fileName, int64(rid.BlockNumber()))
	rp, err := NewRecordPage(ts.tx, block, ts.layout)
	if err != nil {
		panic(err)
	}
	ts.recordPage = rp
	ts.currentSlot = rid.Slot()
}

func (ts *TableScan) GetRID() *RID {
	return NewRID(int32(ts.recordPage.Block().BlockNumber()), ts.currentSlot)
}

func (ts *TableScan) moveToBlock(blockNum int64) error {
	ts.Close()
	block := file.NewBlockId(ts.fileName, blockNum)
	rp, err := NewRecordPage(ts.tx, block, ts.layout)
	if err != nil {
		return err
	}

	ts.recordPage = rp
	ts.currentSlot = -1
	return nil
}

func (ts *TableScan) moveToNewBlock() error {
	ts.Close()
	block := ts.tx.Append(ts.fileName)
	rp, err := NewRecordPage(ts.tx, block, ts.layout)
	if err != nil {
		return err
	}

	ts.recordPage = rp
	ts.recordPage.Format()
	ts.currentSlot = -1
	return nil
}

func (ts *TableScan) atLastBlock() bool {
	return ts.recordPage.Block().BlockNumber() == int64(ts.tx.Size(ts.fileName)-1)
}
