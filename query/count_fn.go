package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
)

var _ AggregateFn = (*CountFn)(nil)

type CountFn struct {
	fieldName string
	count     int32
}

func NewCountFn(fieldName string) *CountFn {
	return &CountFn{
		fieldName: fieldName,
		count:     0,
	}
}

func (c *CountFn) ProcessFirst(scan Scan) error {
	c.count = 1
	return nil
}

func (c *CountFn) ProcessNext(scan Scan) error {
	c.count++
	return nil
}

func (c *CountFn) FieldName() string {
	return fmt.Sprintf("count(%s)", c.fieldName)
}

func (c *CountFn) Value() *record.Constant {
	return record.NewConstantWithInt(c.count)
}
