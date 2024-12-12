package recovery

import (
	"fmt"

	"s1mpleasia.com/tinydb/buffer"
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/log"
)

type Transaction interface {
	Pin(block *file.BlockId) error
	SetString(block *file.BlockId, offset int32, val string, logRecord bool) error
	SetInt(block *file.BlockId, offset int32, val int32, logRecord bool) error
	Unpin(block *file.BlockId)
}

// Undo-only recovery with value-granularity data items
type RecoveryMgmt struct {
	lm    *log.LogMgmt
	bm    *buffer.BufferMgmt
	tx    Transaction
	txNum int32
}

func NewRecoveryMgmt(tx Transaction, txNum int32, lm *log.LogMgmt, bm *buffer.BufferMgmt) (*RecoveryMgmt, error) {
	if _, err := newStartRecord(txNum).WriteToLog(lm); err != nil {
		return nil, err
	}

	return &RecoveryMgmt{
		lm:    lm,
		bm:    bm,
		tx:    tx,
		txNum: txNum,
	}, nil
}

func (recoveryMgmt *RecoveryMgmt) Commit() error {
	// Write-ahead log in this method
	recoveryMgmt.bm.FlushAll(int(recoveryMgmt.txNum))

	lsn, err := newCommitRecord(recoveryMgmt.txNum).WriteToLog(recoveryMgmt.lm)
	if err != nil {
		return err
	}

	recoveryMgmt.lm.Flush(lsn)
	return nil
}

func (recoveryMgmt *RecoveryMgmt) Rollback() error {
	err := recoveryMgmt.doRollback()
	if err != nil {
		return err
	}

	recoveryMgmt.bm.FlushAll(int(recoveryMgmt.txNum))

	lsn, err := newRollbackRecord(recoveryMgmt.txNum).WriteToLog(recoveryMgmt.lm)
	if err != nil {
		return err
	}

	recoveryMgmt.lm.Flush(lsn)
	return nil
}

func (recoveryMgmt *RecoveryMgmt) Recover() error {
	err := recoveryMgmt.doRecover()
	if err != nil {
		return err
	}

	recoveryMgmt.bm.FlushAll(int(recoveryMgmt.txNum))

	lsn, err := newCheckpointRecord().WriteToLog(recoveryMgmt.lm)
	if err != nil {
		return err
	}

	recoveryMgmt.lm.Flush(lsn)
	return nil
}

func (recoveryMgmt *RecoveryMgmt) SetInt(buff *buffer.Buffer, offset int32, newVal int32) (int, error) {
	oldVal := buff.Contents().GetInt(int(offset))
	blk := buff.Block()

	return newSetIntRecord(recoveryMgmt.txNum, blk, offset, oldVal).WriteToLog(recoveryMgmt.lm)
}

func (recoveryMgmt *RecoveryMgmt) SetString(buff *buffer.Buffer, offset int32, newVal string) (int, error) {
	oldVal := buff.Contents().GetString(int(offset))
	blk := buff.Block()
	return newSetStringRecord(recoveryMgmt.txNum, blk, offset, oldVal).WriteToLog(recoveryMgmt.lm)
}

func (recoveryMgmt *RecoveryMgmt) doRollback() error {
	it := recoveryMgmt.lm.Iterator()

	for it.HasNext() {
		bytes := it.Next()

		rec, err := NewLogRecord(bytes)
		if err != nil {
			return fmt.Errorf("recovery.doRollback for %s: %w", string(bytes), err)
		}

		if rec.TxNumber() == recoveryMgmt.txNum {
			if rec.Op() == START {
				return nil
			}

			if err := rec.Undo(recoveryMgmt.tx); err != nil {
				return fmt.Errorf("Undo: %w", err)
			}
		}
	}

	return nil
}

// Quiescent checkpoint
func (recoveryMgmt *RecoveryMgmt) doRecover() error {
	finishedTx := make(map[int32]struct{})

	it := recoveryMgmt.lm.Iterator()

	for it.HasNext() {
		bytes := it.Next()

		rec, err := NewLogRecord(bytes)
		if err != nil {
			return fmt.Errorf("recovery.doRecover for %s: %w", string(bytes), err)
		}

		if rec.Op() == CHECKPOINT {
			return nil
		} else if rec.Op() == COMMIT || rec.Op() == ROLLBACK {
			finishedTx[rec.TxNumber()] = struct{}{}
		} else if _, ok := finishedTx[recoveryMgmt.txNum]; !ok {
			if err := rec.Undo(recoveryMgmt.tx); err != nil {
				return fmt.Errorf("Undo: %w", err)
			}
		}
	}

	return nil
}
