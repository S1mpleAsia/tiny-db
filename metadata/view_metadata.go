package metadata

import (
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

const MAX_VIEWDEF = 100
const (
	VIEW_FILE = "viewcat"
	VIEW_NAME = "viewname"
	VIEW_DEF  = "viewdef"
)

type ViewMgmt struct {
	tableMgmt *TableMgmt
}

func NewViewMgmt(isNew bool, tblMgmt *TableMgmt, tx *transaction.Transaction) (*ViewMgmt, error) {
	vm := &ViewMgmt{
		tableMgmt: tblMgmt,
	}

	if isNew {
		sch := record.NewSchema()
		sch.AddStringField(VIEW_NAME, MAX_NAME)
		sch.AddStringField(VIEW_DEF, MAX_VIEWDEF)
		err := vm.tableMgmt.CreateTable(VIEW_FILE, sch, tx)
		if err != nil {
			return nil, err
		}
	}

	return vm, nil
}

func (vm *ViewMgmt) CreateView(viewName string, viewDef string, tx *transaction.Transaction) error {
	layout, err := vm.tableMgmt.GetLayout(VIEW_FILE, tx)
	if err != nil {
		return err
	}

	ts, err := record.NewTableScan(tx, VIEW_FILE, layout)
	if err != nil {
		return err
	}
	defer ts.Close()

	if err := ts.Insert(); err != nil {
		return err
	}

	if err = ts.SetString(VIEW_NAME, viewName); err != nil {
		return err
	}
	if err = ts.SetString(VIEW_DEF, viewDef); err != nil {
		return err
	}

	return nil
}

func (vm *ViewMgmt) GetViewDef(viewName string, tx *transaction.Transaction) (string, error) {
	var result string = ""
	layout, err := vm.tableMgmt.GetLayout(VIEW_FILE, tx)
	if err != nil {
		return "", err
	}

	ts, err := record.NewTableScan(tx, VIEW_FILE, layout)
	if err != nil {
		return "", err
	}

	for ts.Next() {
		vname, err := ts.GetString(VIEW_NAME)
		if err != nil {
			return "", err
		}

		if vname == viewName {
			result, err = ts.GetString(VIEW_DEF)
			if err != nil {
				return "", err
			}

			break
		}
	}

	ts.Close()
	return result, nil
}
