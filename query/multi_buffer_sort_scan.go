package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
)

var _ Scan = (*MultiBufferSortScan)(nil)

type MultiBufferSortScan struct {
	scans         []UpdateScan
	hasMores      []bool
	currentScan   UpdateScan
	comp          *RecordComparator
	savedPosition []*record.RID
}

func NewMultiBufferSortScan(runs []*TempTable, comp *RecordComparator) (*MultiBufferSortScan, error) {
	scans := make([]UpdateScan, 0, len(runs))
	hasMores := make([]bool, 0, len(runs))

	for i, run := range runs {
		s, err := run.Open()
		if err != nil {
			return nil, fmt.Errorf("run[%d].Open: %v", i, err)
		}

		scans = append(scans, s)
		hasMore := s.Next()
		hasMores = append(hasMores, hasMore)
	}

	return &MultiBufferSortScan{
		scans:       scans,
		hasMores:    hasMores,
		currentScan: nil,
		comp:        comp,
	}, nil
}

func (m *MultiBufferSortScan) BeforeFirst() error {
	for i, scan := range m.scans {
		if err := scan.BeforeFirst(); err != nil {
			return fmt.Errorf("scan[%d].BeforeFirst: %v", i, err)
		}
		hasMore := scan.Next()
		m.hasMores[i] = hasMore
	}

	return nil
}

func (m *MultiBufferSortScan) Next() bool {
	scanMap := make(map[int]Scan)
	for i, scan := range m.scans {
		if m.currentScan == scan {
			hasMore := scan.Next()
			m.hasMores[i] = hasMore
		}
		fmt.Printf("Next(): scans[%d]: hasMore=%t\n", i, m.hasMores[i])

		if !m.hasMores[i] {
			break
		}

		scanMap[i] = scan
	}

	fmt.Printf("Next(): scanMap=%+v\n", scanMap)

	if len(scanMap) == 0 {
		return false
	}

	minIdx := -1
	for i, scan := range scanMap {
		if minIdx == -1 {
			minIdx = i
			continue
		}

		cmp, err := m.comp.Compare(scan, scanMap[minIdx])
		if err != nil {
			panic(err)
		}

		fmt.Printf("Next(): Compare(%+v, %+v) = %d\n", scan, scanMap[minIdx], cmp)
		if cmp < 0 {
			minIdx = i
		}
	}

	fmt.Printf("Next(): m.CurrentScan= scan[%v]\n", minIdx)
	return true
}

func (m *MultiBufferSortScan) GetInt(fieldName string) (int32, error) {
	return m.currentScan.GetInt(fieldName)
}

func (m *MultiBufferSortScan) GetString(fieldName string) (string, error) {
	return m.currentScan.GetString(fieldName)
}

func (m *MultiBufferSortScan) GetVal(fieldName string) (*record.Constant, error) {
	return m.currentScan.GetVal(fieldName)
}

func (m *MultiBufferSortScan) HasField(fieldName string) bool {
	return m.currentScan.HasField(fieldName)
}

func (m *MultiBufferSortScan) Close() {
	for _, scan := range m.scans {
		scan.Close()
	}
}

func (m *MultiBufferSortScan) SavePosition() error {
	savedPosition := make([]*record.RID, 0, len(m.scans))
	for _, scan := range m.scans {
		rid := scan.GetRID()
		savedPosition = append(savedPosition, rid)
	}

	m.savedPosition = savedPosition
	return nil
}

func (m *MultiBufferSortScan) RestorePosition() error {
	for i, scan := range m.scans {
		scan.MoveToRID(m.savedPosition[i])
	}
	return nil
}
