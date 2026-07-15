# query — relational operators (scan layer)

Implements the relational-operator layer of TinyDB. Each relational operator is realized as a `Scan` that iterates over records; scans compose into a query tree, and query predicates/expressions are evaluated against them.

## Responsibilities

- Define the iteration contract (`Scan`, `UpdateScan`) that every relational operator implements.
- Provide concrete operators: table access, selection, projection, product (join), sort/merge, aggregation, index access, and multibuffer variants.
- Model query predicates as conjunctions of equality terms over expressions (field name or constant).
- Support materialization via temporary tables and record comparison for sorting/grouping.
- Expose cost hints used by planning (`ReductionFactor`, `DistinctValues` via `PlanLike`).

## Key interfaces

- `Scan` — a node in the query tree. Positioning + read access: `BeforeFirst`, `Next`, `GetInt`/`GetString`/`GetVal`, `HasField`, `Close`. `Next` advances and reports whether a record exists.
- `UpdateScan` — extends `Scan` with mutation: `SetInt`/`SetString`/`SetVal`, `Insert`, `Delete`, `GetRID`, `MoveToRID`. Only scans backed by stored records are updatable.
- `Index` — access-path abstraction: `BeforeFirst(searchKey)`, `Next`, `GetDataRID`, `Insert`, `Delete`, `Close`. Positions on index entries matching a key and yields data RIDs.
- `AggregateFn` — per-group aggregation: `ProcessFirst` (start group), `ProcessNext` (fold in a record), `FieldName` (result column name), `Value` (current result).

## Scans / operators

Base:
- `TableScan` (`table_scan.go`) — reads/writes a stored table file (`.tbl`) via record pages; implements `UpdateScan`.
- `SelectScan` (`select_scan.go`) — filters an underlying scan by a `Predicate`; passes through updates when the source is updatable.
- `ProjectScan` (`project_scan.go`) — restricts visible fields to a field list (read-only).
- `ProductScan` (`product_scan.go`) — Cartesian product of two scans (nested-loop join basis).

Predicates / expressions:
- `Expression` (`expression.go`) — a constant or a field reference; `Evaluate` resolves it against a scan.
- `Term` (`term.go`) — equality between two expressions; also detects `F=c` / `F1=F2` forms and computes a reduction factor.
- `Predicate` (`predicate.go`) — conjunction (AND) of terms; supports `SelectSubPred`/`JoinSubPred` splitting and reduction-factor estimation. Only equality and AND are supported.

Materialize / sort:
- `TempTable` (`temp_table.go`) — a uniquely named temporary stored table used for materialized runs.
- `RecordComparator` (`record_comparator.go`) — compares records (or value maps) across a list of fields.
- `SortScan` (`sort_scan.go`) — merges 1 or 2 sorted runs (k=2 merge); supports `SavePosition`/`RestorePosition`.
- `MergeJoinScan` (`merge_join_scan.go`) — equi-join of two `SortScan`s sorted on the join fields.

Aggregation:
- `AggregateFn` implementations: `CountFn` (`count_fn.go`), `MaxFn` (`max_fn.go`), `MinFn` (`min_fn.go`).
- `GroupValue` (`group_value.go`) — captures the grouping-field values of the current record.
- `GroupByScan` (`group_by_scan.go`) — consumes a `SortScan` sorted on the group fields, emitting one row per group with aggregate results.

Index:
- `IndexSelectScan` (`index_select_scan.go`) — uses an `Index` to fetch matching records of an underlying (updatable) scan for a constant key.
- `IndexJoinScan` (`index_join_scan.go`) — for each LHS record, probes an index on the RHS join field.

Multibuffer:
- `ChunkScan` (`chunk_scan.go`) — scans a contiguous range of blocks pinned together as one chunk.
- `MultiBufferProductScan` (`multi_buffer_product_scan.go`) — product that streams the RHS one chunk at a time, sized by available buffers.
- `MultiBufferSortScan` (`multi_buffer_sort_scan.go`) — merges an arbitrary number of runs at once.
- `multi_buffer.go` — buffer-sizing helpers `BufferNeedsBestRoot` and `BufferNeedsBestFactor`.

## Key API

- `Scan`: `BeforeFirst() error`, `Next() bool`, `GetInt(field) (int32, error)`, `GetString(field) (string, error)`, `GetVal(field) (*record.Constant, error)`, `HasField(field) bool`, `Close()`.
- `UpdateScan` (adds): `SetInt`/`SetString`/`SetVal`, `Insert() error`, `Delete() error`, `GetRID() *record.RID`, `MoveToRID(rid)`.
- `Index`: `BeforeFirst(searchKey *record.Constant) error`, `Next() (bool, error)`, `GetDataRID() (*record.RID, error)`, `Insert`/`Delete`, `Close`.
- `AggregateFn`: `ProcessFirst(scan) error`, `ProcessNext(scan) error`, `FieldName() string`, `Value() *record.Constant`.
- Predicate helpers: `IsSatisfied(scan)`, `ConjoinWith`, `SelectSubPred`, `JoinSubPred`, `EquatesWithConstant`, `EquatesWithField`, `ReductionFactor(PlanLike)`.

## How it fits

Depends on `record` (schemas, layouts, record pages, constants, RIDs) and `transaction` for block/buffer access; `TempTable` builds on `TableScan`. Planning cost hints go through the local `PlanLike` interface. These scans are the runtime consumed by the `plan/` layer, which builds and opens the corresponding query trees.

## Notes

- Predicates support only equality terms combined with AND; no inequality, OR, or arithmetic.
- `SortScan` merges at most 2 runs per instance; broader fan-in merging is only available via `MultiBufferSortScan`.
- `MultiBufferSortScan.Next` and `MultiBufferProductScan.useNextChunk` contain debug `fmt.Printf` output, and `MultiBufferSortScan.Next` does not assign `currentScan` from the computed minimum — these multibuffer paths appear incomplete and primarily programmatic rather than fully wired into SQL execution.
- Some passthrough scans (`SelectScan`) return `ErrNotUpdatable` when the source is not an `UpdateScan`; `IndexSelectScan`/`IndexJoinScan` assume their base scan is updatable (they type-assert to `UpdateScan`).
- Tests: `scan_test.go` (select/project over a table, and product+select join) and `table_scan_test.go` (insert/delete round-trip on `TableScan`).
