# record — Record storage and slotted pages

Defines how table records are described (schema, field types), how they are laid out in a page slot, and how records are read from and written to slotted record pages. Also provides typed value wrappers and record identifiers.

## Responsibilities

- Describe table structure via `Schema` (field names, types, lengths).
- Compute physical field offsets and slot size via `Layout`.
- Store and retrieve records in a page using a slotted, fixed-size-slot format (`RecordPage`).
- Track and manipulate empty/used slots (insert, delete, iterate).
- Wrap typed values (int or string) for comparison and hashing (`Constant`).
- Identify a record by block number and slot (`RID`).

## Key types

- `FieldType` — enum for field types: `INT` and `VARCHAR`.
- `Schema` — ordered list of field names plus per-field `FieldInfo` (type and length).
- `FieldInfo` — type and length metadata for a single field.
- `Layout` — maps each field to a byte offset within a slot and holds the total slot size, derived from a `Schema`.
- `RecordFlag` — per-slot status flag: `EMPTY` or `USED`.
- `RecordPage` — a transaction-backed view over one block, addressing records by slot.
- `RID` — record identifier: (block number, slot).
- `Constant` — a value holding either an int32 or a string, with comparison, equality, and hashing.

## Key API

- `NewSchema() *Schema`; `AddField(name string, t FieldType, length int32)`, `AddIntField(name string)`, `AddStringField(name string, length int32)`, `Add(name string, sch *Schema)`, `AddAll(sch *Schema)` — build a schema.
- `(*Schema) Fields() []string`, `HasField(name string) bool`, `Type(name string) FieldType`, `Length(name string) int32` — inspect a schema (`Type`/`Length` panic on unknown field).
- `NewLayoutFromSchema(sch *Schema) *Layout` — compute offsets and slot size from a schema; `NewLayout(sch, offset, slotSize)` reconstructs a layout from stored values.
- `(*Layout) Schema() *Schema`, `Offset(name string) int32`, `SlotSize() int32`.
- `NewRecordPage(tx *transaction.Transaction, block *file.BlockId, layout *Layout) (*RecordPage, error)` — pins the block.
- `(*RecordPage) GetInt/GetString(slot int32, field string)`, `SetInt/SetString(slot, field, val)` — typed field access.
- `(*RecordPage) Format() error` — initialize all slots in the block to `EMPTY` with zeroed fields (unlogged writes).
- `(*RecordPage) InsertAfter(slot int32) (int32, error)` — find and mark the next empty slot as `USED`; returns -1 if none.
- `(*RecordPage) NextAfter(slot int32) (int32, error)` — find the next `USED` slot; returns -1 if none.
- `(*RecordPage) Delete(slot int32) error` — mark a slot `EMPTY`.
- `(*RecordPage) Block() *file.BlockId`.
- `NewRID(blockNum, slot int32) *RID`; `(*RID) BlockNumber()`, `Slot()`, `Equals(*RID) bool`, `String()`.
- `NewConstantWithInt(int32)`, `NewConstantWithString(string)`; `(*Constant) AsInt()`, `AsString()`, `Equals(*Constant)`, `CompareTo(*Constant)`, `HashCode()`, `AnyValue()`, `String()`.

## How it fits

Depends on the `file` package (block size, byte length calculations) and, for `RecordPage`, on the `transaction` package for pinned, logged reads and writes. It is a foundational layer used by the `query` (table scans), `metadata` (catalog layouts), and index packages.

## Notes

- Slots are fixed-size. Each slot begins with a 4-byte flag (`INT_32_BITS`), followed by the fields in schema order.
- A `VARCHAR` field reserves space based on its declared length via `file.MaxLength`.
- `Format` writes with logging disabled; normal `SetInt`/`SetString` are logged.
- `Constant` stores exactly one of int/string; `CompareTo` returns `ErrInvalidConstantType` when the two operands are not the same type, and string hashing uses FNV-32.
- `Schema.Type`/`Schema.Length` panic if the field does not exist.
- Test coverage: `record_test.go` (`TestRecord`) exercises schema/layout construction and insert/delete/scan on a `RecordPage`.
