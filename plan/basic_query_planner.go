package plan

import (
	"s1mpleasia.com/tinydb/metadata"
	"s1mpleasia.com/tinydb/parse"
	"s1mpleasia.com/tinydb/transaction"
)

var _ QueryPlanner = (*BasicQueryPlanner)(nil)

type BasicQueryPlanner struct {
	mdm *metadata.MetadataMgmt
}

func NewBasicQueryPlanner(mdm *metadata.MetadataMgmt) *BasicQueryPlanner {
	return &BasicQueryPlanner{mdm}
}

func (qp *BasicQueryPlanner) CreatePlan(queryData *parse.QueryData, tx *transaction.Transaction) (Plan, error) {
	var result Plan
	var err error

	plans := make([]Plan, 0, 5)

	// Step 1: Create a plan for each related tables or views
	for _, tableName := range queryData.Tables() {
		viewDef, err := qp.mdm.GetViewDef(tableName, tx)
		if err != nil {
			return nil, err
		}

		if viewDef != "" {
			parser, err := parse.NewParser(viewDef)
			if err != nil {
				return nil, err
			}

			viewdata, err := parser.Query()
			if err != nil {
				return nil, err
			}

			plan, err := qp.CreatePlan(viewdata, tx)
			if err != nil {
				return nil, err
			}

			plans = append(plans, plan)
			continue
		}

		plan, err := NewTablePlan(tx, tableName, qp.mdm)
		if err != nil {
			return nil, err
		}

		plans = append(plans, plan)
	}

	// Step 2: Create the product of all table plans
	result = plans[0]
	for _, plan := range plans[1:] {
		result, err = NewProductPlan(result, plan)
		if err != nil {
			return nil, err
		}
	}

	// Step 3: Add a selection plan for the predicate
	result, err = NewSelectPlan(result, queryData.Predicate())
	if err != nil {
		return nil, err
	}

	// Step 4: Project on specified field names
	result, err = NewProjectPlan(result, queryData.Fields())

	return result, err
}
