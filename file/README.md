# file — Block and page I/O

The `file` package is the lowest layer of the storage stack. It manages fixed-size disk blocks, in-memory pages, and the primitive read/write operations that move data between them.

## Responsibilities

- Identify disk blocks by file name and logical block number.
- Represent the contents of a block as an in-memory page.
- Serialize and deserialize primitive values (int, string, byte blobs) into a page buffer.
- Read blocks from disk into pages and write pages back to disk.
- Append new (empty) blocks to a file and report a file's length in blocks.
- Manage the pool of open OS file handles for the database directory.

## Key types

- `BlockId` — identifies a specific block by its file name and 0-indexed block number.
- `Page` — a fixed-size byte buffer holding the contents of one disk block; supports typed reads/writes at arbitrary offsets.
- `FileMgmt` — owns the database directory, block size, and open file handles; performs block-level disk I/O.

## Key API

- `NewBlockId(fileName string, blockNum int64) *BlockId` — creates a block identifier.
- `(*BlockId) FileName() string`, `(*BlockId) BlockNumber() int64` — accessors.
- `(*BlockId) Equals(other *BlockId) bool` — value equality by file name and block number.
- `NewPage(blockSize int64) *Page` — allocates a zeroed page of the given size.
- `NewPageWith(buffer []byte) *Page` — wraps an existing byte slice as a page.
- `(*Page) GetInt(offset int) int32` / `SetInt(offset int, val int32)` — read/write a little-endian 4-byte int.
- `(*Page) GetBytes(offset int) []byte` / `SetBytes(offset int, val []byte)` — read/write a length-prefixed blob (4-byte length followed by the bytes).
- `(*Page) GetString(offset int) string` / `SetString(offset int, val string)` — read/write a length-prefixed UTF-16 string.
- `MaxLength(length int) int` — bytes needed to store a string of `length` characters (`INT_32_BITS + length*UTF_16_SIZE`).
- `NewFileMgmt(dbDir string, blockSize int64) (*FileMgmt, error)` — opens/creates the database directory and clears any `temp*` files.
- `(*FileMgmt) Read(block *BlockId, p *Page) error` — reads a block from disk into a page.
- `(*FileMgmt) Write(block *BlockId, p *Page) error` — writes a page to its block on disk.
- `(*FileMgmt) Append(fileName string) (*BlockId, error)` — extends the file by one empty block and returns its `BlockId`.
- `(*FileMgmt) Length(fileName string) (int64, error)` — number of blocks in a file.
- `(*FileMgmt) BlockSize() int64`, `(*FileMgmt) IsNew() bool` — accessors.

## How it fits

This package depends only on the standard library. It is the foundation used by `log` (which stores log records in pages/blocks) and `buffer` (which caches pages in memory), and transitively by the rest of the storage engine.

## Notes

- Integers are stored little-endian in 4 bytes; strings are stored as UTF-16 code units with a 4-byte byte-length prefix.
- `SetString` writes the length in bytes (`len(runes)*UTF_16_SIZE`); `GetString` divides by `UTF_16_SIZE` to recover the character count.
- `FileMgmt.openFile` (unexported) caches `*os.File` handles keyed by name and opens with `O_RDWR|O_CREATE`.
- The `temp*` cleanup in `NewFileMgmt` calls `os.Remove(file.Name())` with the bare entry name rather than a path joined with `dbDir`.
- Tested in `file_test.go`, which round-trips a string and an int through a page written to and read from disk.
