package record

import (
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/transaction"
)

type RecordFlag int32

const (
	EMPTY RecordFlag = iota
	USED
)

/*
  - A record page format contains slot for storing records
    +---------+---------+---------+---------+--+
    |  Slot 0 |  Slot 1 |  .....  |	 Slot n |--|
    +---------+---------+---------+---------+--+
*/
type RecordPage struct {
	transaction *transaction.Transaction
	block       *file.BlockId
	layout      *Layout
}

func NewRecordPage(tx *transaction.Transaction, block *file.BlockId, layout *Layout) (*RecordPage, error) {
	if err := tx.Pin(block); err != nil {
		return nil, err
	}

	return &RecordPage{
		transaction: tx,
		block:       block,
		layout:      layout,
	}, nil
}

func (recordPage *RecordPage) GetInt(slot int32, fieldName string) (int32, error) {
	fieldPos := recordPage.offset(slot) + recordPage.layout.Offset(fieldName)
	return recordPage.transaction.GetInt(recordPage.block, int(fieldPos))
}

func (recordPage *RecordPage) GetString(slot int32, fieldName string) (string, error) {
	fieldPos := recordPage.offset(slot) + recordPage.layout.Offset(fieldName)

	return recordPage.transaction.GetString(recordPage.block, int(fieldPos))
}

func (recordPage *RecordPage) SetInt(slot int32, fieldName string, val int32) error {
	fieldPos := recordPage.offset(slot) + recordPage.layout.Offset(fieldName)

	return recordPage.transaction.SetInt(recordPage.block, fieldPos, val, true)
}

func (recordPage *RecordPage) SetString(slot int32, fieldName string, val string) error {
	fieldPos := recordPage.offset(slot) + recordPage.layout.Offset(fieldName)

	return recordPage.transaction.SetString(recordPage.block, fieldPos, val, true)
}

func (recordPage *RecordPage) Delete(slot int32) error {
	return recordPage.setFlag(slot, EMPTY)
}

func (recordPage *RecordPage) Format() error {
	var slot int32 = 0

	for recordPage.isValidSlot(slot) {
		err := recordPage.transaction.SetInt(recordPage.block, recordPage.offset(slot), int32(EMPTY), false)
		if err != nil {
			return err
		}

		schema := recordPage.layout.schema

		for _, fieldName := range schema.fields {
			fieldPos := recordPage.offset(slot) + recordPage.layout.Offset(fieldName)
			if schema.Type(fieldName) == INT {
				err = recordPage.transaction.SetInt(recordPage.block, fieldPos, 0, false)
			} else {
				err = recordPage.transaction.SetString(recordPage.block, fieldPos, "", false)
			}

			if err != nil {
				return err
			}
		}

		slot++
	}
	return nil
}

func (recordPage *RecordPage) NextAfter(slot int32) (int32, error) {
	return recordPage.searchAfter(slot, USED)
}

func (recordPage *RecordPage) InsertAfter(slot int32) (int32, error) {
	newSlot, err := recordPage.searchAfter(slot, EMPTY)
	if err != nil {
		return 0, err
	}
	if newSlot >= 0 {
		if err := recordPage.setFlag(newSlot, USED); err != nil {
			return 0, err
		}
	}

	return newSlot, nil
}

func (recordPage *RecordPage) Block() *file.BlockId {
	return recordPage.block
}

func (recordPage *RecordPage) setFlag(slot int32, flag RecordFlag) error {
	return recordPage.transaction.SetInt(recordPage.block, recordPage.offset(slot), int32(flag), true)
}

func (recordPage *RecordPage) searchAfter(slot int32, flag RecordFlag) (int32, error) {
	slot++

	for recordPage.isValidSlot(slot) {
		slotFlag, err := recordPage.transaction.GetInt(recordPage.block, int(recordPage.offset(slot)))
		if err != nil {
			return 0, err
		}

		if RecordFlag(slotFlag) == flag {
			return slot, nil
		}

		slot++
	}

	return -1, nil
}

func (recordPage *RecordPage) isValidSlot(slot int32) bool {
	return recordPage.offset(slot+1) <= int32(recordPage.transaction.BlockSize())
}

func (recordPage *RecordPage) offset(slot int32) int32 {
	return slot * recordPage.layout.slotSize
}
