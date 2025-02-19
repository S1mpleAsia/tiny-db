package index

import "s1mpleasia.com/tinydb/record"

type Index interface {
	// Move the index pointer to the desired record
	BeforeFirst(searchKey *record.Constant) error

	// Get the data point by index. Return false if there is no more data to get
	Next() (bool, error)

	// Get the actual data by RID value
	GetDataRID() (*record.RID, error)

	Insert(dataval *record.Constant, datarid *record.RID) error
	Delete(dataval *record.Constant, datarid *record.RID) error
	Close() error
}
