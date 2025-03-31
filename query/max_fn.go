package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
)

var _ AggregateFn = (*MaxFn)(nil)

type MaxFn struct {
	fieldName string
	val       *record.Constant
}

func NewMaxFn(fieldName string) *MaxFn {
	return &MaxFn{fieldName: fieldName, val: nil}
}

func (m *MaxFn) ProcessFirst(scan Scan) error {
	val, err := scan.GetVal(m.fieldName)
	if err != nil {
		return fmt.Errorf("scan.GetVal(): %w", err)
	}

	m.val = val
	return nil
}

func (m *MaxFn) ProcessNext(scan Scan) error {
	val, err := scan.GetVal(m.fieldName)
	if err != nil {
		return fmt.Errorf("scan.GetVal(): %w", err)
	}

	cmp, err := val.CompareTo(m.val)
	if err != nil {
		return fmt.Errorf("val.CompareTo(): %w", err)
	}

	if cmp > 0 {
		m.val = val
	}

	return nil
}

func (m *MaxFn) FieldName() string {
	return fmt.Sprintf("max(%s)", m.fieldName)
}

func (m *MaxFn) Value() *record.Constant {
	return m.val
}
