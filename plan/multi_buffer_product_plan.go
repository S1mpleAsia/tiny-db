package plan

import (
	"fmt"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ Plan = (*MultiBufferProductPlan)(nil)

type MultiBufferProductPlan struct {
	tx       *transaction.Transaction
	lhs, rhs Plan
	schema   *record.Schema
}

func NewMultiBufferProductPlan(tx *transaction.Transaction, lhs, rhs Plan) *MultiBufferProductPlan {
	schema := record.NewSchema()
	schema.AddAll(lhs.Schema())
	schema.AddAll(rhs.Schema())

	return &MultiBufferProductPlan{
		tx:     tx,
		lhs:    NewMaterializePlan(lhs, tx),
		rhs:    rhs,
		schema: schema,
	}
}

/*
	A scan for this query is created and returned, as follows.

First, the method materializes its LHS and RHS queries.
It then determines the optimal chunk size,
based on the size of the materialized RHS file and the
number of available buffers.
It creates a chunk plan for each chunk, saving them in a list.
Finally, it creates a multiscan for this list of plans,
and returns that scan.
@see simpledb.plan.Plan#open()
*/
func (m *MultiBufferProductPlan) Open() (query.Scan, error) {
	leftScan, err := m.lhs.Open()
	if err != nil {
		return nil, fmt.Errorf("p.lhs.Open: %w", err)
	}

	tt, err := m.copyRecordsFrom(m.rhs)
	if err != nil {
		return nil, fmt.Errorf("p.rhs.Open: %w", err)
	}

	return query.NewMultiBufferProductScan(m.tx, leftScan, tt.TableName(), tt.Layout())
}

func (m *MultiBufferProductPlan) BlockAccessed() int32 {
	avail := m.tx.AvailableBuffs()
	size := NewMaterializePlan(m.rhs, m.tx).BlockAccessed()
	numChunks := size / int32(avail)

	blockAccessed := m.rhs.BlockAccessed() + (m.lhs.BlockAccessed() * numChunks)

	fmt.Printf("BlockAccessed(): numChunks = size(%d) / avail(%d) = %d\n", size, avail, numChunks)
	fmt.Printf("BlockAccessed() = rhs(%d) + (lhs(%d) * numChunks(%d)) = %d\n", m.rhs.BlockAccessed(), m.lhs.BlockAccessed(), numChunks, blockAccessed)
	return blockAccessed
}

func (m *MultiBufferProductPlan) RecordsOutput() int32 {
	return m.lhs.RecordsOutput() * m.rhs.RecordsOutput()
}

func (m *MultiBufferProductPlan) DistinctValues(fieldName string) int32 {
	if m.lhs.Schema().HasField(fieldName) {
		return m.lhs.DistinctValues(fieldName)
	}
	return m.rhs.DistinctValues(fieldName)
}

func (m *MultiBufferProductPlan) Schema() *record.Schema {
	return m.schema
}

func (m *MultiBufferProductPlan) copyRecordsFrom(plan Plan) (*query.TempTable, error) {
	src, err := plan.Open()
	if err != nil {
		return nil, fmt.Errorf("plan.Open: %w", err)
	}
	defer src.Close()

	sch := plan.Schema()
	t := query.NewTempTable(m.tx, sch)
	dest, err := t.Open()
	if err != nil {
		return nil, fmt.Errorf("t.Open: %w", err)
	}
	defer dest.Close()

	for {
		next := src.Next()
		if !next {
			break
		}

		err = dest.Insert()
		if err != nil {
			return nil, fmt.Errorf("dest.Insert: %w", err)
		}

		for _, fldName := range sch.Fields() {
			val, err := src.GetVal(fldName)
			if err != nil {
				return nil, fmt.Errorf("src.GetVal(%s): %w", fldName, err)
			}

			err = dest.SetVal(fldName, val)
			if err != nil {
				return nil, fmt.Errorf("dest.SetVal(%s, %s): %w", fldName, val, err)
			}
		}
	}

	return t, nil
}
