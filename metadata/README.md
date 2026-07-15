# metadata — System catalog and statistics

Maintains the database's system catalog: table and field definitions, view definitions, index definitions, and cost statistics. Provides a single façade (`MetadataMgmt`) for creating and querying this metadata.

## Responsibilities

- Persist and retrieve table schemas/layouts (table and field catalogs).
- Persist and retrieve view definitions.
- Persist index definitions and build `IndexInfo` (including index layout and cost estimates).
- Compute and cache per-table statistics (block count, record count, distinct-value estimates).
- Bootstrap the catalog tables when the database is new.

## Key types

- `MetadataMgmt` — top-level façade composing the table, view, statistics, and index managers.
- `TableMgmt` — manages the table catalog (`tblcat`) and field catalog (`fldcat`); stores and reconstructs `record.Layout`.
- `ViewMgmt` — manages the view catalog (`viewcat`).
- `StatMgmt` — computes and caches `StatInfo` per table; periodically refreshes.
- `StatInfo` — per-table statistics: block count, record count, and distinct-value estimates.
- `IndexMgmt` — manages the index catalog (`idxcat`) and builds `IndexInfo` maps.
- `IndexInfo` — describes one index: its layout, backing schema, statistics, and cost estimates; opens the physical index.

## Key API

- `NewMetadataMgmt(isNew bool, tx *transaction.Transaction) (*MetadataMgmt, error)` — constructs all sub-managers; creates catalog tables when `isNew`.
- `(*MetadataMgmt) CreateTable(name string, sch *record.Schema, tx)`, `GetLayout(name string, tx) (*record.Layout, error)`.
- `(*MetadataMgmt) CreateView(name, def string, tx)`, `GetViewDef(name string, tx) (string, error)`.
- `(*MetadataMgmt) CreateIndex(idxName, tableName, fieldName string, tx)`, `GetIndexInfo(tableName string, tx) (map[string]*IndexInfo, error)` — keyed by field name.
- `(*MetadataMgmt) GetStatInfo(name string, layout *record.Layout, tx) (*StatInfo, error)`, `ForceRefreshStatistics(tx) error`.
- `NewTableMgmt(isNew bool, tx)`; `(*TableMgmt) CreateTable(...)`, `GetLayout(...)`.
- `(*StatInfo) BlockAccessed() int32`, `RecordsOutput() int32`, `DistinctValues(field string) int32`.
- `(*IndexInfo) Open() query.Index`, `BlockAccessed()`, `RecordsOutput()`, `DistinctValues(field string)`.

## How it fits

Depends on `record` (schemas, layouts), `query` (table scans used to read/write catalogs), `transaction` (durable access), and `index/btree` (the physical index opened by `IndexInfo`). It is the catalog layer consumed by planning/optimization and DDL execution higher in the stack.

## Notes

- Catalog tables maintained: `tblcat` (table name, slot size) and `fldcat` (table name, field name, type, length, offset) for schemas; `viewcat` (view name, definition) for views; `idxcat` (index name, table name, field name) for indexes.
- Fixed name lengths: `MAX_NAME` = 16; view definitions up to `MAX_VIEWDEF` = 100.
- Statistics are estimates: `DistinctValues` assumes ~1/3 of a field's values are distinct (`1 + numRecords/3`), a documented simplification to be refined.
- `StatMgmt` caches statistics in memory (mutex-guarded) and force-refreshes every 100 `GetStatInfo` calls; `refreshStatistics` and `calcTableStats` print diagnostic output.
- `IndexInfo.Open` currently returns a B-tree index (`btree.NewBTreeIndex`) and panics on error; a hash-index alternative is present but commented out.
- Test coverage: `metadata_test.go` (`TestMetadata`, end-to-end over table/stat/view/index metadata) and `table_metadata_test.go` (`TestTableMgmt`, `TestCatalog`).
