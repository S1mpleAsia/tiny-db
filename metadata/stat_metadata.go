package metadata

import (
	"fmt"
	"sync"

	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

/*
For simplicity, statistic metadata will only following 3 kinds of statistical attribute:
-	The number of blocks used by each table T
-	The number of records in each table T
-	For each field F of table T, the number of distinct F-values in T

** Note: V(T, F) is assumed to be approximately 1/3 of the values of any fields are distinct => Need to be rectified
*/
type StatInfo struct {
	numBlocks  int32
	numRecords int32
}

func NewStatInfo(numBlocks int32, numRecords int32) *StatInfo {
	return &StatInfo{
		numBlocks:  numBlocks,
		numRecords: numRecords,
	}
}

func (si *StatInfo) BlockAccessed() int32 {
	return si.numBlocks
}

func (si *StatInfo) RecordsOutput() int32 {
	return si.numRecords
}

func (si *StatInfo) DistinctValues(fieldName string) int32 {
	return 1 + (si.numRecords / 3)
}

type StatMgmt struct {
	tblMgmt  *TableMgmt
	tblStat  map[string]*StatInfo
	numCalls int
	mux      *sync.Mutex
}

func NewStatMgmt(tblMgmt *TableMgmt, tx *transaction.Transaction) (*StatMgmt, error) {
	statMgmt := &StatMgmt{
		tblMgmt:  tblMgmt,
		tblStat:  nil,
		numCalls: 0,
		mux:      &sync.Mutex{},
	}

	err := statMgmt.refreshStatistics(tx)
	if err != nil {
		return nil, err
	}

	return statMgmt, nil
}

func (sm *StatMgmt) GetStatInfo(tblName string, layout *record.Layout, tx *transaction.Transaction) (*StatInfo, error) {
	sm.mux.Lock()
	defer sm.mux.Unlock()

	sm.numCalls++
	if sm.numCalls > 100 {
		err := sm.refreshStatistics(tx)
		if err != nil {
			return nil, err
		}
	}

	si := sm.tblStat[tblName]
	if si == nil {
		si, err := sm.calcTableStats(tblName, layout, tx)
		if err != nil {
			return nil, err
		}

		sm.tblStat[tblName] = si
	}

	return sm.tblStat[tblName], nil
}

func (sm *StatMgmt) ForceRefreshStatistics(tx *transaction.Transaction) error {
	return sm.refreshStatistics(tx)
}

func (sm *StatMgmt) refreshStatistics(tx *transaction.Transaction) error {
	fmt.Println("refresh statistics")
	sm.tblStat = make(map[string]*StatInfo)
	sm.numCalls = 0

	tblCatalogLayout, err := sm.tblMgmt.GetLayout(TBL_CATALOG_FILE, tx)
	if err != nil {
		return err
	}

	tblCatalog, err := record.NewTableScan(tx, TBL_CATALOG_FILE, tblCatalogLayout)
	if err != nil {
		return err
	}

	defer tblCatalog.Close()

	for tblCatalog.Next() {
		tableName, err := tblCatalog.GetString(TBL_CATALOG_TABLE_NAME)
		if err != nil {
			return err
		}

		layout, err := sm.tblMgmt.GetLayout(tableName, tx)
		if err != nil {
			return err
		}

		si, err := sm.calcTableStats(tableName, layout, tx)
		if err != nil {
			return err
		}
		sm.tblStat[tableName] = si
	}

	return nil
}

func (sm *StatMgmt) calcTableStats(tblName string, layout *record.Layout, tx *transaction.Transaction) (*StatInfo, error) {
	numRecs := int32(0)
	numBlocks := int32(0)

	ts, err := record.NewTableScan(tx, tblName, layout)
	if err != nil {
		return nil, err
	}
	defer ts.Close()

	for ts.Next() {
		numRecs++
		rid := ts.GetRID()

		numBlocks = rid.BlockNumber() + 1
	}

	fmt.Printf("%q calcTableStats: numRecs=%d, numBlocks=%d\n", tblName, numRecs, numBlocks)
	return NewStatInfo(numBlocks, numRecs), nil
}
