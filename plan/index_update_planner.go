package plan

import (
	"errors"
	"s1mpleasia.com/tinydb/metadata"
	"s1mpleasia.com/tinydb/parse"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/transaction"
)

var _ UpdatePlanner = (*IndexUpdatePlanner)(nil)

type IndexUpdatePlanner struct {
	mdm *metadata.MetadataMgmt
}

func NewIndexUpdatePlanner(mdm *metadata.MetadataMgmt) *IndexUpdatePlanner {
	return &IndexUpdatePlanner{mdm}
}

func (up *IndexUpdatePlanner) ExecuteInsert(insertData *parse.InsertData, tx *transaction.Transaction) (int, error) {
	tablePlan, err := NewTablePlan(tx, insertData.TableName(), up.mdm)
	if err != nil {
		return 0, err
	}

	scan, err := tablePlan.Open()
	if err != nil {
		return 0, err
	}

	defer scan.Close()

	updateScan, ok := scan.(query.UpdateScan)
	if !ok {
		return 0, errors.New("ExecuteInsert: plan is not a table plan")
	}

	if err = updateScan.Insert(); err != nil {
		return 0, err
	}

	rid := updateScan.GetRID()

	indexInfo, err := up.mdm.GetIndexInfo(insertData.TableName(), tx)
	if err != nil {
		return 0, err
	}

	for i, field := range insertData.Fields() {
		val := insertData.Values()[i]
		if err := updateScan.SetVal(field, val); err != nil {
			return 0, err
		}

		ii, ok := indexInfo[field]
		if !ok {
			continue
		}

		idx := ii.Open()
		if err := idx.Insert(val, rid); err != nil {
			return 0, err
		}

		if err = idx.Close(); err != nil {
			return 0, err
		}
	}

	return 1, nil
}

func (up *IndexUpdatePlanner) ExecuteDelete(deleteData *parse.DeleteData, tx *transaction.Transaction) (int, error) {
	tableName := deleteData.TableName()
	tablePlan, err := NewTablePlan(tx, tableName, up.mdm)
	if err != nil {
		return 0, err
	}

	selectPlan, err := NewSelectPlan(tablePlan, deleteData.Pred())
	if err != nil {
		return 0, err
	}

	indexInfo, err := up.mdm.GetIndexInfo(tableName, tx)
	if err != nil {
		return 0, err
	}

	scan, err := selectPlan.Open()
	if err != nil {
		return 0, err
	}

	defer scan.Close()

	updateScan, ok := scan.(query.UpdateScan)
	if !ok {
		return 0, errors.New("ExecuteDelete: plan is not a table plan")
	}

	count := 0
	for {
		if !scan.Next() {
			break
		}

		rid := updateScan.GetRID()

		for fieldName, ii := range indexInfo {
			val, err := scan.GetVal(fieldName)
			if err != nil {
				return 0, err
			}

			idx := ii.Open()
			if err = idx.Delete(val, rid); err != nil {
				return 0, err
			}

			if err = idx.Close(); err != nil {
				return 0, err
			}
		}

		if err := updateScan.Delete(); err != nil {
			return 0, err
		}

		count++
	}

	return count, nil
}

func (up *IndexUpdatePlanner) ExecuteModify(modifyData *parse.ModifyData, tx *transaction.Transaction) (int, error) {
	tablePlan, err := NewTablePlan(tx, modifyData.TableName(), up.mdm)
	if err != nil {
		return 0, err
	}

	selectPlan, err := NewSelectPlan(tablePlan, modifyData.Pred())
	if err != nil {
		return 0, err
	}

	indexInfo, err := up.mdm.GetIndexInfo(modifyData.TableName(), tx)
	if err != nil {
		return 0, err
	}

	var idx query.Index
	if ii, ok := indexInfo[modifyData.TargetField()]; ok {
		idx = ii.Open()
	}

	scan, err := selectPlan.Open()
	if err != nil {
		return 0, err
	}

	defer scan.Close()

	updateScan, ok := scan.(query.UpdateScan)
	if !ok {
		return 0, errors.New("ExecuteModify: plan is not a table plan")
	}

	count := 0
	for {
		if !scan.Next() {
			break
		}
		newVal, err := modifyData.NewValue().Evaluate(scan)
		if err != nil {
			return 0, err
		}

		oldVal, err := scan.GetVal(modifyData.TargetField())
		if err != nil {
			return 0, err
		}

		if err = updateScan.SetVal(modifyData.TargetField(), newVal); err != nil {
			return 0, err
		}

		count++

		if idx == nil {
			continue
		}

		rid := updateScan.GetRID()
		if err = idx.Delete(oldVal, rid); err != nil {
			return 0, err
		}

		if err = idx.Insert(newVal, rid); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (up *IndexUpdatePlanner) ExecuteCreateTable(createTableData *parse.CreateTableData, tx *transaction.Transaction) (int, error) {
	tableName := createTableData.TableName()
	sch := createTableData.NewSchema()

	err := up.mdm.CreateTable(tableName, sch, tx)
	return 0, err
}

func (up *IndexUpdatePlanner) ExecuteCreateView(createViewData *parse.CreateViewData, tx *transaction.Transaction) (int, error) {
	viewName := createViewData.ViewName()
	viewDef := createViewData.ViewDef()

	err := up.mdm.CreateView(viewName, viewDef, tx)
	return 0, err
}

func (up *IndexUpdatePlanner) ExecuteCreateIndex(createIndexData *parse.CreateIndexData, tx *transaction.Transaction) (int, error) {
	indexName := createIndexData.IndexName()
	tableName := createIndexData.TableName()
	fieldName := createIndexData.FieldName()

	err := up.mdm.CreateIndex(indexName, tableName, fieldName, tx)
	return 0, err
}
