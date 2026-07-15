# buffer — Buffer pool management

The `buffer` package implements an in-memory buffer pool that caches disk pages, mediating access between the storage engine and the `file` package. It pins/unpins buffers, tracks modifications, and enforces write-ahead logging on flush.

## Responsibilities

- Hold a fixed-size pool of buffers, each wrapping one in-memory page.
- Associate buffers with disk blocks and read block contents into them.
- Pin and unpin buffers, tracking how many clients are using each.
- Track which transaction modified a buffer and the LSN of its most recent change.
- Flush dirty buffers to disk, writing the log first (write-ahead logging).
- Block and wait (with a timeout) when no buffer is available, aborting on suspected deadlock.

## Key types

- `Buffer` — wraps a `file.Page` plus its assigned block, pin count, modifying transaction number, and LSN.
- `BufferMgmt` — the buffer pool; assigns blocks to buffers and coordinates concurrent pin/unpin requests via a mutex and condition variable.

## Key API

- `NewBuffer(fm *file.FileMgmt, lm *log.LogMgmt) *Buffer` — creates an unpinned, unassigned buffer.
- `(*Buffer) Contents() *file.Page` — the buffer's page.
- `(*Buffer) Block() *file.BlockId` — the currently assigned block (nil if none).
- `(*Buffer) IsPinned() bool` — whether the pin count is greater than zero.
- `(*Buffer) SetModified(txNum int, lsn int)` — marks the buffer dirty for a transaction and records the LSN (when non-negative).
- `(*Buffer) ModifyingTx() int` — the transaction number that last modified the buffer (-1 if clean).
- `NewBufferMgmt(fm *file.FileMgmt, lm *log.LogMgmt, numBuffs int) *BufferMgmt` — builds a pool of `numBuffs` buffers.
- `(*BufferMgmt) Pin(block *file.BlockId) (*Buffer, error)` — pins a buffer to the block, waiting up to `MAX_TIME` and returning an error if none becomes available.
- `(*BufferMgmt) Unpin(buffer *Buffer)` — unpins a buffer and wakes waiters if it becomes free.
- `(*BufferMgmt) Available() int` — number of unpinned buffers.
- `(*BufferMgmt) FlushAll(txNum int)` — flushes every buffer modified by the given transaction.

## How it fits

Depends on the `file` package (for pages, blocks, and disk I/O) and the `log` package (to flush the log before writing dirty pages). It sits above both and is consumed by higher layers of the engine that need cached, pinned access to disk blocks.

## Notes

- `Buffer.flush()` (unexported) flushes the log up to the buffer's LSN, writes the page to disk, and resets `txNum` to -1 — only when the buffer is dirty (`txNum >= 0`).
- A fresh buffer has `txNum = -1` and `lsn = -1`; `assignToBlock` flushes any prior contents before reading the new block and resets the pin count to 0.
- `MAX_TIME` is 10 seconds; `Pin` loops, waiting on the condition variable and re-trying until a buffer frees up or the timeout elapses, then returns a "Buffer abort exception. Deadlock" error.
- `chooseUnpinnedBuffer` uses a naive first-unpinned-found replacement policy.
- The `Pin` wait loop spawns a goroutine that calls `cond.Wait()` and unlocks the mutex, coordinating with a timer via channels; this is a non-standard use of `sync.Cond`.
- Tested in `buffer_test.go` (`TestBuffer`, `TestBufferMgmt`), covering pinning, modification, replacement, and pool exhaustion.
