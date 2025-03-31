package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
)

var _ AggregateFn = (*MinFn)(nil)

type MinFn struct {
	fieldName string
	val       *record.Constant
}

func NewMinFn(fieldName string) *MinFn {
	return &MinFn{fieldName: fieldName, val: nil}
}

func (m *MinFn) ProcessFirst(scan Scan) error {
	val, err := scan.GetVal(m.fieldName)
	if err != nil {
		return fmt.Errorf("scan.GetVal(): %w", err)
	}

	m.val = val
	return nil
}

func (m *MinFn) ProcessNext(scan Scan) error {
	val, err := scan.GetVal(m.fieldName)
	if err != nil {
		return fmt.Errorf("scan.GetVal(): %w", err)
	}

	cmp, err := val.CompareTo(m.val)
	if err != nil {
		return fmt.Errorf("val.CompareTo(): %w", err)
	}

	if cmp < 0 {
		m.val = val
	}
	return nil
}

func (m *MinFn) FieldName() string {
	return fmt.Sprintf("min(%s)", m.fieldName)
}

func (m *MinFn) Value() *record.Constant {
	return m.val
}
