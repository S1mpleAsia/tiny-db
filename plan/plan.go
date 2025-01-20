package plan

import (
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

type Plan interface {
	Open() (query.Scan, error)
	BlockAccessed() int32
	RecordsOutput() int32
	DistinctValues(fieldName string) int32
	Schema() *record.Schema
}
