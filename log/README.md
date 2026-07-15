# log — Write-ahead log management

The `log` package implements an append-only write-ahead log built on top of the `file` package. It packs variable-length log records into blocks and lets callers iterate over them in reverse chronological order.

## Responsibilities

- Append raw log records to the log file and assign each a monotonically increasing LSN (log sequence number).
- Pack records into a page from right to left so they can later be read newest-first.
- Track which records live in memory versus which have been flushed to disk.
- Allocate new log blocks when the current page fills.
- Provide an iterator that walks all log records from most recent to oldest.

## Key types

- `LogMgmt` — manages the log file: holds the in-memory log page, current block, and the latest/last-saved LSNs.
- `LogIterator` — iterates over log records, moving block by block from the current block backward and reading records left to right within each block.

## Key API

- `NewLogMgmt(fileMgmt *file.FileMgmt, logFile string) (*LogMgmt, error)` — opens (or initializes) the log file, reading the last block or appending a first one.
- `(*LogMgmt) Append(record []byte) (int, error)` — appends a record and returns its LSN, allocating a new block if the current page is full.
- `(*LogMgmt) Flush(lsn int)` — flushes the log page to disk if the given LSN has not yet been saved.
- `(*LogMgmt) Iterator() *LogIterator` — flushes, then returns an iterator positioned at the current block.
- `NewIterator(fileMgmt *file.FileMgmt, block *file.BlockId) *LogIterator` — constructs an iterator starting at the given block.
- `(*LogIterator) HasNext() bool` — reports whether more records remain in this or earlier blocks.
- `(*LogIterator) Next() []byte` — returns the next record, advancing to the previous block when the current one is exhausted.

## How it fits

Depends on the `file` package for `FileMgmt`, `BlockId`, and `Page`. It is used by the `buffer` package, whose buffers flush the log up to their page's LSN before writing dirty data to disk, ensuring the write-ahead logging discipline.

## Notes

- The first 4 bytes of each log page store the boundary: the offset of the most recently written record. Records grow downward from the end of the page toward this boundary.
- A new block is allocated when there is not enough room to hold the record plus its 4-byte length while still leaving room for the boundary field.
- `Flush(lsn)` only forces a flush when `lsn >= lastSavedLSN`; the unexported `flush()` always writes the current block and updates `lastSavedLSN` to `latestLSN`.
- LSNs start at 1 (incremented before being returned); `latestLSN` and `lastSavedLSN` initialize to 0.
- Tested in `log_test.go`, which appends batches of records and verifies they iterate back in reverse order across block boundaries.
