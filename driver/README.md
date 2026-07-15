# driver — `database/sql` driver for TinyDB

A Go `database/sql` driver that exposes the embedded TinyDB engine through the standard library's SQL interfaces. It is registered under the name `"tinydb"`; the DSN passed to `sql.Open` is the directory where the database files live.

## Responsibilities

- Register the `"tinydb"` driver with `database/sql` at package init.
- Open a TinyDB engine instance for a given data directory and hand back a connection.
- Bridge `database/sql` calls to the engine's planner and transactions.
- Translate engine query scans and update counts into `driver.Rows` and `driver.Result`.

## Key types

- `TinyDBDriver` — implements `driver.Driver`; opens connections from a directory name.
- `Connection` — implements `driver.Conn` and `driver.ConnBeginTx`; holds the engine, planner, and current transaction.
- `ConnectionTransaction` — implements `driver.Tx`; wraps a `transaction.Transaction` and clears it from the connection on commit/rollback.
- `TinyDBStmt` — implements `driver.Stmt`, `driver.StmtExecContext`, and `driver.StmtQueryContext`; carries the query text, planner, and transaction.
- `TinyDbRows` — implements `driver.Rows`; iterates a `query.Scan` and maps schema fields to result columns.
- `TinyDBResult` — implements `driver.Result`; reports rows affected.

## Key API

- `func (d *TinyDBDriver) Open(name string) (driver.Conn, error)` — opens the engine via `server.NewTinyDBWithMetadata(name)` and returns a `Connection`.
- `func NewConnection(db *server.TinyDB, planner *plan.Planner) *Connection` — constructs a connection.
- `func (conn *Connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error)` — starts an engine transaction; `Begin()` delegates to it.
- `func (conn *Connection) Prepare(query string) (driver.Stmt, error)` — builds a `TinyDBStmt` bound to the current transaction.
- `func (s *TinyDBStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error)` — runs `planner.ExecuteUpdate` and returns a `TinyDBResult`.
- `func (s *TinyDBStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error)` — builds and opens a query plan and returns `TinyDbRows`.
- `func (result *TinyDBResult) RowsAffected() (int64, error)` — number of rows changed.

## Usage

```go
package main

import (
	"context"
	"database/sql"
	"log"

	_ "s1mpleasia.com/tinydb/driver"
)

func main() {
	db, err := sql.Open("tinydb", "./playerdb") // DSN = data directory
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.TODO()

	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("create table player (player_id int, name varchar(10), point int)")
	stmt.ExecContext(ctx)
	tx.Commit()

	tx, _ = db.Begin()
	stmt, _ = tx.Prepare("insert into player (player_id, name, point) values (1, 'Nobak', 11055)")
	res, _ := stmt.ExecContext(ctx)
	n, _ := res.RowsAffected()
	log.Printf("inserted %d row(s)", n)
	tx.Commit()

	tx, _ = db.Begin()
	stmt, _ = tx.Prepare("select player_id, name, point from player")
	rows, _ := stmt.QueryContext(ctx)
	for rows.Next() {
		var id, point int
		var name string
		rows.Scan(&id, &name, &point)
		log.Printf("%d %s %d", id, name, point)
	}
	tx.Commit()
}
```

## How it fits

Depends on the `server` package (to open the engine), and on `plan`, `transaction`, `query`, and `record`. It is the public entry point for applications that want to use TinyDB through the standard `database/sql` API rather than calling the engine directly.

## Notes

- The driver works only through the context-based APIs (`ExecContext`, `QueryContext`, `BeginTx`). The legacy non-context `Stmt.Exec` and `Stmt.Query` methods panic with `"unimplemented"`.
- `TinyDBResult.LastInsertId` panics with `"unimplemented"`.
- `NumInput` returns `0`; bound placeholder arguments are not processed (queries are passed to the planner as literal SQL text).
- `Open` and the connection lifecycle print diagnostic messages to stdout.
- Tests live in `driver_test.go`, which exercises the full open/begin/prepare/exec/query/scan/commit/rollback flow.
