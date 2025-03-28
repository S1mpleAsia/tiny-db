package plan

import (
	"fmt"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ (Plan) = (*SortPlan)(nil)

type SortPlan struct {
	plan   Plan
	tx     *transaction.Transaction
	schema *record.Schema
	comp   *query.RecordComparator
}

func NewSortPlan(p Plan, tx *transaction.Transaction, sortFields []string) (*SortPlan, error) {
	return &SortPlan{
		plan:   p,
		tx:     tx,
		schema: p.Schema(),
		comp:   query.NewRecordComparator(sortFields),
	}, nil
}

func (sp *SortPlan) Open() (query.Scan, error) {
	src, err := sp.plan.Open()
	if err != nil {
		return nil, fmt.Errorf("sp.plan.Open(): %w", err)
	}

	runs, err := sp.splitIntoRuns(src)
	if err != nil {
		return nil, fmt.Errorf("sp.splitIntoRuns(): %w", err)
	}

	src.Close()
	for len(runs) > 2 {
		runs, err = sp.doAMergeIteration(runs)
		if err != nil {
			return nil, fmt.Errorf("doAMergeIteration(): %w", err)
		}
	}

	sc, err := query.NewSortScan(runs, sp.comp)
	if err != nil {
		return nil, fmt.Errorf("query.NewSortScan(): %w", err)
	}

	return sc, nil
}

func (sp *SortPlan) BlockAccessed() int32 {
	mp := NewMaterializePlan(sp.plan, sp.tx)
	return mp.BlockAccessed()
}

func (sp *SortPlan) RecordsOutput() int32 {
	return sp.plan.RecordsOutput()
}

func (sp *SortPlan) DistinctValues(fieldName string) int32 {
	return sp.plan.DistinctValues(fieldName)
}

func (sp *SortPlan) Schema() *record.Schema {
	return sp.schema
}

// Split into runs whether each run is individual sorted
/*
	2021 2020 2022 2023 2022 2020 2019 2021 2022 2020
 -> [2021] [2020 2022 2023] [2022] [2020] [2019 2021 2022] [2020]
*/
func (sp *SortPlan) splitIntoRuns(src query.Scan) ([]*query.TempTable, error) {
	temps := make([]*query.TempTable, 0)
	err := src.BeforeFirst()
	if err != nil {
		return nil, fmt.Errorf("src.BeforeFirst(): %w", err)
	}

	next := src.Next()
	if !next {
		return temps, nil
	}

	currentTemp := query.NewTempTable(sp.tx, sp.schema)
	temps = append(temps, currentTemp)
	currentScan, err := currentTemp.Open()
	if err != nil {
		return nil, err
	}

	for {
		next, err = sp.copy(src, currentScan)
		if err != nil {
			return nil, fmt.Errorf("sp.copy(): %w", err)
		}

		if !next {
			break
		}

		cmp, err := sp.comp.Compare(src, currentScan)
		if err != nil {
			return nil, fmt.Errorf("comp.Compare(): %w", err)
		}

		if cmp < 0 {
			currentScan.Close()
			currentTemp = query.NewTempTable(sp.tx, sp.schema)
			temps = append(temps, currentTemp)
			currentScan, err = currentTemp.Open()
			if err != nil {
				return nil, err
			}
		}
	}

	currentScan.Close()
	return temps, nil
}

// Do 2-merge
func (sp *SortPlan) doAMergeIteration(runs []*query.TempTable) ([]*query.TempTable, error) {
	result := make([]*query.TempTable, 0)
	for len(runs) > 1 {
		p1 := runs[0]
		p2 := runs[1]
		runs = runs[2:]
		merged, err := sp.mergeTwoRuns(p1, p2)
		if err != nil {
			return nil, fmt.Errorf("mergeTwoRuns(): %w", err)
		}

		result = append(result, merged)
	}

	if len(runs) == 1 {
		result = append(result, runs[0])
	}

	return result, nil
}

func (sp *SortPlan) mergeTwoRuns(p1 *query.TempTable, p2 *query.TempTable) (*query.TempTable, error) {
	src1, err := p1.Open()
	if err != nil {
		return nil, fmt.Errorf("p1.Open(): %w", err)
	}
	defer src1.Close()

	src2, err := p2.Open()
	if err != nil {
		return nil, fmt.Errorf("p2.Open(): %w", err)
	}
	defer src2.Close()

	result := query.NewTempTable(sp.tx, sp.schema)
	dest, err := result.Open()
	if err != nil {
		return nil, fmt.Errorf("result.Open(): %w", err)
	}
	defer dest.Close()

	hasMore1 := src1.Next()
	hasMore2 := src2.Next()

	for hasMore1 && hasMore2 {
		cmp, err := sp.comp.Compare(src1, src2)
		if err != nil {
			return nil, fmt.Errorf("comp.Compare(src1, src2): %w", err)
		}

		if cmp < 0 {
			hasMore1, err = sp.copy(src1, dest)
			if err != nil {
				return nil, fmt.Errorf("copy(src1, dest): %w", err)
			}
		} else {
			hasMore2, err = sp.copy(src2, dest)
			if err != nil {
				return nil, fmt.Errorf("copy(src2, dest): %w", err)
			}
		}
	}

	if hasMore1 {
		for hasMore1 {
			hasMore1, err = sp.copy(src1, dest)
			if err != nil {
				return nil, fmt.Errorf("copy(src1, dest): %w", err)
			}
		}
	} else {
		for hasMore2 {
			hasMore2, err = sp.copy(src2, dest)
			if err != nil {
				return nil, fmt.Errorf("copy(src2, dest): %w", err)
			}
		}
	}

	return result, nil
}

func (sp *SortPlan) copy(src query.Scan, dest query.UpdateScan) (bool, error) {
	err := dest.Insert()
	if err != nil {
		return false, fmt.Errorf("dest.Insert(): %w", err)
	}

	for _, fieldName := range sp.schema.Fields() {
		val, err := src.GetVal(fieldName)
		if err != nil {
			return false, fmt.Errorf("src.GetVal(%s): %w", fieldName, err)
		}

		err = dest.SetVal(fieldName, val)
		if err != nil {
			return false, fmt.Errorf("dest.SetVal(%s): %w", fieldName, err)
		}
	}

	next := src.Next()
	return next, nil
}
