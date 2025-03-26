package plan

import (
	"errors"
	"s1mpleasia.com/tinydb/metadata"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

var _ Plan = (*IndexSelectPlan)(nil)

type IndexSelectPlan struct {
	plan      Plan
	indexInfo *metadata.IndexInfo
	val       *record.Constant
}

func NewIndexSelectPlan(p Plan, indexInfo *metadata.IndexInfo, val *record.Constant) *IndexSelectPlan {
	return &IndexSelectPlan{p, indexInfo, val}
}

func (i *IndexSelectPlan) Open() (query.Scan, error) {
	scan, err := i.plan.Open()
	if err != nil {
		return nil, err
	}

	ts, ok := scan.(*query.TableScan)
	if !ok {
		return nil, errors.New("Open: plan is not a table plan")
	}

	idx := i.indexInfo.Open()
	return query.NewIndexSelectScan(ts, idx, i.val), nil
}

func (i *IndexSelectPlan) BlockAccessed() int32 {
	return i.indexInfo.BlockAccessed() + i.RecordsOutput()
}

func (i *IndexSelectPlan) RecordsOutput() int32 {
	return i.indexInfo.RecordsOutput()
}

func (i *IndexSelectPlan) DistinctValues(fieldName string) int32 {
	return i.indexInfo.DistinctValues(fieldName)
}

func (i *IndexSelectPlan) Schema() *record.Schema {
	return i.plan.Schema()
}
