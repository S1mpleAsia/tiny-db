# TinyDB

## A Relational Database Engine Built From Scratch in Go

---

## 👋 Introduction

**TinyDB** is a small, from-scratch relational database engine written in Go. It implements a full database stack — **disk/file management, a write-ahead log, a buffer pool, ACID transactions, record storage, a system catalog, a SQL parser, a query planner, relational operators, indexing (hash + B-tree), and a `database/sql` driver** — so you can create tables and run real SQL against real files on disk.

The project is a hands-on study of how a database actually works under the hood: rather than reaching for an existing engine, every layer — from raw disk blocks up to the SQL front end — is implemented and wired together by hand. The system is intended for **pedagogical use only**; it is not tuned for performance or meant for production.

| | |
| :--- | :--- |
| **Author** | S1mpleAsia |
| **Language** | Go (1.23+) |
| **Dependencies** | [`stretchr/testify`](https://github.com/stretchr/testify) (tests only) |

---

## 🏗️ Architecture

TinyDB is built bottom-up as a set of layered packages — each layer depends only on the ones beneath it.

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

| Layer | Package(s) | Status |
|-------|-----------|--------|
| Disk & File Management | `file/` | ✅ Implemented |
| Log Management | `log/` | ✅ Implemented |
| Buffer Management | `buffer/` | ✅ Implemented |
| Transaction Management (recovery + concurrency) | `transaction/`, `transaction/recovery/`, `transaction/concurrency/` | ✅ Implemented (undo-only recovery; lock-based concurrency) |
| Record Management | `record/` | ✅ Implemented |
| Metadata Management | `metadata/` | ✅ Implemented |
| Query Processing (scans/operators) | `query/` | ✅ Implemented |
| SQL Parsing | `parse/` | ✅ Implemented (grammar subset) |
| Planning | `plan/` | ✅ Basic planner |
| JDBC-style driver | `driver/` | ✅ `database/sql` driver (context APIs) |
| Indexing | `index/`, `index/btree/` | ✅ Hash + B-tree indexes |
| Materialization & Sorting | `query/`, `plan/` (sort, merge join, group by, aggregates) | ✅ Implemented |
| Effective Buffer Utilization (multibuffer) | `query/`, `plan/` (chunk / multibuffer product & sort) | ✅ Implemented |
| Query Optimization (heuristic planner) | — | ❌ Not implemented |

---

## ✨ What's Implemented

### 💾 Storage & Runtime

- **Block/page abstraction** over OS files with a configurable block size (default **400 bytes**), little-endian integers and UTF-16 strings (`file/`).
- **Write-ahead log** with reverse-order iteration and LSN tracking (`log/`).
- **Buffer pool** with pinning, naive replacement, and deadlock-avoidance via a wait timeout (default pool size **8 buffers**) (`buffer/`).
- **Transactions** with **commit / rollback / recover**, undo-based recovery with quiescent checkpoints, and **lock-based concurrency control** (shared/exclusive locks, timeout used as deadlock detection) (`transaction/`).

### 🗃️ Records & Catalog

- **Slotted record pages**, schemas, layouts, RIDs, and typed constants (int / varchar) (`record/`).
- **System catalog**: table metadata (`tblcat`/`fldcat`), views (`viewcat`), indexes (`idxcat`), and cost statistics (`metadata/`).

### 🔎 Query Engine

- **Relational operators as scans**: table, select, project, product, plus index select/join, merge join, sort, group-by with **count / min / max** aggregates, and multibuffer variants (`query/`).
- **SQL lexer + recursive-descent parser** (`parse/`).
- **Query planning**: a basic query planner (view expansion, cross products, select, project) and both a **basic** and an **index-aware** update planner (`plan/`).
- **External merge sort** (`SortPlan`) and materialized temp tables.

### 🌳 Indexing

- Disk-backed **B-tree index** (leaf/dir pages, splits, overflow) in `index/btree/` — this is the index type `metadata.IndexInfo.Open` currently returns, so it's the index used whenever one is opened.
- Static **hash index** (100 buckets) in `index/` — fully implemented, but the line that would select it in `IndexInfo.Open` is commented out, so it is inactive unless switched back on.

### 🔌 Access

- A `database/sql` driver registered as **`tinydb`**, supporting prepared statements, transactions, exec, and query via the context-based APIs (`driver/`).

---

## 🧩 Supported SQL

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

---

## ⚙️ Getting Started

Requires **Go 1.23+**.

```bash
git clone <this-repo>
cd tiny-db
go test ./...
```

### Using TinyDB through `database/sql`

The easiest way to use the engine is via the registered driver. The DSN is simply the directory where the database files should live.

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

See `driver/driver_test.go` for a complete, runnable example that creates a table, inserts, queries, updates, rolls back, and deletes.

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

---

## 📁 Project Layout

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

> Each package has its own `README.md` with details on its types, API, and how it fits into the stack — see [`docs/README.md`](docs/README.md) for the full index.

---

## 🧪 Testing

```bash
go test ./...
```

Test coverage exists for: `file`, `log`, `buffer`, `transaction` (including concurrency and lock-timeout tests), `record`, `metadata`, `parse` (lexer, predicate parser, query parser), `query` (scans, join+select, table scan), `plan` (single/multi-table, sort, merge-join, group-by), `index` (hash retrieval), and `driver` (end-to-end through `database/sql`).

There are currently no dedicated tests for `server`, `cli`, `testlib`, the standalone top-level `btree/`, or the `index/btree/` B-tree index.

---

## 🚧 Limitations

- **Heuristic / cost-based query optimizer.** `NewOptimizedTinyDB` and the `useBasic=false` branch in `server/server.go` are a `// TODO` — only the basic planner works. The basic planner does not use indexes for reads and does not reorder joins.
- **`ORDER BY`, `GROUP BY`, and aggregate functions are not in the SQL grammar.** The sort, group-by, and aggregate *plans/scans exist and are tested*, but they can only be used programmatically, not from a SQL string.
- **Indexes are not exercised by the default planners.** Although `IndexInfo.Open` returns a B-tree index, `server.go` installs `BasicQueryPlanner` / `BasicUpdatePlanner`, which neither read through indexes nor maintain them on writes. The index-aware `IndexUpdatePlanner` and the index scans/plans exist and are tested, but are only reached by constructing them directly.
- Recovery is **undo-only** (no redo).
- `Planner.verifyQuery` / `verifyUpdate` are no-op stubs (no semantic validation).
- The driver's legacy non-context `Stmt.Exec` / `Stmt.Query` panic; use the **context** variants (`ExecContext` / `QueryContext` / `BeginTx`), as `database/sql` does automatically.
- The interactive shell (`cli/`) is a skeleton that prints placeholders and is not wired to the engine. `main.go` is effectively empty.
- The top-level `btree/` package is a standalone integer-keyed B-tree, separate from the database and not connected to the rest of the engine.

---

## 📄 License

See [LICENSE](LICENSE).
