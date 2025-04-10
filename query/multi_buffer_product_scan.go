package query

import (
	"fmt"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/transaction"
)

var _ Scan = (*MultiBufferProductScan)(nil)

type MultiBufferProductScan struct {
	tx         *transaction.Transaction
	lhs        Scan
	rhs        *ChunkScan
	prod       *ProductScan
	fileName   string
	layout     *record.Layout
	chunkSize  int
	nextBlkNum int32
	fileSize   int // Number of blocks in the file
}

func NewMultiBufferProductScan(
	tx *transaction.Transaction,
	lhs Scan,
	tableName string,
	layout *record.Layout) (*MultiBufferProductScan, error) {

	fileName := tableName + ".tbl"
	fileSize := tx.Size(fileName)

	available := tx.AvailableBuffs()
	chunkSize := BufferNeedsBestFactor(available, fileSize)

	return &MultiBufferProductScan{
		tx:        tx,
		lhs:       lhs,
		fileName:  fileName,
		layout:    layout,
		chunkSize: chunkSize,
		fileSize:  fileSize,
	}, nil
}

func (m *MultiBufferProductScan) BeforeFirst() error {
	m.nextBlkNum = 0
	_, err := m.useNextChunk()
	if err != nil {
		return err
	}

	return nil
}

/*
*
Moves to the next record in the current scan.
If there are no more records in the current chunk,
then move to the next LHS record and the beginning of that chunk.
If there are no more LHS records, then move to the next chunk
and begin again.
@see tinyDB.query.Scan#next()
*/
func (m *MultiBufferProductScan) Next() bool {
	for {
		next := m.prod.Next()
		if next {
			return true
		}

		ok, err := m.useNextChunk()
		if err != nil {
			return false
		}
		if !ok {
			return false
		}
	}
}

func (m *MultiBufferProductScan) GetInt(fieldName string) (int32, error) {
	return m.prod.GetInt(fieldName)
}

func (m *MultiBufferProductScan) GetString(fieldName string) (string, error) {
	return m.prod.GetString(fieldName)
}

func (m *MultiBufferProductScan) GetVal(fieldName string) (*record.Constant, error) {
	return m.prod.GetVal(fieldName)
}

func (m *MultiBufferProductScan) HasField(fieldName string) bool {
	return m.prod.HasField(fieldName)
}

func (m *MultiBufferProductScan) Close() {
	m.prod.Close()
}

func (m *MultiBufferProductScan) useNextChunk() (bool, error) {
	if int(m.nextBlkNum) >= m.fileSize {
		fmt.Printf("useNextChunk(): nextBlkNum=%d >= fileSize=%d\n", m.nextBlkNum, m.fileSize)
		return false, nil
	}

	if m.rhs != nil {
		m.rhs.Close()
	}

	end := int(m.nextBlkNum) + m.chunkSize - 1
	if end >= m.fileSize {
		end = m.fileSize - 1
	}

	rhs, err := NewChunkScan(m.tx, m.fileName, m.layout, int64(m.nextBlkNum), int64(end))
	if err != nil {
		return false, err
	}

	if err = m.lhs.BeforeFirst(); err != nil {
		return false, fmt.Errorf("m.lhs.BeforeFirst: %w", err)
	}

	m.prod, err = NewProductScan(m.lhs, rhs)
	if err != nil {
		return false, fmt.Errorf("NewProductScan: %w", err)
	}

	m.nextBlkNum = int32(end + 1)
	return true, nil
}
