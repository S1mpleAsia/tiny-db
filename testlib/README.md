# testlib — test-data seeding helpers

Helpers that seed a sample university schema (`student`, `dept`, `course`, `section`, `enroll`) into a TinyDB instance via SQL, for use in tests across the codebase.

## Responsibilities

- Create the sample tables (and a couple of indexes) through the planner.
- Insert small, medium, or full-scale example datasets.
- Force a refresh of table statistics after loading so cost-based planning has current metadata.

## Key types

- `Student` — `SId, SName, GradYear, MajorId`.
- `Dept` — `DId, DName`.
- `Course` — `CId, Title, DeptId`.
- `Section` — `SectId, CourseId, Prof, YearOffered`.
- `Enroll` — `EId, SectionId, StudentId, Grade`.

## Key API

- `func InsertSmallTestData(t *testing.T, db *server.TinyDB) error` — creates `student` and `dept`, inserts 3 depts and 10 students, refreshes statistics, commits.
- `func InsertMiddleTestData(t *testing.T, db *server.TinyDB) error` — creates `student` and `enroll`, inserts 10 students and 100 enroll rows (reversed order), refreshes statistics, commits.
- `func InsertLargeTestData(t *testing.T, db *server.TinyDB) error` — creates all five tables and inserts a larger scaled dataset (450 students, 40 depts, 50 courses, 250 sections, 1,500 enrolls), then refreshes statistics.

All three take a `*testing.T` and a `*server.TinyDB`, run their own transaction obtained from `db.NewTx()`, and drive the engine through `db.Planner()`.

## How it fits

Depends on `server` (for the engine and transactions), `plan` (the planner), and `transaction`. It is a test-support package consumed by tests elsewhere in the module; it is not part of the runtime engine.

## Notes

- Insert queries are built with `fmt.Sprintf` string interpolation of the example values; the helpers are for controlled test data, not untrusted input.
- Table statistics are refreshed via `db.MetadataMgmt().ForceRefreshStatistics(tx)`. `InsertSmallTestData` and `InsertMiddleTestData` commit their transaction; `InsertLargeTestData` does not call `tx.Commit()` before returning.
- The commented cardinality table at the top documents the full-scale schema sizes the sample models; the actual inserted counts are much smaller.
- The package exports test helpers but contains no `_test.go` files of its own.
