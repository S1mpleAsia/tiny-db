package query

import "s1mpleasia.com/tinydb/record"

type AggregateFn interface {
	ProcessFirst(scan Scan) error // Start a new group using the current record as the first record of the group
	ProcessNext(scan Scan) error  // Add another record into current group
	FieldName() string
	Value() *record.Constant
}
