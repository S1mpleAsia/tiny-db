package plan

import (
	"errors"

	"s1mpleasia.com/tinydb/metadata"
	"s1mpleasia.com/tinydb/parse"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/transaction"
)

var _ UpdatePlanner = (*BasicUpdatePlanner)(nil)

type BasicUpdatePlanner struct {
	mdm *metadata.MetadataMgmt
}

func NewBasicUpdatePlanner(mdm *metadata.MetadataMgmt) *BasicUpdatePlanner {
	return &BasicUpdatePlanner{mdm}
}

func (up *BasicUpdatePlanner) ExecuteInsert(insertData *parse.InsertData, tx *transaction.Transaction) (int, error) {
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

	for i, field := range insertData.Fields() {
		val := insertData.Values()[i]
		if err := updateScan.SetVal(field, val); err != nil {
			return 0, err
		}
	}

	return 1, nil
}

func (up *BasicUpdatePlanner) ExecuteDelete(deleteData *parse.DeleteData, tx *transaction.Transaction) (int, error) {
	tableName := deleteData.TableName()
	tablePlan, err := NewTablePlan(tx, tableName, up.mdm)
	if err != nil {
		return 0, err
	}

	selectPlan, err := NewSelectPlan(tablePlan, deleteData.Pred())
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

		if err := updateScan.Delete(); err != nil {
			return 0, err
		}

		count++
	}

	return count, nil
}

func (up *BasicUpdatePlanner) ExecuteModify(modifyData *parse.ModifyData, tx *transaction.Transaction) (int, error) {
	tablePlan, err := NewTablePlan(tx, modifyData.TableName(), up.mdm)
	if err != nil {
		return 0, err
	}

	selectPlan, err := NewSelectPlan(tablePlan, modifyData.Pred())
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

		if err = updateScan.SetVal(modifyData.TargetField(), newVal); err != nil {
			return 0, err
		}

		count++
	}

	return count, nil
}

func (up *BasicUpdatePlanner) ExecuteCreateTable(createTableData *parse.CreateTableData, tx *transaction.Transaction) (int, error) {
	tableName := createTableData.TableName()
	sch := createTableData.NewSchema()

	err := up.mdm.CreateTable(tableName, sch, tx)
	return 0, err
}

func (up *BasicUpdatePlanner) ExecuteCreateView(createViewData *parse.CreateViewData, tx *transaction.Transaction) (int, error) {
	viewName := createViewData.ViewName()
	viewDef := createViewData.ViewDef()

	err := up.mdm.CreateView(viewName, viewDef, tx)
	return 0, err
}

func (up *BasicUpdatePlanner) ExecuteCreateIndex(createIndexData *parse.CreateIndexData, tx *transaction.Transaction) (int, error) {
	indexName := createIndexData.IndexName()
	tableName := createIndexData.TableName()
	fieldName := createIndexData.FieldName()

	err := up.mdm.CreateIndex(indexName, tableName, fieldName, tx)
	return 0, err
}
