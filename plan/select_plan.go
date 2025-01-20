package plan

import (
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

var _ Plan = (*SelectPlan)(nil)

type SelectPlan struct {
	plan Plan
	pred *query.Predicate
}

func NewSelectPlan(p Plan, pred *query.Predicate) (*SelectPlan, error) {
	return &SelectPlan{p, pred}, nil
}

func (p *SelectPlan) Open() (query.Scan, error) {
	scan, err := p.plan.Open()
	if err != nil {
		return nil, err
	}

	return query.NewSelectScan(scan, p.pred), nil
}

func (p *SelectPlan) BlockAccessed() int32 {
	return p.plan.BlockAccessed()
}

func (p *SelectPlan) RecordsOutput() int32 {
	return p.plan.RecordsOutput() / p.pred.ReductionFactor(p.plan)
}

func (p *SelectPlan) DistinctValues(fieldName string) int32 {
	if p.pred.EquatesWithConstant(fieldName) != nil {
		return 1
	}

	otherField := p.pred.EquatesWithField(fieldName)
	if otherField != "" {
		return min(p.plan.DistinctValues(fieldName), p.plan.DistinctValues(otherField))
	}

	return p.plan.DistinctValues(fieldName)
}

func (p *SelectPlan) Schema() *record.Schema {
	return p.plan.Schema()
}
