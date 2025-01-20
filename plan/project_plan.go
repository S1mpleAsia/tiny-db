package plan

import (
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

var _ Plan = (*ProjectPlan)(nil)

/*
Project scan keeps all records in the table, filtering only desired columns
*/
type ProjectPlan struct {
	plan   Plan
	schema *record.Schema
}

func NewProjectPlan(p Plan, fieldList []string) (*ProjectPlan, error) {
	sch := record.NewSchema()
	for _, field := range fieldList {
		sch.Add(field, p.Schema())
	}

	return &ProjectPlan{p, sch}, nil
}

func (p *ProjectPlan) Open() (query.Scan, error) {
	scan, err := p.plan.Open()
	if err != nil {
		return nil, err
	}

	return query.NewProjectScan(scan, p.schema.Fields()), nil
}

func (p *ProjectPlan) BlockAccessed() int32 {
	return p.plan.BlockAccessed()
}

func (p *ProjectPlan) RecordsOutput() int32 {
	return p.plan.RecordsOutput()
}

func (p *ProjectPlan) DistinctValues(fieldName string) int32 {
	return p.plan.DistinctValues(fieldName)
}

func (p *ProjectPlan) Schema() *record.Schema {
	return p.schema
}
