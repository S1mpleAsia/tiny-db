# server — embedded TinyDB engine

The `server` package defines `TinyDB`, the top-level embedded database struct that wires together the file, log, buffer, metadata, and planner subsystems and exposes them behind a single handle.

## Responsibilities

- Construct and connect the storage stack (file management, write-ahead log, buffer pool).
- Initialize metadata management and detect whether the database is new or being recovered.
- Build the query/update planner and make it available to callers.
- Start new transactions against the shared subsystems.
- Expose accessors to the underlying managers.

## Key types

- `TinyDB` — owns `*file.FileMgmt`, `*log.LogMgmt`, `*buffer.BufferMgmt`, `*metadata.MetadataMgmt`, and `*plan.Planner`.

## Key API

- `func NewTinyDB(dbDir string, blockSize int, bufferSize int) (*TinyDB, error)` — builds only the storage stack (file/log/buffer); metadata and planner are left nil.
- `func NewTinyDBWithMetadata(dirName string) (*TinyDB, error)` — full initialization with metadata and the basic query/update planner; recovers an existing database or creates a new one.
- `func NewOptimizedTinyDB(dirName string) (*TinyDB, error)` — intended to use a heuristic planner; see Notes.
- `func (db *TinyDB) NewTx() (*transaction.Transaction, error)` — starts a new transaction.
- `func (db *TinyDB) FileMgmt() *file.FileMgmt` — file manager accessor.
- `func (db *TinyDB) LogMgmt() *log.LogMgmt` — log manager accessor.
- `func (db *TinyDB) BufferMgmt() *buffer.BufferMgmt` — buffer manager accessor.
- `func (db *TinyDB) MetadataMgmt() *metadata.MetadataMgmt` — metadata manager accessor.
- `func (db *TinyDB) Planner() *plan.Planner` — planner accessor.

## Constants

- `BLOCK_SIZE = 400` — block size in bytes used by `NewTinyDBWithMetadata`/`NewOptimizedTinyDB`.
- `BUFFER_SIZE = 8` — number of buffers in the pool.
- `logFile = "./tinydb.log"` — log file name (unexported).

## Usage

```go
package main

import (
	"log"

	"s1mpleasia.com/tinydb/server"
)

func main() {
	db, err := server.NewTinyDBWithMetadata("./mydb")
	if err != nil {
		log.Fatal(err)
	}

	tx, err := db.NewTx()
	if err != nil {
		log.Fatal(err)
	}

	planner := db.Planner()
	if _, err := planner.ExecuteUpdate("create table t(id int)", tx); err != nil {
		log.Fatal(err)
	}
	tx.Commit()
}
```

## How it fits

Sits above the storage and query subsystems (`file`, `log`, `buffer`, `metadata`, `plan`, `transaction`) and composes them into a usable engine. It is consumed directly by embedding callers and by the `driver` package, which opens a `TinyDB` per connection.

## Notes

- `NewOptimizedTinyDB` is non-functional: it routes to the internal initializer with `useBasic = false`, but that branch is only a `// TODO: Heuristic planner` stub. No planner is assigned, so the resulting `TinyDB` has a nil planner and cannot execute queries. Use `NewTinyDBWithMetadata` instead.
- `NewTinyDB` returns an engine without metadata or planner; it is a lower-level constructor used internally by `newTinyDBWithMetadata`.
- Construction prints status messages ("creating new database" / "recovering existing database") to stdout.
- There are no test files in this package.
