package concurrency

import "s1mpleasia.com/tinydb/file"

var lockTable = newLockTable()

// Similar to LockTable method but for transaction-specific. Each transaction for its own concurrency mgmt
type ConcurrencyManagement struct {
	locks map[file.BlockId]string
}

func NewConcurrencyMgmt() *ConcurrencyManagement {
	return &ConcurrencyManagement{
		locks: make(map[file.BlockId]string),
	}
}

func (m *ConcurrencyManagement) SLock(block file.BlockId) error {
	if m.locks[block] != "" {
		return nil
	}

	err := lockTable.SLock(block)

	if err != nil {
		return err
	}

	m.locks[block] = "S"
	return nil
}

func (m *ConcurrencyManagement) XLock(block file.BlockId) error {
	if m.hasXLock(block) {
		return nil
	}

	err := m.SLock(block)
	if err != nil {
		return err
	}

	err = lockTable.XLock(block)
	if err != nil {
		return err
	}

	m.locks[block] = "X"
	return nil
}

func (m *ConcurrencyManagement) Release() {
	for block := range m.locks {
		lockTable.Unlock(block)
	}

	clear(m.locks)
}

func (m *ConcurrencyManagement) hasXLock(block file.BlockId) bool {
	lockType := m.locks[block]
	return lockType == "X"
}
