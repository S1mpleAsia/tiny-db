# btree (index/btree) — Disk-backed B-tree index

A disk-backed B-tree index (`BTreeIndex`) that implements the `query.Index`
interface. It stores index entries in transaction-managed, block-structured leaf
and directory pages, supporting duplicate keys via overflow leaf chains.

## Responsibilities

- Provide ordered, equality-search index access over a table field via `query.Index`.
- Maintain a multi-level directory tree that routes a search key down to the
  correct leaf block.
- Store leaf entries (`dataval`, `block`, `id`) pointing back to data records by RID.
- Insert and delete index entries, splitting leaf and directory pages (and growing
  the tree root) as pages fill.
- Handle duplicate keys through overflow leaf blocks chained by the page flag.
- Estimate index search cost for the query planner.

## Key types

- `BTreeIndex` — top-level index; owns leaf/directory layouts, the leaf table name,
  the root directory block, and the currently open leaf.
- `BTreeDir` — a directory (internal) node wrapper over a `BTreePage`; performs
  search routing, entry insertion, splitting, and new-root creation.
- `BTreeLeaf` — a leaf node wrapper over a `BTreePage`; iterates matching entries,
  inserts/deletes RIDs, and follows overflow chains.
- `BTreePage` — low-level page abstraction over a pinned block: a sorted slot list
  with a 4-byte flag header (directory level, or overflow-leaf pointer) plus a
  record count; provides slot search, record shifting, and page splitting.
- `DirEntry` — an in-memory directory entry pairing a `dataval` with a child block
  number.

## Key API

- `NewBTreeIndex(tx *transaction.Transaction, idxName string, leafLayout *record.Layout) (*BTreeIndex, error)`
  — create/open the leaf table (`idxName+"leaf"`) and directory table
  (`idxName+"dir"`), formatting and seeding the root directory if empty.
- `(*BTreeIndex) BeforeFirst(searchKey *record.Constant) error` — search the
  directory from the root to the leaf block and open the leaf at the search key.
- `(*BTreeIndex) Next() (bool, error)` / `GetDataRID() (*record.RID, error)` —
  iterate leaf entries matching the key and read their data RIDs.
- `(*BTreeIndex) Insert(dataVal *record.Constant, dataRID *record.RID) error` —
  insert a leaf entry, propagating splits up the directory and creating a new root
  when the top level splits.
- `(*BTreeIndex) Delete(dataVal *record.Constant, dataRID *record.RID) error` —
  remove the matching leaf entry.
- `(*BTreeIndex) Close() error` — close the open leaf.
- `SearchCost(numBlocks, rpb int32) int32` — planner cost estimate
  (`1 + log_rpb(numBlocks)`).
- `NewBTreeDir(...)`, `(*BTreeDir) Search / Insert / InsertEntry / MakeNewRoot` —
  directory navigation and maintenance.
- `NewBTreeLeaf(...)`, `(*BTreeLeaf) Next / Insert / Delete / GetDataRID` — leaf
  iteration and maintenance, including overflow handling.
- `NewBTreePage(...)`, `(*BTreePage) Format / FindSlotBefore / Split / IsFull /
  InsertDir / InsertLeaf / GetDataVal / GetDataRID / GetFlag / SetFlag` — page
  primitives shared by directory and leaf nodes.
- `NewDirEntry(dataval *record.Constant, block int32) *DirEntry` — build a
  directory entry.

## How it fits

Depends on `file` (block identifiers), `record` (schema, layout, constants, RID),
`transaction` (pinning, typed reads/writes), and `query` (the `Index` interface it
implements). It is produced by `metadata.IndexInfo.Open`.

## Notes

- Wiring status: this B-tree index IS currently the index type connected to the
  catalog. `metadata/index_metadata.go` `IndexInfo.Open` returns
  `btree.NewBTreeIndex(...)`, so it is what SQL index access uses end-to-end today.
  (The alternative `index.HashIndex` construction in `Open` is commented out.)
- This is one of two separate B-tree implementations in the repository. This one
  (`index/btree`) is a disk-backed, RID-carrying database index built on the
  transaction/record layers. It is distinct from the top-level `btree` package,
  which is a standalone integer-keyed B-tree with its own file storage and no
  engine integration.
- Duplicate keys are supported via overflow leaf blocks: a leaf's flag holds the
  block number of the next overflow leaf.
- Tests: there are no test files in this package; correctness is exercised only
  indirectly through the index path used by `index/index_test.go`.
