package query

import (
	"s1mpleasia.com/tinydb/record"
)

type IndexSelectScan struct {
	scan Scan
	idx  Index
	val  *record.Constant
}

var _ Scan = (*IndexSelectScan)(nil)

func NewIndexSelectScan(scan Scan, idx Index, val *record.Constant) *IndexSelectScan {
	return &IndexSelectScan{
		scan: scan,
		idx:  idx,
		val:  val,
	}
}

func (i *IndexSelectScan) BeforeFirst() error {
	return i.idx.BeforeFirst(i.val)
}

// Move to next index record
func (i *IndexSelectScan) Next() bool {
	ok, err := i.idx.Next()
	if err != nil || !ok {
		return false
	}

	rid, err := i.idx.GetDataRID()
	if err != nil {
		return false
	}

	us, ok := i.scan.(UpdateScan)
	if !ok {
		return false
	}

	us.MoveToRID(rid)
	return true
}

func (i *IndexSelectScan) GetInt(fieldName string) (int32, error) {
	return i.scan.GetInt(fieldName)
}

func (i *IndexSelectScan) GetString(fieldName string) (string, error) {
	return i.scan.GetString(fieldName)
}

func (i *IndexSelectScan) GetVal(fieldName string) (*record.Constant, error) {
	return i.scan.GetVal(fieldName)
}

func (i *IndexSelectScan) HasField(fieldName string) bool {
	return i.scan.HasField(fieldName)
}

func (i *IndexSelectScan) Close() {
	if err := i.idx.Close(); err != nil {
		panic(err)
	}
	i.scan.Close()
}
