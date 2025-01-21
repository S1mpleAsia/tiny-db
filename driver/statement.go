package driver

import (
	"context"
	"database/sql/driver"

	"s1mpleasia.com/tinydb/plan"
	"s1mpleasia.com/tinydb/transaction"
)

var _ driver.Stmt = (*TinyDBStmt)(nil)
var _ driver.StmtExecContext = (*TinyDBStmt)(nil)
var _ driver.StmtQueryContext = (*TinyDBStmt)(nil)

type TinyDBStmt struct {
	query   string
	planner *plan.Planner
	tx      *transaction.Transaction
}

func NewTinyDBStmt(query string, planner *plan.Planner, tx *transaction.Transaction) *TinyDBStmt {
	return &TinyDBStmt{
		query:   query,
		planner: planner,
		tx:      tx,
	}
}

func (s *TinyDBStmt) Exec(args []driver.Value) (driver.Result, error) {
	panic("unimplemented")
}

func (s *TinyDBStmt) Close() error {
	return nil
}

func (s *TinyDBStmt) NumInput() int {
	return 0
}

func (s *TinyDBStmt) Query(args []driver.Value) (driver.Rows, error) {
	panic("unimplemented")
}

func (s *TinyDBStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	rows, err := s.planner.ExecuteUpdate(s.query, s.tx)
	if err != nil {
		return nil, err
	}

	return NewResult(rows), nil

}

func (s *TinyDBStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	plan, err := s.planner.CreateQueryPlan(s.query, s.tx)
	if err != nil {
		return nil, err
	}

	scan, err := plan.Open()
	if err != nil {
		return nil, err
	}

	return NewTinyDbRows(plan.Schema(), scan), nil
}
