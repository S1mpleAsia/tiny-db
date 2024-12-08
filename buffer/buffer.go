package buffer

import (
	"fmt"
	"sync"
	"time"

	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/log"
)

/*
	A buffer object keep track of those information about its page:
	-	block: reference to the block assigned to its page. if no block assigned -> null
	-	pins: Number of time the page is pinned. The pins count is incremented on each pin and decremented on each unpin
	-	txNnum: Indicates that the page have been modified. txNum = -1 -> Not been changed, otherwise -> have been changed
	-	lsn: If the page have been modified -> the buffer holds the lsn of the most recent log record
*/
type Buffer struct {
	fileMgmt 	*file.FileMgmt
	logMgmt 	*log.LogMgmt
	contents 	*file.Page
	block 		*file.BlockId
	pins 		int
	txNum 		int
	lsn 		int
}

func NewBuffer(fm *file.FileMgmt, lm *log.LogMgmt) (*Buffer) {
	return &Buffer{
		fileMgmt: fm,
		logMgmt: lm,
		contents: file.NewPage(fm.BlockSize()),
		pins: 0,
		txNum: -1,
		lsn: -1,
	}
}

func (buffer *Buffer) Contents() *file.Page {
	return buffer.contents
}

func (buffer *Buffer) Block() *file.BlockId {
	return buffer.block
}

func (buffer *Buffer) IsPinned() bool {
	return buffer.pins > 0
}

func (buffer *Buffer) SetModified(txNum int, lsn int) {
	buffer.txNum = txNum
	if lsn >= 0 {
		buffer.lsn = lsn
	}
}

func (buffer *Buffer) ModifyingTx() int {
	return buffer.txNum
}

func (buffer *Buffer) flush() {
	if (buffer.txNum >= 0) {
		buffer.logMgmt.Flush(buffer.lsn)
		buffer.fileMgmt.Write(buffer.block, buffer.contents)
		buffer.txNum = -1
	}
}

/*
	Associates the buffer with a disk block. 
	1. The buffer is first flushed, so that any modifications to prev block are preversed
	2. The buffer then associated with the specified block, reading its contents from disk
*/
func (buffer *Buffer) assignToBlock(block *file.BlockId) {
	buffer.flush()
	buffer.block = block
	buffer.fileMgmt.Read(block, buffer.contents)
	buffer.pins = 0
}

func (buffer *Buffer) pin() {
	buffer.pins++
}

func (buffer *Buffer) unpin() {
	buffer.pins--
}

const MAX_TIME = 10 * time.Second

type BufferMgmt struct {
	bufferPool 		[]*Buffer
	numAvailable 	int
	mu				sync.Mutex
	cond			*sync.Cond
}

func NewBufferMgmt(fm *file.FileMgmt, lm *log.LogMgmt, numBuffs int) *BufferMgmt {
	bufferPool := make([]*Buffer, numBuffs)

	for	i := range numBuffs {
		bufferPool[i] = NewBuffer(fm, lm)
	}

	bm := &BufferMgmt{
		bufferPool: bufferPool,
		numAvailable: numBuffs,
	}

	bm.cond = sync.NewCond(&bm.mu)
	return bm
}

func (bm *BufferMgmt) Pin(block *file.BlockId) (*Buffer, error) {
	bm.mu.Lock()

	startTime := time.Now()
	buff := bm.tryToPin(block)

	for buff == nil && !bm.waitingTooLong(startTime) {
		remainingTime := MAX_TIME - time.Since(startTime)

		if remainingTime <= 0 {
			break
		}

		timer := time.NewTimer(remainingTime)
		done := make(chan struct{})

		go func() {
			defer close(done)
			bm.cond.Wait()
			bm.mu.Unlock()
		}()

		select {
		case <- timer.C:
			fmt.Println("Timer expired")
			// Timer expired: do nothing, loop will check conditions again.
		case <- done:
			fmt.Println("Woken up by Signal or Broadcast")
			// Woken up by Signal or Broadcast: do nothing, loop will check conditions again.
		}
		
		timer.Stop()
		bm.mu.Lock()
		buff = bm.tryToPin(block)
	}

	bm.mu.Unlock()

	if buff == nil {
		return nil, fmt.Errorf("Buffer abort exception. Deadlock")
	}

	return buff, nil
}

func (bm *BufferMgmt) Unpin(buffer *Buffer) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	buffer.unpin()
	if !buffer.IsPinned() {
		bm.numAvailable++
		bm.cond.Broadcast()
	}
}

func (bm *BufferMgmt) Available() int {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	return bm.numAvailable
}

func (bm *BufferMgmt) FlushAll(txNum int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	
	for _, buff := range bm.bufferPool {
		if buff.ModifyingTx() == txNum {
			buff.flush()
		}
	}
}

func (bm *BufferMgmt) tryToPin(block *file.BlockId) *Buffer {
	buff := bm.findExistingBuffer(block)

	if buff == nil {
		buff = bm.chooseUnpinnedBuffer()
		if buff == nil {
			return nil
		}

		buff.assignToBlock(block)
	}

	if !buff.IsPinned() {
		bm.numAvailable--;
	}

	buff.pin()
	return buff
}

func (bm *BufferMgmt) waitingTooLong(startTime time.Time) bool {
	return time.Since(startTime) > MAX_TIME
}

func (bm *BufferMgmt) findExistingBuffer(block *file.BlockId) *Buffer {
	for _, buff := range bm.bufferPool {
		if buff.block != nil && buff.block == block {
			return buff
		}
	}

	return nil
}

// Naive algorithm
func (bm *BufferMgmt) chooseUnpinnedBuffer() *Buffer {
	for _, buff := range bm.bufferPool {
		if !buff.IsPinned() {
			return buff
		}
	}

	return nil
}
