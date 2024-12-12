package transaction

import (
	"fmt"
	"slices"
	"sync"

	"s1mpleasia.com/tinydb/buffer"
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/log"
	"s1mpleasia.com/tinydb/transaction/concurrency"
	"s1mpleasia.com/tinydb/transaction/recovery"
)

const endOfFile = -1

var (
	txMutex         = &sync.Mutex{}
	nextTxNum int32 = 0
)

type Transaction struct {
	recoveryMgmt *recovery.RecoveryMgmt
	concurMgmt   *concurrency.ConcurrencyManagement
	bm           *buffer.BufferMgmt
	fm           *file.FileMgmt
	txNum        int32
	myBuffers    BufferList
}

func NewTransaction(fm *file.FileMgmt, lm *log.LogMgmt, bm *buffer.BufferMgmt) (*Transaction, error) {
	tx := &Transaction{
		concurMgmt: concurrency.NewConcurrencyMgmt(),
		fm:         fm,
		bm:         bm,
		txNum:      nextTxNumber(),
		myBuffers:  *newBufferList(bm),
	}

	rm, err := recovery.NewRecoveryMgmt(tx, tx.txNum, lm, bm)
	if err != nil {
		return nil, err
	}

	tx.recoveryMgmt = rm
	return tx, nil
}

// Transaction's lifespan
func (tx *Transaction) Commit() {
	fmt.Printf("transaction %d committing...\n", tx.txNum)
	if err := tx.recoveryMgmt.Commit(); err != nil {
		panic(err)
	}

	tx.concurMgmt.Release()
	tx.myBuffers.unpinAll()
	fmt.Printf("transaction %d committed\n", tx.txNum)
}

func (tx *Transaction) Rollback() {
	fmt.Printf("transaction %d rolling back...\n", tx.txNum)
	if err := tx.recoveryMgmt.Rollback(); err != nil {
		panic(err)
	}
	tx.concurMgmt.Release()
	tx.myBuffers.unpinAll()

	fmt.Printf("transaction %d rolled back\n", tx.txNum)
}

func (tx *Transaction) Recover() {
	tx.bm.FlushAll(int(tx.txNum))

	if err := tx.recoveryMgmt.Recover(); err != nil {
		panic(err)
	}

}

// Method for access buffers
func (tx *Transaction) Pin(block *file.BlockId) error {
	tx.myBuffers.pin(block)
	fmt.Printf("(%q) Pin(%+v)", block.FileName(), block)

	return nil
}

func (tx *Transaction) Unpin(block *file.BlockId) {
	tx.myBuffers.unpin(block)
	fmt.Printf("(%q) Unpin(%+v)", block.FileName(), block)
}

func (tx *Transaction) GetInt(block *file.BlockId, offset int) int32 {
	err := tx.concurMgmt.SLock(*block)
	if err != nil {
		panic(err)
	}

	buff := tx.myBuffers.buffers[block]
	return buff.Contents().GetInt(offset)
}

func (tx *Transaction) GetString(block *file.BlockId, offset int) string {
	err := tx.concurMgmt.SLock(*block)
	if err != nil {
		panic(err)
	}

	buff := tx.myBuffers.buffers[block]
	return buff.Contents().GetString(offset)
}

func (tx *Transaction) SetInt(block *file.BlockId, offset int32, val int32, okToLog bool) error {
	err := tx.concurMgmt.XLock(*block)

	if err != nil {
		return err
	}

	buff := tx.myBuffers.buffers[block]
	var lsn int = -1
	if okToLog {
		lsn, err = tx.recoveryMgmt.SetInt(buff, offset, val)
		if err != nil {
			return err
		}
	}

	p := buff.Contents()
	p.SetInt(int(offset), val)
	buff.SetModified(int(tx.txNum), lsn)
	return nil
}

func (tx *Transaction) SetString(block *file.BlockId, offset int32, val string, okToLog bool) error {
	err := tx.concurMgmt.XLock(*block)
	if err != nil {
		return err
	}

	buff := tx.myBuffers.buffers[block]
	var lsn int = -1
	if okToLog {
		lsn, err = tx.recoveryMgmt.SetString(buff, offset, val)
		if err != nil {
			return err
		}
	}

	p := buff.Contents()
	p.SetString(int(offset), val)
	buff.SetModified(int(tx.txNum), lsn)

	return nil
}

func (tx *Transaction) AvailableBuffs() int {
	return tx.bm.Available()
}

// Method related to the file mgmt
func (tx *Transaction) Size(filename string) int {
	dummyBlock := file.NewBlockId(filename, endOfFile)
	if err := tx.concurMgmt.SLock(*dummyBlock); err != nil {
		panic(err)
	}

	len, err := tx.fm.Length(filename)
	if err != nil {
		panic(err)
	}

	return int(len)
}

func (tx *Transaction) Append(filename string) *file.BlockId {
	dummyBlock := file.NewBlockId(filename, endOfFile)
	if err := tx.concurMgmt.XLock(*dummyBlock); err != nil {
		panic(err)
	}

	block, err := tx.fm.Append(filename)
	if err != nil {
		panic(err)
	}

	fmt.Printf("(%q) wrote block from append %+v", block.FileName(), block)
	return block
}

func (tx *Transaction) BlockSize() int {
	return int(tx.fm.BlockSize())
}

func nextTxNumber() int32 {
	txMutex.Lock()
	defer txMutex.Unlock()

	nextTxNum++
	fmt.Printf("New transaction: %d\n", nextTxNum)
	return nextTxNum
}

type BufferList struct {
	buffers map[*file.BlockId]*buffer.Buffer
	pins    []*file.BlockId
	bm      *buffer.BufferMgmt
}

func newBufferList(bm *buffer.BufferMgmt) *BufferList {
	return &BufferList{
		buffers: make(map[*file.BlockId]*buffer.Buffer),
		pins:    make([]*file.BlockId, 0),
		bm:      bm,
	}
}

func (b *BufferList) pin(block *file.BlockId) {
	buf, err := b.bm.Pin(block)
	if err != nil {
		panic(fmt.Sprintf("block %v not pinned", block))
	}

	b.buffers[block] = buf
	b.pins = append(b.pins, block)
}

func (b *BufferList) unpin(block *file.BlockId) {
	buf, ok := b.buffers[block]
	if !ok {
		panic(fmt.Sprintf("block %v not unpinned", block))
	}

	b.bm.Unpin(buf)
	for i, p := range b.pins {
		if p == block {
			b.pins = slices.Delete(b.pins, i, i+1)
			break
		}
	}

	if !slices.Contains(b.pins, block) {
		delete(b.buffers, block)
	}
}

func (b *BufferList) unpinAll() {
	for _, block := range b.pins {
		buf, ok := b.buffers[block]

		if !ok {
			panic(fmt.Sprintf("block %+v not pinned", block))
		}
		b.bm.Unpin(buf)
	}

	b.buffers = make(map[*file.BlockId]*buffer.Buffer)
	b.pins = make([]*file.BlockId, 0)
}
