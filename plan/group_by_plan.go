package plan

import (
	"fmt"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ Plan = (*GroupByPlan)(nil)

type GroupByPlan struct {
	plan        *SortPlan
	groupFields []string
	aggFns      []query.AggregateFn
	sch         *record.Schema
}

func NewGroupByPlan(tx *transaction.Transaction, plan Plan, groupFields []string, aggFns []query.AggregateFn) (*GroupByPlan, error) {
	sortPlan, err := NewSortPlan(plan, tx, groupFields)
	if err != nil {
		return nil, fmt.Errorf("NewSortPlan: %w", err)
	}

	sch := record.NewSchema()
	for _, fldName := range groupFields {
		sch.Add(fldName, plan.Schema())
	}

	for _, fn := range aggFns {
		sch.AddIntField(fn.FieldName())
	}

	return &GroupByPlan{
		plan:        sortPlan,
		groupFields: groupFields,
		aggFns:      aggFns,
		sch:         sch,
	}, nil
}

func (gp *GroupByPlan) Open() (query.Scan, error) {
	s, err := gp.plan.Open()
	if err != nil {
		return nil, fmt.Errorf("gp.plan.Open(): %w", err)
	}

	sortScan, ok := s.(*query.SortScan)
	if !ok {
		return nil, fmt.Errorf("gp.plan.Open(): wrong scan type: %T", s)
	}

	return query.NewGroupByScan(sortScan, gp.groupFields, gp.aggFns), nil
}

func (gp *GroupByPlan) BlockAccessed() int32 {
	return gp.plan.BlockAccessed()
}

func (gp *GroupByPlan) RecordsOutput() int32 {
	numGroups := int32(1)

	for _, fieldName := range gp.groupFields {
		numGroups *= gp.plan.DistinctValues(fieldName)
	}

	return numGroups
}

func (gp *GroupByPlan) DistinctValues(fieldName string) int32 {
	if gp.sch.HasField(fieldName) {
		return gp.plan.DistinctValues(fieldName)
	}

	return gp.RecordsOutput()
}

func (gp *GroupByPlan) Schema() *record.Schema {
	return gp.sch
}
