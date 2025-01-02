package query

import (
	"errors"

	"s1mpleasia.com/tinydb/record"
)

var ErrFieldNotFound = errors.New("field not found")

/*
Project scan keep all records in the table, filtering desired columns
*/
type ProjectScan struct {
	scan      Scan
	fieldList []string
}

func NewProjectScan(s Scan, fieldList []string) *ProjectScan {
	return &ProjectScan{s, fieldList}
}

func (ps *ProjectScan) BeforeFirst() error {
	return ps.scan.BeforeFirst()
}

func (ps *ProjectScan) Next() bool {
	return ps.scan.Next()
}

func (ps *ProjectScan) GetInt(fieldName string) (int32, error) {
	if ps.HasField(fieldName) {
		return ps.scan.GetInt(fieldName)
	}

	return 0, ErrFieldNotFound
}

func (ps *ProjectScan) GetString(fieldName string) (string, error) {
	if ps.HasField(fieldName) {
		return ps.scan.GetString(fieldName)
	}

	return "", ErrFieldNotFound
}

func (ps *ProjectScan) GetVal(fieldName string) (*record.Constant, error) {
	if ps.HasField(fieldName) {
		return ps.scan.GetVal(fieldName)
	}

	return nil, ErrFieldNotFound
}

func (ps *ProjectScan) HasField(fieldName string) bool {
	for _, field := range ps.fieldList {
		if field == fieldName {
			return true
		}
	}
	return false
}

func (ps *ProjectScan) Close() {
	ps.scan.Close()
}
