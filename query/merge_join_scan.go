package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
)

var _ Scan = (*MergeJoinScan)(nil)

type MergeJoinScan struct {
	s1, s2             *SortScan
	fldName1, fldName2 string
	joinVal            *record.Constant
}

func NewMergeJoinScan(s1, s2 *SortScan, fldName1, fldName2 string) *MergeJoinScan {
	return &MergeJoinScan{
		s1:       s1,
		s2:       s2,
		fldName1: fldName1,
		fldName2: fldName2,
		joinVal:  nil,
	}
}

func (mjs *MergeJoinScan) BeforeFirst() error {
	err := mjs.s1.BeforeFirst()
	if err != nil {
		return fmt.Errorf("mjs.s1.BeforeFirst(): %w", err)
	}

	err = mjs.s2.BeforeFirst()
	if err != nil {
		return fmt.Errorf("mjs.s2.BeforeFirst(): %w", err)
	}

	return nil
}

func (mjs *MergeJoinScan) Next() bool {
	var v1, v2 *record.Constant
	var err error
	hasMore2 := mjs.s2.Next()

	if hasMore2 {
		v2, err = mjs.s2.GetVal(mjs.fldName2)
		if err != nil {
			return false
		}

		if mjs.joinVal != nil && v2.Equals(mjs.joinVal) {
			return true
		}
	}

	hasMore1 := mjs.s1.Next()
	if hasMore1 {
		v1, err = mjs.s1.GetVal(mjs.fldName1)
		if err != nil {
			return false
		}

		if mjs.joinVal != nil && v1.Equals(mjs.joinVal) {
			err = mjs.s2.RestorePosition()
			if err != nil {
				return false
			}

			return true
		}
	}

	for hasMore1 && hasMore2 {
		v1, err = mjs.s1.GetVal(mjs.fldName1)
		if err != nil {
			return false
		}

		v2, err = mjs.s2.GetVal(mjs.fldName2)
		if err != nil {
			return false
		}

		cmp, err := v1.CompareTo(v2)
		if err != nil {
			return false
		}

		if cmp < 0 {
			hasMore1 = mjs.s1.Next()
		} else if cmp > 0 {
			hasMore2 = mjs.s2.Next()
		} else {
			err = mjs.s2.SavePosition()
			if err != nil {
				return false
			}

			val, err := mjs.s2.GetVal(mjs.fldName2)
			if err != nil {
				return false
			}

			mjs.joinVal = val
			return true
		}
	}

	return false
}

func (mjs *MergeJoinScan) GetInt(fieldName string) (int32, error) {
	if mjs.s1.HasField(fieldName) {
		return mjs.s1.GetInt(fieldName)
	}

	return mjs.s2.GetInt(fieldName)
}

func (mjs *MergeJoinScan) GetString(fieldName string) (string, error) {
	if mjs.s1.HasField(fieldName) {
		return mjs.s1.GetString(fieldName)
	}
	return mjs.s2.GetString(fieldName)
}

func (mjs *MergeJoinScan) GetVal(fieldName string) (*record.Constant, error) {
	if mjs.s1.HasField(fieldName) {
		return mjs.s1.GetVal(fieldName)
	}
	return mjs.s2.GetVal(fieldName)
}

func (mjs *MergeJoinScan) HasField(fieldName string) bool {
	return mjs.s1.HasField(fieldName) || mjs.s2.HasField(fieldName)
}

func (mjs *MergeJoinScan) Close() {
	mjs.s1.Close()
	mjs.s2.Close()
}
