package query

import "s1mpleasia.com/tinydb/record"

/*
	A Scan object corresponds to a node in a query tree. TinyDB contains a Scan for each relational operator
*/
type Scan interface {
	BeforeFirst() error
	Next() bool
	GetInt(fieldName string) (int32, error)
	GetString(fieldName string) (string, error)
	GetVal(fieldName string) (*record.Constant, error)
	HasField(fieldName string) bool
	Close()
}

type UpdateScan interface {
	Scan
	SetInt(fieldName string, val int32) error
	SetString(fieldName string, val string) error
	SetVal(fieldName string, val *record.Constant) error
	Insert() error
	Delete() error
	GetRID() *record.RID
	MoveToRID(rid *record.RID)
}
