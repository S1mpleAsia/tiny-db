# plan — query planning

Builds executable query trees (`Plan`s) from parsed SQL and executes update
commands. Each `Plan` estimates its own cost and opens a `query.Scan` that
produces records.

## Responsibilities

- Define the `Plan` abstraction: open a scan plus report cost/statistics estimates and the output schema.
- Provide physical plan implementations for table access, selection, projection, product, materialization, sorting, joins, grouping, and index-based access.
- Translate `parse.QueryData` into a plan tree (query planning) and execute `parse.UpdateCmd` values (update planning).
- Expose a `Planner` facade that parses, verifies, and dispatches SQL to the query/update planners.

## Key types

- `Plan` — interface every plan implements (see Key API).
- `QueryPlanner` / `UpdatePlanner` — interfaces for producing query plans and executing update commands.
- `Planner` — facade holding a `QueryPlanner` and `UpdatePlanner`; parses SQL and routes it.
- `BasicQueryPlanner` — naive query planner: table/view plans, then product, select, project.
- `BasicUpdatePlanner` — executes inserts/deletes/modifies and DDL without index maintenance.
- `IndexUpdatePlanner` — like `BasicUpdatePlanner` but also maintains all indexes on the affected table.
- `TablePlan` — scans a stored table, sourcing stats from `metadata`.
- `SelectPlan` — applies a `*query.Predicate` to a sub-plan.
- `ProjectPlan` — restricts output to a field list.
- `ProductPlan` — nested-loop Cartesian product of two plans.
- `MaterializePlan` — copies a sub-plan into a temp table.
- `SortPlan` — external merge sort (split into runs, then repeated merge passes) over a `query.RecordComparator`.
- `MergeJoinPlan` — merge join of two `SortPlan`s on equality of two fields.
- `GroupByPlan` — grouping over a `SortPlan` with a set of `query.AggregateFn`.
- `IndexSelectPlan` — selection using an index and a constant value.
- `IndexJoinPlan` — join using an index on the right-hand table's join field.
- `MultiBufferProductPlan` — chunked, multi-buffer product; materializes its LHS.

## Key API

- `Plan` interface:
  - `Open() (query.Scan, error)` — create the scan for this plan.
  - `BlockAccessed() int32` — estimated block accesses.
  - `RecordsOutput() int32` — estimated output record count.
  - `DistinctValues(fieldName string) int32` — estimated distinct values of a field.
  - `Schema() *record.Schema` — output schema.
- `NewPlanner(queryPlanner QueryPlanner, updatePlanner UpdatePlanner) *Planner`.
- `func (*Planner) CreateQueryPlan(query string, tx *transaction.Transaction) (Plan, error)` — parse + verify + plan a `SELECT`.
- `func (*Planner) ExecuteUpdate(cmd string, tx *transaction.Transaction) (int, error)` — parse + verify + execute an update/DDL command; returns affected-row count.
- `func (QueryPlanner) CreatePlan(queryData *parse.QueryData, tx *transaction.Transaction) (Plan, error)`.
- `UpdatePlanner` methods: `ExecuteInsert`, `ExecuteDelete`, `ExecuteModify`, `ExecuteCreateTable`, `ExecuteCreateView`, `ExecuteCreateIndex` (each `(data, tx) (int, error)`).
- Planner constructors: `NewBasicQueryPlanner(mdm *metadata.MetadataMgmt)`, `NewBasicUpdatePlanner(mdm)`, `NewIndexUpdatePlanner(mdm)`.
- Plan constructors: `NewTablePlan(tx, tableName, md)`, `NewSelectPlan(p, pred)`, `NewProjectPlan(p, fieldList)`, `NewProductPlan(p1, p2)`, `NewMaterializePlan(srcPlan, tx)`, `NewSortPlan(p, tx, sortFields)`, `NewMergeJoinPlan(tx, p1, p2, fldName1, fldName2)`, `NewGroupByPlan(tx, plan, groupFields, aggFns)`, `NewIndexSelectPlan(p, indexInfo, val)`, `NewIndexJoinPlan(p1, p2, indexInfo, joinField)`, `NewMultiBufferProductPlan(tx, lhs, rhs)`.

## How it fits

Depends on `parse` (AST input and view re-parsing), `query` (scans, predicates,
expressions, temp tables, aggregate functions, comparators), `record` (schema,
layout, constants), `metadata` (layouts, statistics, index info), and
`transaction`. It is the layer that turns SQL text into runnable scans and is the
primary entry point (`Planner`) for higher-level code that executes SQL.

## Notes

- Only naive planning is implemented: `BasicQueryPlanner` always joins tables with `ProductPlan` in listing order, then selects and projects. There is no heuristic or cost-based planner that chooses among the join/index plans, so `MergeJoinPlan`, `IndexSelectPlan`, `IndexJoinPlan`, and `MultiBufferProductPlan` are constructed only directly (e.g. in tests), not by any planner.
- `Planner.verifyQuery` and `Planner.verifyUpdate` are stubs that always return `nil` (marked `TODO`).
- `multi_buffer_sort_plan.go` contains only the package declaration — no multi-buffer sort plan is implemented.
- `SortPlan` merges down to at most 2 runs and relies on `query.SortScan` to do the final merge; `MergeJoinPlan` and `GroupByPlan` type-assert their input scan to `*query.SortScan`.
- `MaterializePlan` and `MultiBufferProductPlan` emit `fmt.Println`/`Printf` debug output during `Open`/`BlockAccessed`.
- Tests: `plan_test.go`, `sort_plan_test.go`, `merge_join_plan_test.go`, `group_by_plan_test.go`, plus test-data directories `single_table_plan_test/`, `multiple_table_plan_test/`, `sort_plan_test/`, and `merge_join_plan_test/`.
