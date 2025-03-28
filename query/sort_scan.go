package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
)

var _ Scan = (*SortScan)(nil)

/*
	Sample: 2 6 20 4 1 16 19 3 18

	Run1: 2 6 20		| 1st iteration: Merge 1 and 2
	Run2: 4				| -> Run5: 2 4 6 20						|
															-->	| 2nd iteration: Merge 5 and 6 -> Run7 (stop)
	Run3: 1 16 19		| 1st iteration: Merge 3 and 4			|		1 2 3 4 6 16 18 19 20
	Run4: 3 18			| -> Run6: 1 3 16 18 19
*/

// Merging with 2 runs at a time (k = 2)
type SortScan struct {
	s1, s2             UpdateScan
	currentScan        UpdateScan
	comp               *RecordComparator
	hasMore1, hasMore2 bool
	savedPosition      []*record.RID
}

func NewSortScan(runs []*TempTable, comp *RecordComparator) (*SortScan, error) {
	if !(len(runs) == 1 || len(runs) == 2) {
		return nil, fmt.Errorf("runs must have 1 or 2 elements, but got %d", len(runs))
	}

	s1, err := runs[0].Open()
	if err != nil {
		return nil, fmt.Errorf("runs[0].Open: %w", err)
	}

	hasMore1 := s1.Next()

	var s2 UpdateScan
	var hasMore2 bool

	if len(runs) > 1 {
		s2, err = runs[1].Open()
		if err != nil {
			return nil, fmt.Errorf("runs[1].Open: %w", err)
		}

		hasMore2 = s2.Next()
	}

	return &SortScan{
		s1:          s1,
		s2:          s2,
		currentScan: nil,
		hasMore1:    hasMore1,
		hasMore2:    hasMore2,
		comp:        comp,
	}, nil
}

func (ss *SortScan) BeforeFirst() error {
	var err error
	if err = ss.s1.BeforeFirst(); err != nil {
		return fmt.Errorf("s1.BeforeFirst: %w", err)
	}

	ss.hasMore1 = ss.s1.Next()

	if ss.s2 != nil {
		if err = ss.s2.BeforeFirst(); err != nil {
			return fmt.Errorf("s2.BeforeFirst: %w", err)
		}

		ss.hasMore2 = ss.s2.Next()
	}

	return nil
}

func (ss *SortScan) Next() bool {
	if ss.currentScan == ss.s1 {
		ss.hasMore1 = ss.s1.Next()
	} else if ss.s2 != nil && ss.currentScan == ss.s2 {
		ss.hasMore2 = ss.s2.Next()
	}

	if !ss.hasMore1 && !ss.hasMore2 {
		return false
	} else if ss.hasMore1 && ss.hasMore2 {
		cmp, err := ss.comp.Compare(ss.s1, ss.s2)
		if err != nil {
			panic(err)
		}

		if cmp < 0 {
			ss.currentScan = ss.s1
		} else {
			ss.currentScan = ss.s2
		}
	} else if ss.hasMore1 {
		ss.currentScan = ss.s1
	} else if ss.hasMore2 {
		ss.currentScan = ss.s2
	}

	return true
}

func (ss *SortScan) GetInt(fieldName string) (int32, error) {
	return ss.currentScan.GetInt(fieldName)
}

func (ss *SortScan) GetString(fieldName string) (string, error) {
	return ss.currentScan.GetString(fieldName)
}

func (ss *SortScan) GetVal(fieldName string) (*record.Constant, error) {
	return ss.currentScan.GetVal(fieldName)
}

func (ss *SortScan) HasField(fieldName string) bool {
	return ss.currentScan.HasField(fieldName)
}

func (ss *SortScan) Close() {
	ss.s1.Close()
	if ss.s2 != nil {
		ss.s2.Close()
	}
}

func (ss *SortScan) SavePosition() error {
	rid1 := ss.s1.GetRID()

	if ss.s2 != nil {
		rid2 := ss.s2.GetRID()
		ss.savedPosition = []*record.RID{rid1, rid2}
	} else {
		ss.savedPosition = []*record.RID{rid1}
	}

	return nil
}

func (ss *SortScan) RestorePosition() error {
	rid1 := ss.savedPosition[0]
	ss.s1.MoveToRID(rid1)

	if len(ss.savedPosition) > 1 {
		rid2 := ss.savedPosition[1]
		ss.s2.MoveToRID(rid2)
	}

	return nil
}
