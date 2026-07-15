# index — Static hash index

A static, bucket-based hash index (`HashIndex`) that implements the `query.Index`
interface. Index entries are spread across a fixed number of buckets, each bucket
being backed by its own record table.

## Responsibilities

- Provide equality-search access to a table's records via an indexed field.
- Map a search key to one of a fixed set of buckets and scan only that bucket.
- Insert and delete index records (`dataval`, `block`, `id`) that point back to
  the indexed data records by RID.
- Estimate index search cost for the query planner.

## Key types

- `HashIndex` — static hash index over `NUM_BUCKETS` (100) buckets; holds a
  transaction, index name, record layout, current search key, and an open
  `query.TableScan` on the matched bucket.
- `NUM_BUCKETS` — compile-time constant (100) fixing the number of buckets.

## Key API

- `NewHashIndex(tx *transaction.Transaction, idxName string, layout *record.Layout) *HashIndex`
  — construct an index over the given index name and layout.
- `(*HashIndex) BeforeFirst(searchKey *record.Constant) error` — compute the
  bucket from `searchKey.HashCode() % NUM_BUCKETS` and open a table scan on that
  bucket's table (`idxName + bucket`).
- `(*HashIndex) Next() (bool, error)` — advance to the next bucket record whose
  `dataval` equals the search key.
- `(*HashIndex) GetDataRID() (*record.RID, error)` — read the `block`/`id` of the
  current record and return the corresponding data RID.
- `(*HashIndex) Insert(dataval *record.Constant, datarid *record.RID) error` —
  position at the search key's bucket and append an index record.
- `(*HashIndex) Delete(dataval *record.Constant, datarid *record.RID) error` —
  scan the bucket for the matching RID and delete it.
- `(*HashIndex) Close() error` — close the current bucket table scan.
- `SearchCost(numBlock, recordsPerBlock int) int` — planner cost estimate
  (`numBlock / recordsPerBlock`).

## How it fits

Depends on `query` (for `Index` and `TableScan`), `record`, and `transaction`.
It satisfies the `query.Index` interface and is intended to be produced by
`metadata.IndexInfo.Open` and consumed by index-based query operators.

## Notes

- Wiring status: `HashIndex` is currently NOT connected to the catalog. In
  `metadata/index_metadata.go`, `IndexInfo.Open` returns a B-tree index
  (`index/btree.NewBTreeIndex`); the `index.NewHashIndex(...)` call is present but
  commented out. As a result `HashIndex` is fully implemented but not reached from
  SQL end-to-end.
- Buckets are static (fixed at 100); there is no rehashing or dynamic growth.
- `Next`, `GetDataRID`, `Insert`, and `Delete` assume `BeforeFirst` has opened a
  bucket scan; calling them beforehand would dereference a nil scan.
- Tests: `index/index_test.go` (`TestIndexRetrieval`) lives in this package but
  drives whatever `metadata.IndexInfo.Open` returns; because that currently yields
  a B-tree index, the test does not exercise `HashIndex` today. There is no test
  that constructs `HashIndex` directly.
