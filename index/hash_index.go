package index

import (
	"strconv"

	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ query.Index = (*HashIndex)(nil)

const NUM_BUCKETS = 100

type HashIndex struct {
	tx        *transaction.Transaction
	idxName   string
	layout    *record.Layout
	searchKey *record.Constant
	ts        *query.TableScan
}

func NewHashIndex(tx *transaction.Transaction, idxName string, layout *record.Layout) *HashIndex {
	return &HashIndex{
		tx:        tx,
		idxName:   idxName,
		layout:    layout,
		searchKey: nil,
		ts:        nil,
	}
}

// This will open TableScan for the bucket that matched with search key
func (hi *HashIndex) BeforeFirst(searchKey *record.Constant) error {
	hi.Close()
	hi.searchKey = searchKey

	bucket := searchKey.HashCode() % NUM_BUCKETS
	tblName := hi.idxName + strconv.Itoa(int(bucket))

	ts, err := query.NewTableScan(hi.tx, tblName, hi.layout)
	if err != nil {
		return err
	}

	hi.ts = ts
	return nil
}

func (hi *HashIndex) Next() (bool, error) {
	for hi.ts.Next() {
		dataval, err := hi.ts.GetVal("dataval")
		if err != nil {
			return false, err
		}

		if dataval.Equals(hi.searchKey) {
			return true, nil
		}
	}

	return false, nil
}

func (hi *HashIndex) GetDataRID() (*record.RID, error) {
	blockNum, err := hi.ts.GetInt("block")
	if err != nil {
		return nil, err
	}

	id, err := hi.ts.GetInt("id")
	if err != nil {
		return nil, err
	}

	return record.NewRID(blockNum, id), nil
}

// Move to the beginning of the bucket. Find the empty slot -> insert
func (hi *HashIndex) Insert(dataval *record.Constant, datarid *record.RID) error {
	err := hi.BeforeFirst(dataval)
	if err != nil {
		return err
	}

	err = hi.ts.Insert()
	if err != nil {
		return err
	}

	hi.ts.SetInt("block", datarid.BlockNumber())
	hi.ts.SetInt("id", datarid.Slot())
	hi.ts.SetVal("dataval", dataval)

	return nil
}

func (hi *HashIndex) Delete(dataval *record.Constant, datarid *record.RID) error {
	err := hi.BeforeFirst(dataval)
	if err != nil {
		return err
	}

	for {
		hasNext, err := hi.Next()
		if err != nil || !hasNext {
			return err
		}

		if rid, err := hi.GetDataRID(); err != nil {
			return err
		} else if rid.Equals(datarid) {
			return hi.ts.Delete()
		}
	}
}

func (hi *HashIndex) Close() error {
	if hi.ts != nil {
		hi.ts.Close()
	}

	return nil
}

func SearchCost(numBlock int, recordsPerBlock int) int {
	return numBlock / recordsPerBlock
}
