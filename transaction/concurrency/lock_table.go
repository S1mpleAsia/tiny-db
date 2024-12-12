package concurrency

import (
	"fmt"
	"sync"
	"time"

	"s1mpleasia.com/tinydb/file"
)

const maxLockTime = 10 * time.Second

var ErrTimeout = fmt.Errorf("timeout error")

// Excluse lock -> locks[block] < 0
// Share lock -> locks[block] >= 0
type LockTable struct {
	locks 	map[file.BlockId]int
	cond	*sync.Cond
}

func newLockTable() *LockTable {
	return &LockTable{
		locks: make(map[file.BlockId]int),
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (lt *LockTable) SLock(block file.BlockId) error {
	lt.cond.L.Lock()
	defer lt.cond.L.Unlock()

	startTime := time.Now()

	for {
		if time.Since(startTime) > maxLockTime {
			return ErrTimeout
		} else if !lt.hasXLock(block) {
			break
		}

		lt.waitWithTimeout(maxLockTime)
	}

	lt.locks[block]++
	return nil
}

func (lt *LockTable) XLock(block file.BlockId) error {
	lt.cond.L.Lock()
	defer lt.cond.L.Unlock()

	startTime := time.Now()

	for {
		if time.Since(startTime) > maxLockTime {
			return ErrTimeout
		} else if !lt.hasOtherSLocks(block) {
			break
		}

		lt.waitWithTimeout(maxLockTime)
	}

	lt.locks[block] = -1
	return nil
}

func (lt *LockTable) Unlock(block file.BlockId) {
	lt.cond.L.Lock()
	defer lt.cond.L.Unlock()

	if lt.locks[block] > 1 {
		lt.locks[block]--
	} else {
		delete(lt.locks, block)
		lt.cond.Broadcast()
	}
}

func (lt *LockTable) hasXLock(block file.BlockId) bool {
	return lt.locks[block] < 0 
}

func (lt *LockTable) hasOtherSLocks(block file.BlockId) bool {
	return lt.locks[block] > 1
}

// Equivalent to wait(MAX_TIME) in Java
func (lt *LockTable) waitWithTimeout(timeout time.Duration) {
	timer := time.AfterFunc(timeout, func() {
		lt.cond.L.Lock()
		defer lt.cond.L.Unlock()
		lt.cond.Broadcast()
	})	

	lt.cond.Wait()
	timer.Stop()
}