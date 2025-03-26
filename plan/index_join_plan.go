package plan

import (
	"s1mpleasia.com/tinydb/metadata"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

var _ Plan = (*IndexJoinPlan)(nil)

type IndexJoinPlan struct {
	plan1     Plan
	plan2     Plan
	indexInfo *metadata.IndexInfo // T2's index on B
	joinField string
	schema    *record.Schema
}

func NewIndexJoinPlan(p1 Plan, p2 Plan, indexInfo *metadata.IndexInfo, joinField string) *IndexJoinPlan {
	sch := record.NewSchema()
	sch.AddAll(p1.Schema())
	sch.AddAll(p2.Schema())

	return &IndexJoinPlan{p1, p2, indexInfo, joinField, sch}
}

func (i *IndexJoinPlan) Open() (query.Scan, error) {
	scan1, err := i.plan1.Open()
	if err != nil {
		return nil, err
	}

	scan2, err := i.plan2.Open()
	if err != nil {
		return nil, err
	}

	idx := i.indexInfo.Open()
	return query.NewIndexJoinScan(scan1, idx, i.joinField, scan2)
}

func (i *IndexJoinPlan) BlockAccessed() int32 {
	return i.plan1.BlockAccessed() + (i.plan1.RecordsOutput() + i.indexInfo.BlockAccessed()) + i.RecordsOutput()
}

func (i *IndexJoinPlan) RecordsOutput() int32 {
	return i.plan1.RecordsOutput() + i.indexInfo.RecordsOutput()
}

func (i *IndexJoinPlan) DistinctValues(fieldName string) int32 {
	if i.plan1.Schema().HasField(fieldName) {
		return i.plan1.DistinctValues(fieldName)
	}

	return i.plan2.DistinctValues(fieldName)
}

func (i *IndexJoinPlan) Schema() *record.Schema {
	return i.schema
}
