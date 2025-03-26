package query

import "s1mpleasia.com/tinydb/record"

type IndexJoinScan struct {
	lhs       Scan
	idx       Index
	joinField string
	rhs       Scan // Using index - Select scan
}

var _ Scan = (*IndexJoinScan)(nil)

func NewIndexJoinScan(lhs Scan, idx Index, joinField string, rhs Scan) (*IndexJoinScan, error) {
	s := &IndexJoinScan{lhs, idx, joinField, rhs}
	if err := s.BeforeFirst(); err != nil {
		return nil, err
	}

	return s, nil
}

func (i *IndexJoinScan) BeforeFirst() error {
	if err := i.lhs.BeforeFirst(); err != nil {
		return err
	}

	i.lhs.Next()
	return i.resetIndex()
}

func (i *IndexJoinScan) Next() bool {
	for {
		ok, err := i.idx.Next()
		if err != nil {
			return false
		}

		if ok {
			rid, err := i.idx.GetDataRID()
			if err != nil {
				return false
			}

			us := i.rhs.(UpdateScan)
			us.MoveToRID(rid)
			return true
		} else if !ok {
			lNext := i.lhs.Next()
			if !lNext {
				return false
			}

			if err := i.resetIndex(); err != nil {
				return false
			}
		}
	}
}

func (i *IndexJoinScan) GetInt(fieldName string) (int32, error) {
	if i.rhs.HasField(fieldName) {
		return i.rhs.GetInt(fieldName)
	} else {
		return i.lhs.GetInt(fieldName)
	}
}

func (i *IndexJoinScan) GetString(fieldName string) (string, error) {
	if i.rhs.HasField(fieldName) {
		return i.rhs.GetString(fieldName)
	} else {
		return i.lhs.GetString(fieldName)
	}
}

func (i *IndexJoinScan) GetVal(fieldName string) (*record.Constant, error) {
	if i.rhs.HasField(fieldName) {
		return i.rhs.GetVal(fieldName)
	} else {
		return i.lhs.GetVal(fieldName)
	}
}

func (i *IndexJoinScan) HasField(fieldName string) bool {
	return i.rhs.HasField(fieldName) || i.lhs.HasField(fieldName)
}

func (i *IndexJoinScan) Close() {
	i.lhs.Close()
	err := i.idx.Close()
	if err != nil {
		panic(err)
	}
	i.rhs.Close()
}

func (i *IndexJoinScan) resetIndex() error {
	searchKey, err := i.lhs.GetVal(i.joinField)
	if err != nil {
		return err
	}

	return i.idx.BeforeFirst(searchKey)
}
