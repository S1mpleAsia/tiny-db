package query

import (
	"errors"

	"s1mpleasia.com/tinydb/record"
)

var ErrNotUpdatable = errors.New("scan is not updatable")

/* 	Select scan keep the original columns, retrieve records that sastified the predicate
 */
type SelectScan struct {
	scan Scan
	pred *Predicate
}

func NewSelectScan(s Scan, pred *Predicate) *SelectScan {
	return &SelectScan{s, pred}
}

// Scan method
func (ss *SelectScan) BeforeFirst() error {
	return ss.scan.BeforeFirst()
}

func (ss *SelectScan) Next() bool {
	for ss.scan.Next() {
		sastified, err := ss.pred.IsSatisfied(ss.scan)
		if err != nil {
			panic(err)
		}

		if sastified {
			return true
		}
	}

	return false
}

func (ss *SelectScan) GetInt(fieldName string) (int32, error) {
	return ss.scan.GetInt(fieldName)
}

func (ss *SelectScan) GetString(fieldName string) (string, error) {
	return ss.scan.GetString(fieldName)
}

func (ss *SelectScan) GetVal(fieldName string) (*record.Constant, error) {
	return ss.scan.GetVal(fieldName)
}

func (ss *SelectScan) HasField(fieldName string) bool {
	return ss.scan.HasField(fieldName)
}

func (ss *SelectScan) Close() {
	ss.scan.Close()
}

// Update scan method
func (ss *SelectScan) SetInt(fieldName string, val int32) error {
	us, ok := ss.scan.(UpdateScan)
	if !ok {
		return ErrNotUpdatable
	}

	return us.SetInt(fieldName, val)
}

func (ss *SelectScan) SetString(fieldName string, val string) error {
	us, ok := ss.scan.(UpdateScan)
	if !ok {
		return ErrNotUpdatable
	}

	return us.SetString(fieldName, val)
}

func (ss *SelectScan) SetVal(fieldName string, val *record.Constant) error {
	us, ok := ss.scan.(UpdateScan)
	if !ok {
		return ErrNotUpdatable
	}

	return us.SetVal(fieldName, val)
}

func (ss *SelectScan) Insert() error {
	us, ok := ss.scan.(UpdateScan)
	if !ok {
		return ErrNotUpdatable
	}

	return us.Insert()
}

func (ss *SelectScan) Delete() error {
	us, ok := ss.scan.(UpdateScan)
	if !ok {
		return ErrNotUpdatable
	}
	return us.Delete()
}

func (ss *SelectScan) GetRID() *record.RID {
	us, ok := ss.scan.(UpdateScan)
	if !ok {
		return nil
	}

	return us.GetRID()
}

func (ss *SelectScan) MoveToRID(rid *record.RID) {
	us, ok := ss.scan.(UpdateScan)
	if !ok {
		panic(ErrNotUpdatable)
	}

	us.MoveToRID(rid)
}
