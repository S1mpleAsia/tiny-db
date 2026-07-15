# btree — Standalone integer-keyed B-tree

A self-contained, file-backed B-tree over integer keys. It manages its own
fixed-size block storage on disk and is independent of the database engine's
transaction, record, and query layers.

## Responsibilities

- Store integer keys in a balanced B-tree persisted to a single file.
- Read and write nodes as fixed-size blocks, one node per block.
- Insert keys, splitting full children and growing the tree height when the root
  fills.
- Serialize a subtree to a JSON-like string for inspection.

## Key types

- `BTree` — tree metadata and storage: in-memory `Root`, the backing `*os.File`,
  `BlockSize`, `NumBlocks`, and `Degree` (minimum degree).
- `Node` — a single node: its `Block` number in the file, a slice of integer
  `Keys`, and a slice of child `Block` numbers (`Children`).

## Key API

- `NewBTree(fileName string) (*BTree, error)` — open/create the backing file and
  derive block size (4096), degree, and current block count from the file size.
- `(*BTree) Insert(key int) error` — insert a key, splitting a full root into a new
  root as needed, then descending to a leaf.
- `(*BTree) ReadNode(block int) (*Node, error)` / `WriteNode(node *Node) error` —
  deserialize/serialize a node to/from its block (little-endian; keys packed from
  offset 4, children from `BlockSize/2`).
- `(*BTree) SplitChild(parent *Node, index int) error` — split a full child around
  its median, promoting the median key into the parent.
- `(*BTree) FreeBlock() int` — return the next free block number.
- `(*BTree) Seek(block int) error` — position the file at a block offset.
- `(*BTree) Json(node *Node) (string, error)` — render a subtree as a JSON-like
  string of blocks, keys, and children.
- `NewNode() *Node` / `NewNodeWithKeys(keys []int) *Node` — construct empty or
  key-seeded nodes.

## How it fits

Depends only on the Go standard library (`os`, `encoding/binary`, `bytes`, `fmt`).
It has no dependents within the project.

## Notes

- Wiring status: this package is NOT wired into the database engine. Nothing in the
  SQL/query path uses it; it is a standalone data structure.
- This is one of two separate B-tree implementations in the repository. This
  top-level `btree` is a standalone integer-keyed structure with its own file
  format. It is distinct from `index/btree`, which is a disk-backed database index
  (`BTreeIndex`) implementing `query.Index` on top of the transaction/record layers
  and carrying RIDs; the two share no code.
- Keys are plain `int`; there is no value/RID payload, no delete or search-by-key
  method (only `Insert` and node I/O), and no duplicate-key handling.
- Tests: there are no test files in this package.
