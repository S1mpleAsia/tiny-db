## TinyDB

A lightweight, small and simple DB written in Go. The system is intended for pedagogical use only.

For the project overview, feature status, and usage examples, see the [top-level README](../README.md).

### Architecture

TinyDB is built using a bottom-up, modular approach. Each package has its own README with
details on its types, API, and how it fits into the stack.

**Storage & runtime**
- [Disk & File Management](../file/README.md) — blocks, pages, and the file manager (`file/`)
- [Log Management](../log/README.md) — the write-ahead log (`log/`)
- [Buffer Management](../buffer/README.md) — the buffer pool (`buffer/`)
- [Transaction Management](../transaction/README.md) — transactions (`transaction/`)
  - [Recovery](../transaction/recovery/README.md) — undo-based recovery and log records
  - [Concurrency](../transaction/concurrency/README.md) — lock table and concurrency manager

**Records & catalog**
- [Record Management](../record/README.md) — schema, layout, record pages, RID, constants (`record/`)
- [Metadata Management](../metadata/README.md) — the system catalog and statistics (`metadata/`)

**Query engine**
- [Query Processing](../query/README.md) — relational operators / scans (`query/`)
- [Parsing](../parse/README.md) — SQL lexer, parser, and AST (`parse/`)
- [Planning](../plan/README.md) — plans and query/update planners (`plan/`)

**Indexing**
- [Hash Index](../index/README.md) — static hash index (`index/`)
- [B-tree Index](../index/btree/README.md) — disk-backed B-tree index (`index/btree/`)
- [Standalone B-tree](../btree/README.md) — integer-keyed B-tree, separate from the engine (`btree/`)

**Access & wiring**
- [Driver](../driver/README.md) — the `database/sql` driver, `"tinydb"` (`driver/`)
- [Server](../server/README.md) — the embedded `TinyDB` struct that wires everything together (`server/`)
- [CLI](../cli/README.md) — interactive shell skeleton (`cli/`)
- [Test helpers](../testlib/README.md) — sample-data seeding for tests (`testlib/`)
