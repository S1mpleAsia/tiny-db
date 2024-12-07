package log

import (
	"fmt"

	"s1mpleasia.com/tinydb/file"
)

/*
	LogIterator object allocates a page to hold the content of a log block
	-	fileMgmt: Manages file-level operations (etc. reading, writing block from disk)
	-	block: Represent current block in the file being processed
	-	page: Keep the content of a block data in memory page
	-	currentPos: Tracks the current position within the block (or page) being read
	-	boundary: Contains the offset of the most recently add record (1st 4-bytes of the page)


	Note: Appends method place log record in the page from right -> left.
	This strategy enables the log iterator to read record in reverse order by reading from left -> right
*/
type LogIterator struct {
	fileMgmt 		*file.FileMgmt
	block 			*file.BlockId
	page 			*file.Page
	currentPos 		int64
	boundary 		int64
}

func NewIterator(fileMgmt *file.FileMgmt, block *file.BlockId) *LogIterator {
	b := make([]byte, fileMgmt.BlockSize())
	page := file.NewPageWith(b)

	it := &LogIterator{
		fileMgmt: fileMgmt,
		block: block,
		page: page,
		currentPos: 0,
		boundary: 0,
	}

	it.moveToBlock(block)

	return it
}

func (it *LogIterator) moveToBlock(block *file.BlockId) {
	it.fileMgmt.Read(block, it.page)
	it.boundary = int64(it.page.GetInt(0))
	it.currentPos = it.boundary
}

/*
	+----------+---------+----------+
	|  Page 6  |  Page 5 |	Page 4  | ....
	+----------+---------+----------+
*/
func (it *LogIterator) HasNext() bool {
	return it.currentPos < it.fileMgmt.BlockSize() || it.block.BlockNumber() > 0
}

func (it *LogIterator) Next() []byte {
	if it.currentPos == it.fileMgmt.BlockSize() {
		it.block = file.NewBlockId(it.block.FileName(), it.block.BlockNumber() - 1)
		it.moveToBlock(it.block)
	}

	rec := it.page.GetBytes(int(it.currentPos))
	it.currentPos += int64(file.INT_32_BITS + len(rec))

	return rec
}

/*
	Responsible for log read/write operations
	-	fileMgmt: Manages file-level operations (etc. reading, writing block from disk)
	-	logFile: Name of the log file
	-	logPage: Pointer to a instance of Page, represents a memory buffer for handling a block data in log file
	-	currentBlock: Pointer to a instance of BlockId, identifies the current block in the log file
	-	lastestLSN: LSN that store in the page memory
	-	lastSavedLSN: LSN that was flushed to the disk
*/
type LogMgmt struct {
	fileMgmt 		*file.FileMgmt
	logFile 		string
	logPage 		*file.Page
	currentBlock 	*file.BlockId
	latestLSN 		int
	lastSavedLSN 	int
}

func NewLogMgmt(fileMgmt *file.FileMgmt, logFile string) (*LogMgmt, error) {
	b := make([]byte, fileMgmt.BlockSize())
	logPage := file.NewPageWith(b)

	logSize, err := fileMgmt.Length(logFile)
	if err != nil {
		return nil, fmt.Errorf("fileMgmt.Length: %w", err)
	}

	lm := &LogMgmt{
		fileMgmt: fileMgmt,
		logFile: logFile,
		logPage: logPage,
	}

	if logSize == 0 {
		lm.currentBlock, err = lm.appendNewBlock()
		if err != nil {
			return nil, fmt.Errorf("lm.appendNewBlock: %w", err)
		}
	} else {
		lm.currentBlock = file.NewBlockId(logFile, logSize - 1)
		fileMgmt.Read(lm.currentBlock, logPage)
	}

	return lm, nil
}

func (lm *LogMgmt) appendNewBlock() (*file.BlockId, error) {
	block, err := lm.fileMgmt.Append(lm.logFile)

	if err != nil {
		return nil, fmt.Errorf("fileMgmt.Append: %w", err)
	}

	lm.logPage.SetInt(0, int32(lm.fileMgmt.BlockSize()))
	lm.fileMgmt.Write(block, lm.logPage)
	
	return block, nil
}

func (lm *LogMgmt) Flush(lsn int) {
	if lsn < lm.lastSavedLSN {
		return
	}
	lm.flush()
}

func (lm *LogMgmt) flush() {
	lm.fileMgmt.Write(lm.currentBlock, lm.logPage)
	lm.lastSavedLSN = lm.latestLSN
}

func (lm *LogMgmt) Iterator() *LogIterator {
	lm.flush()
	return NewIterator(lm.fileMgmt, lm.currentBlock)
}

func (lm *LogMgmt) Append(record []byte) (int, error) {
	boundary := int(lm.logPage.GetInt(0))
	recordSize := len(record)
	bytesNeeded := recordSize + file.INT_32_BITS

	// If a int -> Need 4 bytes
	// Else -> Need at least 4 bytes for store the metadata
	if boundary - bytesNeeded < file.INT_32_BITS {
		lm.flush()
		currentBlock, err := lm.appendNewBlock()
		if err != nil {
			return 0, fmt.Errorf("lm.appendNewBlock: %w", err)
		}

		lm.currentBlock = currentBlock
		boundary = int(lm.logPage.GetInt(0))
	}

	// Write in reverse order
	recPos := boundary - bytesNeeded
	lm.logPage.SetBytes(recPos, record)
	lm.logPage.SetInt(0, int32(recPos))

	lm.latestLSN += 1

	return lm.latestLSN, nil
}