package plan

import (
	"fmt"

	"s1mpleasia.com/tinydb/parse"
	"s1mpleasia.com/tinydb/transaction"
)

type QueryPlanner interface {
	CreatePlan(queryData *parse.QueryData, tx *transaction.Transaction) (Plan, error)
}

type UpdatePlanner interface {
	ExecuteInsert(insertData *parse.InsertData, tx *transaction.Transaction) (int, error)
	ExecuteDelete(deleteData *parse.DeleteData, tx *transaction.Transaction) (int, error)
	ExecuteModify(modifyData *parse.ModifyData, tx *transaction.Transaction) (int, error)
	ExecuteCreateTable(createTableData *parse.CreateTableData, tx *transaction.Transaction) (int, error)
	ExecuteCreateView(createViewData *parse.CreateViewData, tx *transaction.Transaction) (int, error)
	ExecuteCreateIndex(createIndexData *parse.CreateIndexData, tx *transaction.Transaction) (int, error)
}

type Planner struct {
	queryPlanner  QueryPlanner
	updatePlanner UpdatePlanner
}

func NewPlanner(queryPlanner QueryPlanner, updatePlanner UpdatePlanner) *Planner {
	return &Planner{queryPlanner, updatePlanner}
}

func (p *Planner) CreateQueryPlan(query string, tx *transaction.Transaction) (Plan, error) {
	parser, err := parse.NewParser(query)
	if err != nil {
		return nil, err
	}

	queryData, err := parser.Query()
	if err != nil {
		return nil, err
	}

	err = p.verifyQuery(queryData)
	if err != nil {
		return nil, err
	}

	return p.queryPlanner.CreatePlan(queryData, tx)
}

func (p *Planner) ExecuteUpdate(cmd string, tx *transaction.Transaction) (int, error) {
	parser, err := parse.NewParser(cmd)
	if err != nil {
		return 0, err
	}

	updateCmd, err := parser.UpdateCmd()
	if err != nil {
		return 0, err
	}

	err = p.verifyUpdate(updateCmd)
	if err != nil {
		return 0, err
	}

	switch cmd := updateCmd.(type) {
	case *parse.InsertData:
		return p.updatePlanner.ExecuteInsert(cmd, tx)
	case *parse.DeleteData:
		return p.updatePlanner.ExecuteDelete(cmd, tx)
	case *parse.ModifyData:
		return p.updatePlanner.ExecuteModify(cmd, tx)
	case *parse.CreateTableData:
		return p.updatePlanner.ExecuteCreateTable(cmd, tx)
	case *parse.CreateViewData:
		return p.updatePlanner.ExecuteCreateView(cmd, tx)
	case *parse.CreateIndexData:
		return p.updatePlanner.ExecuteCreateIndex(cmd, tx)
	default:
		return 0, fmt.Errorf("unexpected update command: %v", cmd)
	}
}

func (p *Planner) verifyQuery(queryData *parse.QueryData) error {
	// TODO: Implement
	return nil
}

func (p *Planner) verifyUpdate(updateCmd parse.UpdateCmd) error {
	// TODO: Implement
	return nil
}
