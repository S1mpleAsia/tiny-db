package recovery

import (
	"fmt"

	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/log"
)

type LogRecordType int32

const (
	CHECKPOINT LogRecordType = iota
	START
	COMMIT
	ROLLBACK
	SETINT
	SETSTRING
)

type LogRecord interface {
	Op() LogRecordType
	TxNumber() int32
	Undo(tx Transaction) error
}

func NewLogRecord(bytes []byte) (LogRecord, error) {
	p := file.NewPageWith(bytes)

	switch LogRecordType(p.GetInt(0)) {
	case CHECKPOINT:
		return newCheckpointRecord(), nil
	case START:
		return newStartRecordFrom(p), nil
	case COMMIT:
		return newCommitRecordFrom(p), nil
	case ROLLBACK:
		return newRollbackRecordFrom(p), nil
	case SETINT:
		return newSetIntRecordFrom(p), nil
	case SETSTRING:
		return newSetStringRecordFrom(p), nil
	default:
		return nil, fmt.Errorf("unknown RecordType: %v", p.GetInt(0))
	}
}

/*
--- CHECKPOINT RECORD ---
Record format: <CHECKPOINT>

+----------+------------+
|	 4B	   |	 4B	    |
+----------+------------+

	START	  	txNum
*/
type checkpointRecord struct{}

func newCheckpointRecord() *checkpointRecord {
	return &checkpointRecord{}
}

func (r *checkpointRecord) Op() LogRecordType {
	return CHECKPOINT
}

func (r *checkpointRecord) TxNumber() int32 {
	return 0
}

func (r *checkpointRecord) Undo(tx Transaction) error {
	return nil
}

func (r *checkpointRecord) String() string {
	return "<CHECKPOINT>"
}

func (r *checkpointRecord) WriteToLog(lm *log.LogMgmt) (int, error) {
	tpos := file.INT_32_BITS
	buf := make([]byte, tpos)
	p := file.NewPageWith(buf)
	p.SetInt(0, int32(CHECKPOINT))
	return lm.Append(buf)
}

/*
--- START RECORD ---
Record format: <START, [txNum]>

+----------+------------+
|	 4B	   |	 4B	    |
+----------+------------+

	START	  	txNum
*/
type startRecord struct {
	txNum int32
}

func newStartRecord(txNum int32) *startRecord {
	return &startRecord{
		txNum: txNum,
	}
}

func newStartRecordFrom(p *file.Page) *startRecord {
	return newStartRecord(p.GetInt(file.INT_32_BITS))
}

func (r *startRecord) Op() LogRecordType {
	return START
}

func (r *startRecord) TxNumber() int32 {
	return r.txNum
}

func (r *startRecord) String() string {
	return fmt.Sprintf("<START %d>", r.txNum)
}

func (r *startRecord) Undo(tx Transaction) error {
	return nil
}

func (r *startRecord) WriteToLog(lm *log.LogMgmt) (int, error) {
	tpos := file.INT_32_BITS
	recordLen := tpos + file.INT_32_BITS
	buf := make([]byte, recordLen)

	p := file.NewPageWith(buf)
	p.SetInt(0, int32(START))
	p.SetInt(tpos, r.txNum)
	return lm.Append(buf)
}

/*
	--- COMMIT RECORD ---
	Record format: <COMMIT, [txNum]>

	+----------+------------+
	|	 4B	   |	 4B	    |
	+----------+------------+
		START	  	txNum
*/

type commitRecord struct {
	txNum int32
}

func newCommitRecord(txNum int32) *commitRecord {
	return &commitRecord{
		txNum: txNum,
	}
}

func newCommitRecordFrom(p *file.Page) *commitRecord {
	return newCommitRecord(p.GetInt(file.INT_32_BITS))
}

func (r *commitRecord) Op() LogRecordType {
	return COMMIT
}

func (r *commitRecord) TxNumber() int32 {
	return r.txNum
}

func (r *commitRecord) String() string {
	return fmt.Sprintf("<COMMIT %d>", r.txNum)
}

func (r *commitRecord) Undo(tx Transaction) error {
	return nil
}

func (r *commitRecord) WriteToLog(lm *log.LogMgmt) (int, error) {
	tpos := file.INT_32_BITS
	recordLen := tpos + file.INT_32_BITS
	buf := make([]byte, recordLen)

	p := file.NewPageWith(buf)
	p.SetInt(0, int32(START))
	p.SetInt(tpos, r.txNum)
	return lm.Append(buf)
}

/*
	--- ROLLBACK RECORD ---
	Record format: <ROLLBACK, [txNum]>

	+----------+------------+
	|	 4B	   |	 4B	    |
	+----------+------------+
		START	  	txNum
*/

type rollbackRecord struct {
	txNum int32
}

func newRollbackRecord(txNum int32) *rollbackRecord {
	return &rollbackRecord{
		txNum: txNum,
	}
}

func newRollbackRecordFrom(p *file.Page) *rollbackRecord {
	return newRollbackRecord(p.GetInt(file.INT_32_BITS))
}

func (r *rollbackRecord) Op() LogRecordType {
	return ROLLBACK
}

func (r *rollbackRecord) TxNumber() int32 {
	return r.txNum
}

func (r *rollbackRecord) String() string {
	return fmt.Sprintf("<ROLLBACK %d>", r.txNum)
}

func (r *rollbackRecord) Undo(tx Transaction) error {
	return nil
}

func (r *rollbackRecord) WriteToLog(lm *log.LogMgmt) (int, error) {
	tpos := file.INT_32_BITS
	recordLen := tpos + file.INT_32_BITS
	buf := make([]byte, recordLen)

	p := file.NewPageWith(buf)
	p.SetInt(0, int32(START))
	p.SetInt(tpos, r.txNum)
	return lm.Append(buf)
}

