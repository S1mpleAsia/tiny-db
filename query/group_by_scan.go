package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
	"slices"
)

var _ Scan = (*GroupByScan)(nil)

type GroupByScan struct {
	scan          *SortScan
	groupFields   []string
	aggFns        []AggregateFn
	groupValue    *GroupValue
	hasMoreGroups bool
}

func NewGroupByScan(scan *SortScan, groupFields []string, aggFns []AggregateFn) *GroupByScan {
	return &GroupByScan{
		scan:          scan,
		groupFields:   groupFields,
		aggFns:        aggFns,
		groupValue:    nil,
		hasMoreGroups: true,
	}
}

func (gs *GroupByScan) BeforeFirst() error {
	err := gs.scan.BeforeFirst()
	if err != nil {
		return fmt.Errorf("gc.BeforeFirst(): %w", err)
	}

	next := gs.scan.Next()
	gs.hasMoreGroups = next
	return nil
}

// Next filter only 1 group value for each call and pointer to the first element of the next group value
func (gs *GroupByScan) Next() bool {
	if !gs.hasMoreGroups {
		return false
	}

	for _, fn := range gs.aggFns {
		err := fn.ProcessFirst(gs.scan)
		if err != nil {
			panic(err)
		}
	}

	groupValue, err := NewGroupValue(gs.scan, gs.groupFields)
	if err != nil {
		return false
	}
	gs.groupValue = groupValue

	for {
		hasMore := gs.scan.Next()
		gs.hasMoreGroups = hasMore

		if !hasMore {
			break
		}

		gv, err := NewGroupValue(gs.scan, gs.groupFields)
		if err != nil {
			return false
		}

		if !gs.groupValue.Equal(gv) {
			break
		}

		for _, fn := range gs.aggFns {
			if err = fn.ProcessNext(gs.scan); err != nil {
				return false
			}
		}
	}

	return true
}

func (gs *GroupByScan) GetInt(fieldName string) (int32, error) {
	val, err := gs.GetVal(fieldName)
	if err != nil {
		return 0, fmt.Errorf("gs.GetVal(%s): %w", fieldName, err)
	}

	intVal, err := val.AsInt()
	if err != nil {
		return 0, fmt.Errorf("val.AsInt(%s): %w", fieldName, err)
	}
	return intVal, nil
}

func (gs *GroupByScan) GetString(fieldName string) (string, error) {
	val, err := gs.GetVal(fieldName)
	if err != nil {
		return "", fmt.Errorf("gs.GetVal(%s): %w", fieldName, err)
	}

	strVal, err := val.AsString()
	if err != nil {
		return "", fmt.Errorf("val.AsString(%s): %w", fieldName, err)
	}
	return strVal, nil
}

func (gs *GroupByScan) GetVal(fieldName string) (*record.Constant, error) {
	if slices.Contains(gs.groupFields, fieldName) {
		val, err := gs.groupValue.GetVal(fieldName)
		if err != nil {
			return nil, fmt.Errorf("gs.groupValue.GetVal(): %w", err)
		}
		return val, nil
	}

	for _, fn := range gs.aggFns {
		if fieldName == fn.FieldName() {
			return fn.Value(), nil
		}
	}

	return nil, fmt.Errorf("field %s not found", fieldName)
}

func (gs *GroupByScan) HasField(fieldName string) bool {
	if slices.Contains(gs.groupFields, fieldName) {
		return true
	}

	for _, fn := range gs.aggFns {
		if fieldName == fn.FieldName() {
			return true
		}
	}
	return false
}

func (gs *GroupByScan) Close() {
	gs.scan.Close()
}
