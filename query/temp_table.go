package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
	"sync"
)

var nextTableNum = 0
var mux = &sync.Mutex{}

type TempTable struct {
	tx      *transaction.Transaction
	tblName string
	layout  *record.Layout
}

func NewTempTable(tx *transaction.Transaction, sch *record.Schema) *TempTable {
	layout := record.NewLayoutFromSchema(sch)

	return &TempTable{
		tx:      tx,
		tblName: nextTableName(),
		layout:  layout,
	}
}

func (tt *TempTable) Open() (*TableScan, error) {
	scan, err := NewTableScan(tt.tx, tt.tblName, tt.layout)
	if err != nil {
		return nil, fmt.Errorf("open temp table %s: %w", tt.tblName, err)
	}

	return scan, nil
}

func (tt *TempTable) TableName() string {
	return tt.tblName
}

func (tt *TempTable) Layout() *record.Layout {
	return tt.layout
}

func nextTableName() string {
	mux.Lock()
	defer mux.Unlock()

	nextTableNum++
	return fmt.Sprintf("temp%d", nextTableNum)
}
