package metadata

import (
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

const MAX_NAME = 16

const (
	TBL_CATALOG_FILE       = "tblcat"
	TBL_CATALOG_TABLE_NAME = "tblname"
	TBL_CATALOG_SLOT_SIZE  = "slotsize"

	FIELD_CATALOG_FILE       = "fldcat"
	FIELD_CATALOG_TABLE_NAME = "tblname"
	FIELD_CATALOG_FIELD_NAME = "fldname"
	FIELD_CATALOG_TYPE       = "type"
	FIELD_CATALOG_LENGTH     = "length"
	FIELD_CATALOG_OFFSET     = "offset"
)

// Manage metadata about table catalog and field catalog
type TableMgmt struct {
	tblCatalogLayout *record.Layout
	fldCatalogLayout *record.Layout
}

func NewTableMgmt(isNew bool, tx *transaction.Transaction) (*TableMgmt, error) {
	tblCatalogSchema := record.NewSchema()
	// Set schema for table catalog
	tblCatalogSchema.AddStringField(TBL_CATALOG_TABLE_NAME, MAX_NAME)
	tblCatalogSchema.AddIntField(TBL_CATALOG_SLOT_SIZE)
	tblCatalogLayout := record.NewLayoutFromSchema(tblCatalogSchema)

	fldCatalogSchema := record.NewSchema()
	// Set schema for field catalog
	fldCatalogSchema.AddStringField(FIELD_CATALOG_TABLE_NAME, MAX_NAME)
	fldCatalogSchema.AddStringField(FIELD_CATALOG_FIELD_NAME, MAX_NAME)
	fldCatalogSchema.AddIntField(FIELD_CATALOG_TYPE)
	fldCatalogSchema.AddIntField(FIELD_CATALOG_LENGTH)
	fldCatalogSchema.AddIntField(FIELD_CATALOG_OFFSET)
	fldCatalogLayout := record.NewLayoutFromSchema(fldCatalogSchema)

	tblMgmt := &TableMgmt{
		tblCatalogLayout: tblCatalogLayout,
		fldCatalogLayout: fldCatalogLayout,
	}

	if isNew {
		if err := tblMgmt.CreateTable(TBL_CATALOG_FILE, tblCatalogSchema, tx); err != nil {
			return nil, err
		}

		if err := tblMgmt.CreateTable(FIELD_CATALOG_FILE, fldCatalogSchema, tx); err != nil {
			return nil, err
		}
	}

	return tblMgmt, nil
}

func (tm *TableMgmt) CreateTable(tblName string, sch *record.Schema, tx *transaction.Transaction) error {
	layout := record.NewLayoutFromSchema(sch)

	tblCat, err := record.NewTableScan(tx, TBL_CATALOG_FILE, tm.tblCatalogLayout)
	if err != nil {
		return err
	}

	defer tblCat.Close()
	if err = tblCat.Insert(); err != nil {
		return err
	}

	if err = tblCat.SetString(TBL_CATALOG_TABLE_NAME, tblName); err != nil {
		return err
	}

	if err = tblCat.SetInt(TBL_CATALOG_SLOT_SIZE, layout.SlotSize()); err != nil {
		return err
	}

	fldCat, err := record.NewTableScan(tx, FIELD_CATALOG_FILE, tm.fldCatalogLayout)
	if err != nil {
		return err
	}

	defer fldCat.Close()

	for _, fieldName := range sch.Fields() {
		if err = fldCat.Insert(); err != nil {
			return err
		}

		if err = fldCat.SetString(FIELD_CATALOG_TABLE_NAME, tblName); err != nil {
			return err
		}
		if err = fldCat.SetString(FIELD_CATALOG_FIELD_NAME, fieldName); err != nil {
			return err
		}
		if err = fldCat.SetInt(FIELD_CATALOG_TYPE, int32(sch.Type(fieldName))); err != nil {
			return err
		}
		if err = fldCat.SetInt(FIELD_CATALOG_LENGTH, int32(sch.Length(fieldName))); err != nil {
			return err
		}
		if err = fldCat.SetInt(FIELD_CATALOG_OFFSET, layout.Offset(fieldName)); err != nil {
			return err
		}
	}

	return nil
}

func (tm *TableMgmt) GetLayout(tblName string, tx *transaction.Transaction) (*record.Layout, error) {
	var size int32 = -1
	tblCat, err := record.NewTableScan(tx, TBL_CATALOG_FILE, tm.tblCatalogLayout)
	if err != nil {
		return nil, err
	}

	for tblCat.Next() {
		t, err := tblCat.GetString(TBL_CATALOG_TABLE_NAME)
		if err != nil {
			return nil, err
		}

		if t == tblName {
			size, err = tblCat.GetInt(TBL_CATALOG_SLOT_SIZE)
			if err != nil {
				return nil, err
			}

			break
		}
	}

	tblCat.Close()
	sch := record.NewSchema()
	offsets := make(map[string]int32)

	fldCat, err := record.NewTableScan(tx, FIELD_CATALOG_FILE, tm.fldCatalogLayout)
	if err != nil {
		return nil, err
	}

	for fldCat.Next() {
		f, err := fldCat.GetString(FIELD_CATALOG_TABLE_NAME)
		if err != nil {
			return nil, err
		}

		if f == tblName {
			fldName, err := fldCat.GetString(FIELD_CATALOG_FIELD_NAME)
			if err != nil {
				return nil, err
			}

			fldType, err := fldCat.GetInt(FIELD_CATALOG_TYPE)
			if err != nil {
				return nil, err
			}

			fldLen, err := fldCat.GetInt(FIELD_CATALOG_LENGTH)
			if err != nil {
				return nil, err
			}

			fldOffset, err := fldCat.GetInt(FIELD_CATALOG_OFFSET)
			if err != nil {
				return nil, err
			}

			offsets[fldName] = fldOffset
			sch.AddField(fldName, record.FieldType(fldType), fldLen)
		}
	}

	return record.NewLayout(sch, offsets, size), nil
}
