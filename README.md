# TinyDB

A small, from-scratch relational database engine written in Go, built for learning.

TinyDB is a Go port of **SimpleDB**, the teaching database from Edward Sciore's book
[*Database Design and Implementation*](https://simpledb-java.netlify.app/database-design-and-implementation.pdf)
(2nd ed., Springer). It implements a full stack — disk/file management, a write-ahead log, a
buffer pool, ACID transactions, record storage, a system catalog, a SQL parser, a query
planner, relational operators, indexing (hash + B-tree), and a `database/sql` driver — so you
can create tables and run SQL against real files on disk.

The system is intended for **pedagogical use only**. It is not tuned for performance or meant
for production.

- **Author:** S1mpleAsia
- **Language:** Go (1.23+)
- **Only runtime/test dependency:** [`stretchr/testify`](https://github.com/stretchr/testify)

## Architecture

TinyDB is built bottom-up as a set of layered packages. Each layer corresponds to a chapter (or
group of chapters) in the book.

```
                 database/sql  ──▶  driver/            (JDBC-equivalent driver)
                                        │
                                     server/           (TinyDB: wires everything together)
                                        │
        ┌───────────────────────────────┴───────────────────────────────┐
        │                                                                 │
     plan/          ──▶  parse/                                     metadata/
  (query planning)      (SQL lexer + parser)                       (system catalog)
        │                                                                 │
     query/         (relational operators / scans)                     record/
        │                                                          (records, schema, layout)
        └───────────────────────────────┬───────────────────────────────┘
                                     transaction/       (recovery + concurrency)
                                         │
                                     buffer/            (buffer pool)
                                         │
                                       log/             (write-ahead log)
                                         │
                                       file/            (blocks, pages, file manager)
```

| Layer | Package(s) | Book topic | Status |
|-------|-----------|-----------|--------|
| Disk & File Management | `file/` | Ch. 3 | ✅ Implemented |
| Log Management | `log/` | Ch. 4 | ✅ Implemented |
| Buffer Management | `buffer/` | Ch. 4 | ✅ Implemented |
| Transaction Management (recovery + concurrency) | `transaction/`, `transaction/recovery/`, `transaction/concurrency/` | Ch. 5 | ✅ Implemented (undo-only recovery; lock-based concurrency) |
| Record Management | `record/` | Ch. 6 | ✅ Implemented |
| Metadata Management | `metadata/` | Ch. 7 | ✅ Implemented |
| Query Processing (scans/operators) | `query/` | Ch. 8 | ✅ Implemented |
| SQL Parsing | `parse/` | Ch. 9 | ✅ Implemented (grammar subset) |
| Planning | `plan/` | Ch. 10 | ✅ Basic planner |
| JDBC-style driver | `driver/` | Ch. 11 | ✅ `database/sql` driver (context APIs) |
| Indexing | `index/`, `index/btree/` | Ch. 12 | ✅ Hash + B-tree indexes |
| Materialization & Sorting | `query/`, `plan/` (sort, merge join, group by, aggregates) | Ch. 13 | ✅ Implemented |
| Effective Buffer Utilization (multibuffer) | `query/`, `plan/` (chunk / multibuffer product & sort) | Ch. 14 | ✅ Implemented |
| Query Optimization (heuristic planner) | — | Ch. 15 | ❌ Not implemented |

## What's implemented

**Storage & runtime**
- Block/page abstraction over OS files with a configurable block size (default **400 bytes**),
  little-endian integers and UTF-16 strings (`file/`).
- Write-ahead log with reverse-order iteration and LSN tracking (`log/`).
- Buffer pool with pinning, naive replacement, and deadlock-avoidance via a wait timeout
  (default pool size **8 buffers**) (`buffer/`).
- Transactions with **commit / rollback / recover**, undo-based recovery with quiescent
  checkpoints, and **lock-based concurrency control** (shared/exclusive locks, timeout used as
  deadlock detection) (`transaction/`).

**Records & catalog**
- Slotted record pages, schemas, layouts, RIDs, and typed constants (int / varchar) (`record/`).
- System catalog: table metadata (`tblcat`/`fldcat`), views (`viewcat`), indexes (`idxcat`),
  and cost statistics (`metadata/`).

**Query engine**
- Relational operators as scans: table, select, project, product, plus index select/join,
  merge join, sort, group-by with **count / min / max** aggregates, and multibuffer variants
  (`query/`).
- SQL lexer + recursive-descent parser (`parse/`).
- Basic query planner (view expansion, cross products, select, project) and both a **basic**
  and an **index-aware** update planner (`plan/`).
- External merge sort (`SortPlan`) and materialized temp tables.

**Indexing**
- Static **hash index** (100 buckets) — this is the index type wired into the catalog today.
- Disk-backed **B-tree index** (leaf/dir pages, splits, overflow) — fully implemented in
  `index/btree/` but not yet connected to the catalog path (see limitations).

**Access**
- A `database/sql` driver registered as **`tinydb`**, supporting prepared statements,
  transactions, exec, and query via the context-based APIs (`driver/`).

### Supported SQL

The parser grammar (see the header comment in `parse/parser.go`) covers:

```sql
-- Queries
SELECT f1, f2, ... FROM t1, t2, ... [WHERE <pred>]

-- Updates
INSERT INTO t (f1, ...) VALUES (c1, ...)
DELETE FROM t [WHERE <pred>]
UPDATE t SET f = <expr> [WHERE <pred>]

-- DDL
CREATE TABLE t (f1 INT, f2 VARCHAR(n), ...)
CREATE VIEW v AS <query>
CREATE INDEX i ON t (f)
```

Predicates are conjunctions (`AND`) of equality terms, e.g. `WHERE sid = 3 AND major = 'math'`.

## Not yet implemented / limitations

- **Heuristic / cost-based query optimizer** (Ch. 15). `NewOptimizedTinyDB` and the
  `useBasic=false` branch in `server/server.go` are a `// TODO` — only the basic planner works.
  The basic planner does not use indexes for reads and does not reorder joins.
- **`ORDER BY`, `GROUP BY`, and aggregate functions are not in the SQL grammar.** The sort,
  group-by, and aggregate *plans/scans exist and are tested*, but they can only be used
  programmatically, not from a SQL string.
- **B-tree index is not wired into the catalog.** `metadata.IndexInfo.Open` returns a hash
  index, so end-to-end SQL uses hashing even though the B-tree implementation is complete.
- **Index-aware update planner is not wired into the server** — `server.go` installs
  `BasicUpdatePlanner`, so indexes are not automatically maintained through the default path.
- Recovery is **undo-only** (no redo).
- `Planner.verifyQuery` / `verifyUpdate` are no-op stubs (no semantic validation).
- The driver's legacy non-context `Stmt.Exec` / `Stmt.Query` panic; use the **context**
  variants (`ExecContext` / `QueryContext` / `BeginTx`), as `database/sql` does automatically.
- The interactive shell (`cli/`) is a skeleton that prints placeholders and is not wired to the
  engine. `main.go` is effectively empty.
- The top-level `btree/` package is a standalone integer-keyed B-tree, separate from the
  database and not connected to the rest of the engine.

## Getting started

Requires Go 1.23+.

```bash
git clone <this-repo>
cd tiny-db
go test ./...
```

### Using TinyDB through `database/sql`

The easiest way to use the engine is via the registered driver. The DSN is simply the directory
where the database files should live.

```go
package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "s1mpleasia.com/tinydb/driver" // registers the "tinydb" driver
)

func main() {
	db, err := sql.Open("tinydb", "./playerdb")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx := context.Background()

	// DDL
	tx, _ := db.Begin()
	stmt, _ := tx.Prepare("create table player (player_id int, name varchar(10), point int)")
	stmt.ExecContext(ctx)
	tx.Commit()

	tx, _ = db.Begin()
	ins, _ := tx.Prepare("insert into player (player_id, name, point) values (1, 'Nobak', 11055)")
	ins.ExecContext(ctx)
	tx.Commit()

	// Query
	tx, _ = db.Begin()
	q, _ := tx.Prepare("select player_id, name, point from player")
	rows, _ := q.QueryContext(ctx)
	for rows.Next() {
		var id, point int
		var name string
		rows.Scan(&id, &name, &point)
		fmt.Printf("%d %s %d\n", id, name, point)
	}
	tx.Commit()
}
```

See `driver/driver_test.go` for a complete, runnable example that creates a table, inserts,
queries, updates, rolls back, and deletes.

### Using the engine directly

You can also embed the engine and drive the planner yourself:

```go
db, _ := server.NewTinyDBWithMetadata("./mydb")
tx, _ := db.NewTx()
planner := db.Planner()

planner.ExecuteUpdate("create table t (a int, b varchar(9))", tx)
planner.ExecuteUpdate("insert into t (a, b) values (1, 'hello')", tx)

plan, _ := planner.CreateQueryPlan("select a, b from t where a = 1", tx)
scan, _ := plan.Open()
for scan.Next() {
	// scan.GetInt("a"), scan.GetString("b")
}
tx.Commit()
```

## Project layout

```
file/                 Blocks, pages, and the file manager
log/                  Write-ahead log manager and iterator
buffer/               Buffer pool
transaction/          Transactions
  recovery/           Undo recovery + log records
  concurrency/        Lock table + concurrency manager
record/               Schema, layout, record pages, RID, constants
metadata/             Table / view / index / statistics catalog
query/                Scans and relational operators (incl. sort, join, group-by, multibuffer)
parse/                SQL lexer, parser, and AST
plan/                 Plans and query/update planners
index/                Hash index
  btree/              B-tree index (pages, dir, leaf)
btree/                Standalone integer B-tree (not wired into the engine)
driver/               database/sql driver ("tinydb")
server/               TinyDB struct that wires all layers together
cli/                  Interactive shell skeleton (stub)
testlib/              Test-data helpers (university schema)
```

## Testing

```bash
go test ./...
```

Test coverage exists for: `file`, `log`, `buffer`, `transaction` (including concurrency and
lock-timeout tests), `record`, `metadata`, `parse` (lexer, predicate parser, query parser),
`query` (scans, join+select, table scan), `plan` (single/multi-table, sort, merge-join,
group-by), `index` (hash retrieval), and `driver` (end-to-end through `database/sql`).

There are currently no dedicated tests for `server`, `cli`, `testlib`, the standalone top-level
`btree/`, or the `index/btree/` B-tree index.

## Credits

Based on the design and pseudocode in Edward Sciore, *Database Design and Implementation*
(Springer, 2020). This is an independent Go reimplementation for educational purposes.

## License

See [LICENSE](LICENSE).
```