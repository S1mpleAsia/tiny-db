# recovery — Undo-based recovery management

Implements write-ahead logging and undo-only recovery for a transaction, using value-granularity log records that store the before-image of each update.

## Responsibilities

- Write a start record when a transaction begins and commit/rollback records at its end.
- Record before-images of integer and string updates as log records.
- Roll back a single transaction by undoing its logged changes.
- Recover the database after a crash by undoing changes of unfinished transactions and writing a checkpoint.
- Serialize and deserialize log records to and from the log.

## Key types

- `RecoveryMgmt` — per-transaction recovery manager holding the log manager, buffer manager, owning transaction, and transaction number.
- `Transaction` — interface the recovery manager needs from a transaction (`Pin`, `Unpin`, `SetInt`, `SetString`) to apply undo operations.
- `LogRecord` — interface for all log records: `Op()`, `TxNumber()`, `Undo(tx Transaction) error`.
- `LogRecordType` — enum of record kinds: `CHECKPOINT`, `START`, `COMMIT`, `ROLLBACK`, `SETINT`, `SETSTRING`.
- Record structs `checkpointRecord`, `startRecord`, `commitRecord`, `rollbackRecord`, `setIntRecord`, `setStringRecord` — one per record kind (unexported), each with a `WriteToLog` method.

## Key API

- `NewRecoveryMgmt(tx Transaction, txNum int32, lm *log.LogMgmt, bm *buffer.BufferMgmt) (*RecoveryMgmt, error)` — creates the manager and appends a start record.
- `(*RecoveryMgmt) Commit() error` — flushes all buffers, then writes and flushes a commit record.
- `(*RecoveryMgmt) Rollback() error` — undoes the transaction's changes, flushes buffers, and writes a rollback record.
- `(*RecoveryMgmt) Recover() error` — undoes changes of all unfinished transactions, flushes buffers, and writes a checkpoint.
- `(*RecoveryMgmt) SetInt(buff *buffer.Buffer, offset int32, newVal int32) (int, error)` — appends a set-int record capturing the old value; returns the LSN.
- `(*RecoveryMgmt) SetString(buff *buffer.Buffer, offset int32, newVal string) (int, error)` — appends a set-string record capturing the old value; returns the LSN.
- `NewLogRecord(bytes []byte) (LogRecord, error)` — reconstructs a typed log record from its serialized bytes.

## How it fits

Depends on `buffer`, `file`, and `log`. Used by the `transaction` package, which owns one `RecoveryMgmt` per transaction and delegates commit, rollback, and recovery to it.

## Notes

- Recovery is undo-only: set records store the before-image, so redo is not performed. Commit follows the write-ahead protocol by flushing buffers before logging the commit.
- `doRollback` and `doRecover` scan the log backward via the log iterator; recovery uses a quiescent checkpoint and stops at the first checkpoint record.
- `Undo` is a no-op for checkpoint, start, commit, and rollback records; only set records re-apply their old value.
- The commit and rollback records currently serialize their type field as `START` in `WriteToLog`; deserialization dispatches on the first stored integer.
- No test files are present in this package.
