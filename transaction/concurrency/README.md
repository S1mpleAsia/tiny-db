# concurrency — Lock-based concurrency control

Provides block-granular shared/exclusive locking to enforce transaction isolation, with a per-transaction lock manager layered over a process-wide lock table.

## Responsibilities

- Track the locks held by a single transaction and avoid re-requesting locks it already owns.
- Acquire shared (read) and exclusive (write) locks on blocks, upgrading from shared to exclusive as needed.
- Release all of a transaction's locks at once.
- Maintain the global lock table that grants, queues, and releases locks across transactions.
- Detect lock-wait timeouts and surface them as errors.

## Key types

- `ConcurrencyManagement` — per-transaction lock manager; maps each block to the lock type ("S" or "X") the transaction holds.
- `LockTable` — process-wide table mapping each block to a lock count (negative = exclusive, positive = number of shared holders), synchronized with a condition variable.
- `ErrTimeout` — sentinel error returned when a lock cannot be acquired within the maximum wait time.

## Key API

- `NewConcurrencyMgmt() *ConcurrencyManagement` — creates a lock manager for a transaction.
- `(*ConcurrencyManagement) SLock(block file.BlockId) error` — acquires a shared lock (no-op if any lock is already held on the block).
- `(*ConcurrencyManagement) XLock(block file.BlockId) error` — acquires a shared lock then upgrades to exclusive (no-op if already exclusive).
- `(*ConcurrencyManagement) Release()` — unlocks every block held and clears the transaction's lock set.
- `(*LockTable) SLock(block file.BlockId) error` — waits until no exclusive lock exists, then increments the shared count.
- `(*LockTable) XLock(block file.BlockId) error` — waits until no other shared locks exist, then marks the block exclusively locked.
- `(*LockTable) Unlock(block file.BlockId)` — decrements the shared count or removes the lock and broadcasts to waiters.

## How it fits

Depends only on `file` (for `BlockId`). Used by the `transaction` package, where each transaction owns a `ConcurrencyManagement` that guards `Get*`/`Set*`/`Size`/`Append` operations. The `LockTable` is a single package-global instance shared by all transactions.

## Notes

- Deadlock is handled indirectly: a lock request that waits longer than `maxLockTime` (10 seconds) returns `ErrTimeout` rather than blocking forever.
- `XLock` at the lock-table level relies on the caller already holding the shared lock, so it waits only for *other* shared holders (count greater than 1).
- Waiting uses a condition variable with a timer that broadcasts on timeout to wake blocked waiters.
- No test files are present in this package; concurrency tests live in the parent `transaction` package.
