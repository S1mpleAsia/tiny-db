package plan

import (
	"fmt"
	"math"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ Plan = (*MaterializePlan)(nil)

type MaterializePlan struct {
	srcPlan Plan
	tx      *transaction.Transaction
}

func NewMaterializePlan(srcPlan Plan, tx *transaction.Transaction) *MaterializePlan {
	return &MaterializePlan{
		srcPlan: srcPlan,
		tx:      tx,
	}
}

func (m *MaterializePlan) Open() (query.Scan, error) {
	fmt.Println("MaterializePlan.Open")

	sch := m.srcPlan.Schema()
	tmp := query.NewTempTable(m.tx, sch)
	src, err := m.srcPlan.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open src plan: %w", err)
	}

	defer src.Close()
	dest, err := tmp.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open dest plan: %w", err)
	}

	for {
		next := src.Next()
		if !next {
			break
		}

		err = dest.Insert()
		if err != nil {
			return nil, fmt.Errorf("dest.Insert(): %w", err)
		}

		for _, fieldName := range sch.Fields() {
			val, err := src.GetVal(fieldName)
			if err != nil {
				return nil, fmt.Errorf("src.GetVal(): %w", err)
			}

			err = dest.SetVal(fieldName, val)
			if err != nil {
				return nil, fmt.Errorf("dest.SetVal(): %w", err)
			}
		}
	}

	err = dest.BeforeFirst()
	if err != nil {
		return nil, fmt.Errorf("dest.BeforeFirst(): %w", err)
	}

	fmt.Println("MaterializePlan.Open(): Done")
	return dest, nil
}

func (m *MaterializePlan) BlockAccessed() int32 {
	layout := record.NewLayoutFromSchema(m.srcPlan.Schema())
	rpb := float64(m.tx.BlockSize()) / float64(layout.SlotSize())
	blockAccessed := int32(math.Ceil(float64(m.srcPlan.RecordsOutput())) / rpb)

	return blockAccessed
}

func (m *MaterializePlan) RecordsOutput() int32 {
	return m.srcPlan.RecordsOutput()
}

func (m *MaterializePlan) DistinctValues(fieldName string) int32 {
	return m.srcPlan.DistinctValues(fieldName)
}

func (m *MaterializePlan) Schema() *record.Schema {
	return m.srcPlan.Schema()
}
