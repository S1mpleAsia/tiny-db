package driver

import (
	"context"
	"database/sql/driver"

	"s1mpleasia.com/tinydb/plan"
	"s1mpleasia.com/tinydb/server"
	"s1mpleasia.com/tinydb/transaction"
)

var _ driver.Conn = (*Connection)(nil)
var _ driver.ConnBeginTx = (*Connection)(nil)

type Connection struct {
	db      *server.TinyDB
	tx      *ConnectionTransaction
	planner *plan.Planner
}

func NewConnection(db *server.TinyDB, planner *plan.Planner) *Connection {
	return &Connection{db, nil, planner}
}

// Deprecated
func (conn *Connection) Begin() (driver.Tx, error) {
	return conn.BeginTx(context.TODO(), driver.TxOptions{})
}

func (conn *Connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	tx, err := conn.db.NewTx()
	if err != nil {
		return nil, err
	}

	conn.tx = NewTransactionWithConnect(tx, conn)
	return conn.tx, nil
}

func (conn *Connection) Close() error {
	conn.tx.Commit()
	return nil
}

func (conn *Connection) Prepare(query string) (driver.Stmt, error) {
	return &TinyDBStmt{
		query:   query,
		planner: conn.planner,
		tx:      conn.tx.tx,
	}, nil
}

var _ driver.Tx = (*ConnectionTransaction)(nil)

type ConnectionTransaction struct {
	tx   *transaction.Transaction
	conn *Connection
}

func NewTransactionWithConnect(tx *transaction.Transaction, conn *Connection) *ConnectionTransaction {
	return &ConnectionTransaction{tx, conn}
}

func (t *ConnectionTransaction) Commit() error {
	t.tx.Commit()
	t.conn.tx = nil

	return nil
}

func (t *ConnectionTransaction) Rollback() error {
	t.tx.Rollback()
	t.conn.tx = nil

	return nil
}
