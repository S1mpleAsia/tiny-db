package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ Scan = (*ChunkScan)(nil)

type ChunkScan struct {
	buffs                                 []*record.RecordPage
	tx                                    *transaction.Transaction
	fileName                              string
	layout                                *record.Layout
	startBlkNum, endBlkNum, currentBlkNum int64
	rp                                    *record.RecordPage
	currentSlot                           int32
}

func NewChunkScan(tx *transaction.Transaction, fileName string, layout *record.Layout, startBlkNum, endBlkNum int64) (*ChunkScan, error) {
	buffs := make([]*record.RecordPage, 0, endBlkNum-startBlkNum+1)
	for i := startBlkNum; i <= endBlkNum; i++ {
		blk := file.NewBlockId(fileName, i)
		rp, err := record.NewRecordPage(tx, blk, layout)
		if err != nil {
			return nil, fmt.Errorf("record.NewRecordPage: %w", err)
		}
		buffs = append(buffs, rp)
	}

	s := &ChunkScan{
		buffs:       buffs,
		tx:          tx,
		fileName:    fileName,
		layout:      layout,
		startBlkNum: startBlkNum,
		endBlkNum:   endBlkNum,
	}
	s.moveToBlock(startBlkNum)

	return s, nil
}

func (cs *ChunkScan) BeforeFirst() error {
	cs.moveToBlock(cs.startBlkNum)
	return nil
}

func (cs *ChunkScan) Next() bool {
	var err error
	cs.currentSlot, err = cs.rp.NextAfter(cs.currentSlot)
	if err != nil {
		return false
	}

	for cs.currentSlot < 0 {
		if cs.currentBlkNum == cs.endBlkNum {
			return false
		}

		cs.moveToBlock(cs.currentBlkNum + 1) // Reset the slot to -1 of the next block
		cs.currentSlot, err = cs.rp.NextAfter(cs.currentSlot)
		if err != nil {
			return false
		}
	}

	return true
}

func (cs *ChunkScan) GetInt(fieldName string) (int32, error) {
	return cs.rp.GetInt(cs.currentSlot, fieldName)
}

func (cs *ChunkScan) GetString(fieldName string) (string, error) {
	return cs.rp.GetString(cs.currentSlot, fieldName)
}

func (cs *ChunkScan) GetVal(fieldName string) (*record.Constant, error) {
	if cs.layout.Schema().Type(fieldName) == record.INT {
		i, err := cs.GetInt(fieldName)
		if err != nil {
			return nil, fmt.Errorf("GetInt(%s): %w", fieldName, err)
		}
		return record.NewConstantWithInt(i), nil
	} else if cs.layout.Schema().Type(fieldName) == record.VARCHAR {
		s, err := cs.GetString(fieldName)
		if err != nil {
			return nil, fmt.Errorf("GetString(%s): %w", fieldName, err)
		}
		return record.NewConstantWithString(s), nil
	}

	return nil, fmt.Errorf("unknown field type: %s", fieldName)
}

func (cs *ChunkScan) HasField(fieldName string) bool {
	return cs.layout.Schema().HasField(fieldName)
}

func (cs *ChunkScan) Close() {
	for i := range cs.buffs {
		cs.tx.Unpin(file.NewBlockId(cs.fileName, cs.startBlkNum+int64(i)))
	}
}

func (cs *ChunkScan) moveToBlock(blkNum int64) {
	cs.currentBlkNum = blkNum
	cs.rp = cs.buffs[cs.currentBlkNum-cs.startBlkNum]
	cs.currentSlot = -1
}
