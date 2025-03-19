package metadata

import (
	"s1mpleasia.com/tinydb/index"
	"s1mpleasia.com/tinydb/index/btree"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

const (
	INDEX_INFO_FIELD_BLOCK         = "block"
	INDEX_INFO_FIELD_ID            = "id"
	INDEX_INFO_FIELD_DATA_VAL      = "dataval"
	INDEX_CATALOG_FILE             = "idxcat"
	INDEX_CATALOG_FIELD_INDEX_NAME = "indexname"
	INDEX_CATALOG_FIELD_TABLE_NAME = "tablename"
	INDEX_CATALOG_FIELD_FIELD_NAME = "fieldname"
)

type IndexInfo struct {
	indexName   string
	fieldName   string
	tx          *transaction.Transaction
	tableSchema *record.Schema
	indexLayout *record.Layout
	si          *StatInfo
}

func NewIndexInfo(indexName string, fieldName string,
	tableSChema *record.Schema, tx *transaction.Transaction, si *StatInfo) *IndexInfo {
	ii := &IndexInfo{indexName, fieldName, tx, tableSChema, nil, si}

	ii.indexLayout = ii.createIdxLayout()
	return ii
}

func (ii *IndexInfo) Open() index.Index {
	treeIndex, err := btree.NewBTreeIndex(ii.tx, ii.indexName, ii.indexLayout)
	if err != nil {
		panic(err)
	}
	return treeIndex
	//return index.NewHashIndex(ii.tx, ii.indexName, ii.indexLayout)
}

func (ii *IndexInfo) BlockAccessed() int32 {
	rpb := int32(ii.tx.BlockSize()) / ii.indexLayout.SlotSize()
	numBlocks := ii.si.RecordsOutput() / rpb

	// TODO: Hash/BTree index search cost
	return numBlocks
}

func (ii *IndexInfo) RecordsOutput() int32 {
	return ii.si.RecordsOutput() / ii.si.DistinctValues(ii.fieldName)
}

func (ii *IndexInfo) DistinctValues(fieldName string) int32 {
	if ii.fieldName == fieldName {
		return 1
	} else {
		return ii.si.DistinctValues(fieldName)
	}
}

/*
Index layout:
+------------+-----------+--------------*-----------------*-------------+
| block (4B) |  id (4B)  |	dataval (4B - INT || 4B + 2*len - VARCHAR)	|
+------------+-----------+--------------*-----------------*-------------+
*/
func (ii *IndexInfo) createIdxLayout() *record.Layout {
	sch := record.NewSchema()
	sch.AddIntField(INDEX_INFO_FIELD_BLOCK)
	sch.AddIntField(INDEX_INFO_FIELD_ID)

	if ii.tableSchema.Type(ii.fieldName) == record.INT {
		sch.AddIntField(INDEX_INFO_FIELD_DATA_VAL)
	} else {
		fieldLen := ii.tableSchema.Length(ii.fieldName)
		sch.AddStringField(INDEX_INFO_FIELD_DATA_VAL, fieldLen)
	}

	return record.NewLayoutFromSchema(sch)
}

type IndexMgmt struct {
	layout   *record.Layout
	tblMgmt  *TableMgmt
	statMgmt *StatMgmt
}

func NewIndexMgmt(isNew bool, tblMgmt *TableMgmt, statMgmt *StatMgmt, tx *transaction.Transaction) (*IndexMgmt, error) {
	if isNew {
		sch := record.NewSchema()
		sch.AddStringField(INDEX_CATALOG_FIELD_INDEX_NAME, MAX_NAME)
		sch.AddStringField(INDEX_CATALOG_FIELD_TABLE_NAME, MAX_NAME)
		sch.AddStringField(INDEX_CATALOG_FIELD_FIELD_NAME, MAX_NAME)

		err := tblMgmt.CreateTable(INDEX_CATALOG_FILE, sch, tx)
		if err != nil {
			return nil, err
		}
	}

	layout, err := tblMgmt.GetLayout(INDEX_CATALOG_FILE, tx)
	if err != nil {
		return nil, err
	}

	return &IndexMgmt{
		layout:   layout,
		tblMgmt:  tblMgmt,
		statMgmt: statMgmt,
	}, nil
}

func (im *IndexMgmt) CreateIndex(idxName string, tableName string, fieldName string, tx *transaction.Transaction) error {
	ts, err := query.NewTableScan(tx, INDEX_CATALOG_FILE, im.layout)
	if err != nil {
		return err
	}

	defer ts.Close()
	if err = ts.Insert(); err != nil {
		return err
	}

	if err = ts.SetString(INDEX_CATALOG_FIELD_INDEX_NAME, idxName); err != nil {
		return err
	}
	if err = ts.SetString(INDEX_CATALOG_FIELD_TABLE_NAME, tableName); err != nil {
		return err
	}
	if err = ts.SetString(INDEX_CATALOG_FIELD_FIELD_NAME, fieldName); err != nil {
		return err
	}

	return nil
}

func (im *IndexMgmt) GetIndexInfo(tableName string, tx *transaction.Transaction) (map[string]*IndexInfo, error) {
	result := make(map[string]*IndexInfo)

	ts, err := query.NewTableScan(tx, INDEX_CATALOG_FILE, im.layout)
	if err != nil {
		return nil, err
	}

	defer ts.Close()

	for ts.Next() {
		tblName, err := ts.GetString(INDEX_CATALOG_FIELD_TABLE_NAME)
		if err != nil {
			return nil, err
		}

		if tblName == tableName {
			idxName, err := ts.GetString(INDEX_CATALOG_FIELD_INDEX_NAME)
			if err != nil {
				return nil, err
			}

			fieldName, err := ts.GetString(INDEX_CATALOG_FIELD_FIELD_NAME)
			if err != nil {
				return nil, err
			}

			tblLayout, err := im.tblMgmt.GetLayout(tableName, tx)
			if err != nil {
				return nil, err
			}

			tableStatInfo, err := im.statMgmt.GetStatInfo(tableName, tblLayout, tx)
			if err != nil {
				return nil, err
			}

			ii := NewIndexInfo(idxName, fieldName, tblLayout.Schema(), tx, tableStatInfo)
			result[fieldName] = ii
		}
	}

	return result, nil
}
