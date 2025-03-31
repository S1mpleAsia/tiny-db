package plan

import (
	"errors"
	"fmt"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ Plan = (*MergeJoinPlan)(nil)

type MergeJoinPlan struct {
	p1, p2             *SortPlan
	fldName1, fldName2 string
	sch                *record.Schema
}

func NewMergeJoinPlan(tx *transaction.Transaction, p1, p2 Plan, fldName1, fldName2 string) (*MergeJoinPlan, error) {
	sortPlan1, err := NewSortPlan(p1, tx, []string{fldName1})
	if err != nil {
		return nil, fmt.Errorf("NewSortPlan for p1: %w", err)
	}
	sortPlan2, err := NewSortPlan(p2, tx, []string{fldName2})
	if err != nil {
		return nil, fmt.Errorf("NewSortPlan for p2: %w", err)
	}

	sch := record.NewSchema()
	sch.AddAll(p1.Schema())
	sch.AddAll(p2.Schema())

	return &MergeJoinPlan{
		p1:       sortPlan1,
		p2:       sortPlan2,
		fldName1: fldName1,
		fldName2: fldName2,
		sch:      sch,
	}, nil
}

func (mjp *MergeJoinPlan) Open() (query.Scan, error) {
	s1, err := mjp.p1.Open()
	if err != nil {
		return nil, fmt.Errorf("mjp.p1.Open: %w", err)
	}

	sortScan1, ok := s1.(*query.SortScan)
	if !ok {
		return nil, errors.New("s1 is not a SortScan")
	}

	s2, err := mjp.p2.Open()
	if err != nil {
		return nil, fmt.Errorf("mjp.p2.Open: %w", err)
	}
	sortScan2, ok := s2.(*query.SortScan)
	if !ok {
		return nil, errors.New("s2 is not a SortScan")
	}

	return query.NewMergeJoinScan(sortScan1, sortScan2, mjp.fldName1, mjp.fldName2), nil
}

func (mjp *MergeJoinPlan) BlockAccessed() int32 {
	return mjp.p1.BlockAccessed() + mjp.p2.BlockAccessed()
}

func (mjp *MergeJoinPlan) RecordsOutput() int32 {
	maxVals := max(mjp.p1.DistinctValues(mjp.fldName1), mjp.p2.DistinctValues(mjp.fldName2))
	return mjp.p1.RecordsOutput() * mjp.p2.RecordsOutput() / maxVals
}

func (mjp *MergeJoinPlan) DistinctValues(fieldName string) int32 {
	if mjp.p1.Schema().HasField(fieldName) {
		return mjp.p1.DistinctValues(fieldName)
	}

	if mjp.p2.Schema().HasField(fieldName) {
		return mjp.p2.DistinctValues(fieldName)
	}

	panic("field not found")
}

func (mjp *MergeJoinPlan) Schema() *record.Schema {
	return mjp.sch
}
