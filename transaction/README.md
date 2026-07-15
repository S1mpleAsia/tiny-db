# transaction — Transaction management

Coordinates atomic, isolated access to disk blocks for a single logical unit of work by combining recovery and concurrency control on top of the buffer, file, and log managers.

## Responsibilities

- Assign a unique, monotonically increasing transaction number to each transaction.
- Provide value-level read/write access (`GetInt`/`GetString`, `SetInt`/`SetString`) to buffered blocks.
- Acquire shared/exclusive locks before reads/writes to enforce isolation.
- Log before-images of updates and drive commit, rollback, and recovery through the recovery manager.
- Track and pin/unpin the buffers used by the transaction, releasing them at commit/rollback.
- Expose file-level operations (block size, file length, block append) under appropriate locks.

## Key types

- `Transaction` — a single transaction; owns its recovery manager, concurrency manager, buffer list, and transaction number.
- `BufferList` — tracks the buffers pinned by a transaction and their pin counts, delegating to the buffer manager.

## Key API

- `NewTransaction(fm *file.FileMgmt, lm *log.LogMgmt, bm *buffer.BufferMgmt) (*Transaction, error)` — creates a transaction, assigns its number, and writes a start record.
- `(*Transaction) Commit()` — flushes buffers, writes a commit record, releases locks, and unpins buffers.
- `(*Transaction) Rollback()` — undoes the transaction's changes, writes a rollback record, releases locks, and unpins buffers.
- `(*Transaction) Recover()` — flushes all buffers and runs system recovery over the log.
- `(*Transaction) Pin(block *file.BlockId) error` / `Unpin(block *file.BlockId)` — pin/unpin a block for this transaction.
- `(*Transaction) GetInt(block *file.BlockId, offset int) (int32, error)` / `GetString(...)` — acquire a shared lock and read a value.
- `(*Transaction) SetInt(block *file.BlockId, offset int32, val int32, okToLog bool) error` / `SetString(...)` — acquire an exclusive lock, optionally log the old value, and write the new value.
- `(*Transaction) Size(filename string) int` — shared-locks the end-of-file marker and returns the block count.
- `(*Transaction) Append(filename string) *file.BlockId` — exclusive-locks the end-of-file marker and appends a new block.
- `(*Transaction) BlockSize() int` / `AvailableBuffs() int` — file block size and count of available buffers.

## How it fits

Depends on `buffer`, `file`, and `log` for storage, and on `transaction/recovery` and `transaction/concurrency` for the recovery manager and lock-based concurrency control. It is the entry point used by higher-level components (record, metadata, planning/query layers) that need transactional block access.

## Notes

- The next transaction number is a package-global counter guarded by a mutex; the counter is not persisted across restarts.
- Locking is block-granular via a shared package-global lock table; a lock-acquisition timeout surfaces as an error from `Get*`/`Set*` (and a panic from `Size`/`Append`).
- Several failures are handled with `panic` rather than returned errors.
- `Set*` write the before-image only when `okToLog` is true; undo during rollback/recovery calls `Set*` with logging disabled.
- Tests: `transaction_test.go`, `concurrency_test.go`, and `concurrency_lock_test.go`.
