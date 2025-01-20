package plan

import (
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

// var _ Plan = (*)

type ProductPlan struct {
	p1, p2 Plan
	schema *record.Schema
}

func NewProductPlan(p1 Plan, p2 Plan) (*ProductPlan, error) {
	sch := record.NewSchema()

	sch.AddAll(p1.Schema())
	sch.AddAll(p2.Schema())

	return &ProductPlan{p1, p2, sch}, nil
}

func (p *ProductPlan) Open() (query.Scan, error) {
	s1, err := p.p1.Open()
	if err != nil {
		return nil, err
	}

	s2, err := p.p2.Open()
	if err != nil {
		return nil, err
	}

	return query.NewProductScan(s1, s2)
}

func (p *ProductPlan) BlockAccessed() int32 {
	return p.p1.BlockAccessed() + (p.p1.RecordsOutput() * p.p2.BlockAccessed())
}

func (p *ProductPlan) RecordsOutput() int32 {
	return p.p1.RecordsOutput() * p.p2.RecordsOutput()
}

func (p *ProductPlan) DistinctValues(fieldName string) int32 {
	if p.p1.Schema().HasField(fieldName) {
		return p.p1.DistinctValues(fieldName)
	} else {
		return p.p2.DistinctValues(fieldName)
	}
}

func (p *ProductPlan) Schema() *record.Schema {
	return p.schema
}
