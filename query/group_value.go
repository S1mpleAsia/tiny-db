package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
)

// Hold information about the current group
type GroupValue struct {
	vals map[string]*record.Constant
}

// fields is all sorted-fields filter
func NewGroupValue(scan Scan, fields []string) (*GroupValue, error) {
	vals := make(map[string]*record.Constant)
	for _, fieldName := range fields {
		val, err := scan.GetVal(fieldName)
		if err != nil {
			return nil, fmt.Errorf("scan.GetVal(): %w", err)
		}

		vals[fieldName] = val
	}

	return &GroupValue{vals: vals}, nil
}

func (gv *GroupValue) GetVal(fieldName string) (*record.Constant, error) {
	val, ok := gv.vals[fieldName]
	if !ok {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}

	return val, nil
}

func (gv *GroupValue) Equal(other *GroupValue) bool {
	if len(gv.vals) != len(other.vals) {
		return false
	}

	for fieldName, val := range gv.vals {
		otherVal, ok := other.vals[fieldName]
		if !ok {
			return false
		}

		if !val.Equals(otherVal) {
			return false
		}
	}

	return true
}
