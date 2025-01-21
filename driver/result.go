package driver

import "database/sql/driver"

var _ driver.Result = (*TinyDBResult)(nil)

type TinyDBResult struct {
	rowsAffected int
}

func NewResult(rowsAffected int) *TinyDBResult {
	return &TinyDBResult{rowsAffected}
}

func (result *TinyDBResult) LastInsertId() (int64, error) {
	panic("unimplemented")
}

func (result *TinyDBResult) RowsAffected() (int64, error) {
	return int64(result.rowsAffected), nil
}