/*
	--- SETINT RECORD ---
	Record format: <SETINT, [txNum], [fileName], [block], [offset], [val]>

	+----------+------------+
	|	 4B	   |	 4B	    |
	+----------+------------+
		START	  	txNum
*/

type setIntRecord struct {
	txNum  int32
	offset int32
	val    int32
	block  *file.BlockId
}

func newSetIntRecord(txNum int32, block *file.BlockId, offset int32, val int32) *setIntRecord {
	return &setIntRecord{
		txNum:  txNum,
		offset: offset,
		val:    val,
		block:  block,
	}
}

func newSetIntRecordFrom(p *file.Page) *setIntRecord {
	tpos := file.INT_32_BITS
	txNum := p.GetInt(tpos)

	fpos := tpos + file.INT_32_BITS
	fileName := p.GetString(fpos)

	bpos := fpos + file.MaxLength(len(fileName))
	blockNum := p.GetInt(bpos)
	block := file.NewBlockId(fileName, int64(blockNum))

	opos := bpos + file.INT_32_BITS
	offset := p.GetInt(opos)

	vpos := opos + file.INT_32_BITS
	val := p.GetInt(vpos)

	return newSetIntRecord(txNum, block, offset, val)
}

func (r *setIntRecord) Op() LogRecordType {
	return SETINT
}

func (r *setIntRecord) TxNumber() int32 {
	return r.txNum
}

func (r *setIntRecord) String() string {
	return fmt.Sprintf("<SETINT %d %v %d %d>", r.txNum, r.block, r.offset, r.val)
}

func (r *setIntRecord) Undo(tx Transaction) error {
	if err := tx.Pin(r.block); err != nil {
		return fmt.Errorf("pin: %w", err)
	}

	if err := tx.SetInt(r.block, r.offset, r.val, false); err != nil {
		return fmt.Errorf("setInt: %w", err)
	}

	tx.Unpin(r.block)
	return nil
}

func (r *setIntRecord) WriteToLog(lm *log.LogMgmt) (int, error) {
	tpos := file.INT_32_BITS
	fpos := tpos + file.INT_32_BITS
	bpos := fpos + file.MaxLength(len(r.block.FileName()))
	opos := bpos + file.INT_32_BITS
	vpos := opos + file.INT_32_BITS

	recordLen := vpos + file.INT_32_BITS
	buf := make([]byte, recordLen)
	p := file.NewPageWith(buf)
	p.SetInt(0, int32(SETINT))
	p.SetInt(tpos, r.txNum)
	p.SetString(fpos, r.block.FileName())
	p.SetInt(bpos, int32(r.block.BlockNumber()))
	p.SetInt(opos, r.offset)
	p.SetInt(vpos, r.val)

	return lm.Append(buf)
}

/*
	--- SETSTRING RECORD ---
	Record format: <SETSTRING, [txNum], [fileName], [block], [offset], [val]>

	+----------+------------+
	|	 4B	   |	 4B	    |
	+----------+------------+
		START	  	txNum
*/

type setStringRecord struct {
	txNum  int32
	offset int32
	val    string
	block  *file.BlockId
}

func newSetStringRecord(txNum int32, block *file.BlockId, offset int32, val string) *setStringRecord {
	return &setStringRecord{
		txNum:  txNum,
		offset: offset,
		val:    val,
		block:  block,
	}
}

func newSetStringRecordFrom(p *file.Page) *setStringRecord {
	tpos := file.INT_32_BITS
	txNum := p.GetInt(tpos)
	fpos := tpos + file.INT_32_BITS
	fileName := p.GetString(fpos)
	bpos := fpos + file.MaxLength(len(fileName))
	blockNum := p.GetInt(bpos)
	block := file.NewBlockId(fileName, int64(blockNum))

	opos := bpos + file.INT_32_BITS
	offset := p.GetInt(opos)

	vpos := opos + file.INT_32_BITS
	val := p.GetString(vpos)

	return newSetStringRecord(txNum, block, offset, val)
}

func (r *setStringRecord) Op() LogRecordType {
	return SETSTRING
}

func (r *setStringRecord) TxNumber() int32 {
	return r.txNum
}

func (r *setStringRecord) Undo(tx Transaction) error {
	if err := tx.Pin(r.block); err != nil {
		return fmt.Errorf("pin: %w", err)
	}

	if err := tx.SetString(r.block, r.offset, r.val, false); err != nil {
		return fmt.Errorf("setString: %w", err)
	}

	tx.Unpin(r.block)
	return nil
}

func (r *setStringRecord) String() string {
	return fmt.Sprintf("<SETSTRING %d %v %d %q>", r.txNum, r.block, r.offset, r.val)
}

func (r *setStringRecord) WriteToLog(lm *log.LogMgmt) (int, error) {
	tpos := file.INT_32_BITS
	fpos := tpos + file.INT_32_BITS
	bpos := fpos + file.MaxLength(len(r.block.FileName()))
	opos := bpos + file.INT_32_BITS
	vpos := opos + file.INT_32_BITS

	recordLen := vpos + file.MaxLength(len(r.val))
	buf := make([]byte, recordLen)
	p := file.NewPageWith(buf)
	p.SetInt(0, int32(SETSTRING))
	p.SetInt(tpos, r.txNum)
	p.SetString(fpos, r.block.FileName())
	p.SetInt(bpos, int32(r.block.BlockNumber()))
	p.SetInt(opos, r.offset)
	p.SetString(vpos, r.val)

	return lm.Append(buf)
}
