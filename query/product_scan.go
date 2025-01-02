package query

import (
	"errors"

	"s1mpleasia.com/tinydb/record"
)

var ErrAmbiguousField = errors.New("ambiguous field")

/*
Product scan make a full join between table_1 and table_2, iterating through all possible combinations of records
from its underlying scans s1 and s2
*/
type ProductScan struct {
	s1 Scan
	s2 Scan
}

func NewProductScan(s1, s2 Scan) (*ProductScan, error) {
	ps := &ProductScan{
		s1: s1,
		s2: s2,
	}

	if err := ps.BeforeFirst(); err != nil {
		return nil, err
	}

	return ps, nil
}

func (ps *ProductScan) BeforeFirst() error {
	if err := ps.s1.BeforeFirst(); err != nil {
		return err
	}

	ps.s1.Next()
	if err := ps.s2.BeforeFirst(); err != nil {
		return err
	}

	return nil
}

func (ps *ProductScan) Next() bool {
	if ps.s2.Next() {
		return true
	} else {
		err := ps.s2.BeforeFirst()
		if err != nil {
			panic(err)
		}

		return ps.s1.Next() && ps.s2.Next()
	}
}

func (ps *ProductScan) GetInt(fieldName string) (int32, error) {
	if ps.s1.HasField(fieldName) {
		return ps.s1.GetInt(fieldName)
	} else {
		return ps.s2.GetInt(fieldName)
	}
}

func (ps *ProductScan) GetString(fieldName string) (string, error) {
	if ps.s1.HasField(fieldName) {
		return ps.s1.GetString(fieldName)
	} else {
		return ps.s2.GetString(fieldName)
	}
}

func (ps *ProductScan) GetVal(fieldName string) (*record.Constant, error) {
	if ps.s1.HasField(fieldName) {
		return ps.s1.GetVal(fieldName)
	} else {
		return ps.s2.GetVal(fieldName)
	}
}

func (ps *ProductScan) HasField(fieldName string) bool {
	return ps.s1.HasField(fieldName) || ps.s2.HasField(fieldName)
}

func (ps *ProductScan) Close() {
	ps.s1.Close()
	ps.s2.Close()
}
