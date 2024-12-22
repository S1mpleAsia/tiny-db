package metadata

import (
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

type MetadataMgmt struct {
	tblMgmt   *TableMgmt
	viewMgmt  *ViewMgmt
	statMgmt  *StatMgmt
	indexMgmt *IndexMgmt
}

func NewMetadataMgmt(isNew bool, tx *transaction.Transaction) (*MetadataMgmt, error) {
	tblMgmt, err := NewTableMgmt(isNew, tx)
	if err != nil {
		return nil, err
	}

	viewMgmt, err := NewViewMgmt(isNew, tblMgmt, tx)
	if err != nil {
		return nil, err
	}

	statMgmt, err := NewStatMgmt(tblMgmt, tx)
	if err != nil {
		return nil, err
	}

	idxMgmt, err := NewIndexMgmt(isNew, tblMgmt, statMgmt, tx)
	if err != nil {
		return nil, err
	}

	return &MetadataMgmt{
		tblMgmt:   tblMgmt,
		viewMgmt:  viewMgmt,
		statMgmt:  statMgmt,
		indexMgmt: idxMgmt,
	}, nil
}

func (mm *MetadataMgmt) CreateTable(tableName string, sch *record.Schema, tx *transaction.Transaction) error {
	return mm.tblMgmt.CreateTable(tableName, sch, tx)
}

func (mm *MetadataMgmt) GetLayout(tableName string, tx *transaction.Transaction) (*record.Layout, error) {
	return mm.tblMgmt.GetLayout(tableName, tx)
}

func (mm *MetadataMgmt) CreateView(viewName string, viewDef string, tx *transaction.Transaction) error {
	return mm.viewMgmt.CreateView(viewName, viewDef, tx)
}

func (mm *MetadataMgmt) GetViewDef(viewName string, tx *transaction.Transaction) (string, error) {
	return mm.viewMgmt.GetViewDef(viewName, tx)
}

func (mm *MetadataMgmt) CreateIndex(idxName string, tableName string, fieldName string, tx *transaction.Transaction) error {
	return mm.indexMgmt.CreateIndex(idxName, tableName, fieldName, tx)
}

func (mm *MetadataMgmt) GetIndexInfo(tableName string, tx *transaction.Transaction) (map[string]*IndexInfo, error) {
	return mm.indexMgmt.GetIndexInfo(tableName, tx)
}

func (mm *MetadataMgmt) GetStatInfo(tableName string, layout *record.Layout, tx *transaction.Transaction) (*StatInfo, error) {
	return mm.statMgmt.GetStatInfo(tableName, layout, tx)
}

func (mm *MetadataMgmt) ForceRefreshStatistics(tx *transaction.Transaction) error {
	return mm.statMgmt.ForceRefreshStatistics(tx)
}
