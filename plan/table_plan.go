package plan

import (
	"s1mpleasia.com/tinydb/metadata"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ Plan = (*TablePlan)(nil)

type TablePlan struct {
	tableName string
	tx        *transaction.Transaction
	layout    *record.Layout
	statInfo  *metadata.StatInfo
}

func NewTablePlan(tx *transaction.Transaction, tableName string, md *metadata.MetadataMgmt) (*TablePlan, error) {
	layout, err := md.GetLayout(tableName, tx)
	if err != nil {
		return nil, err
	}

	statInfo, err := md.GetStatInfo(tableName, layout, tx)
	if err != nil {
		return nil, err
	}

	return &TablePlan{
		tableName: tableName,
		tx:        tx,
		layout:    layout,
		statInfo:  statInfo,
	}, nil
}

func (p *TablePlan) Open() (query.Scan, error) {
	return query.NewTableScan(p.tx, p.tableName, p.layout)
}

func (p *TablePlan) BlockAccessed() int32 {
	return p.statInfo.BlockAccessed()
}

func (p *TablePlan) RecordsOutput() int32 {
	return p.statInfo.RecordsOutput()
}

func (p *TablePlan) DistinctValues(fieldName string) int32 {
	return p.statInfo.DistinctValues(fieldName)
}

func (p *TablePlan) Schema() *record.Schema {
	return p.layout.Schema()
}
