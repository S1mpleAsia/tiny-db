package driver

import (
	"database/sql/driver"
	"io"

	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
)

var _ driver.Rows = (*TinyDbRows)(nil)

type TinyDbRows struct {
	schema *record.Schema
	scan   query.Scan
}

func NewTinyDbRows(schema *record.Schema, scan query.Scan) *TinyDbRows {
	return &TinyDbRows{schema, scan}
}

func (r *TinyDbRows) Columns() []string {
	return r.schema.Fields()
}

func (r *TinyDbRows) Close() error {
	r.scan.Close()
	return nil
}

func (r *TinyDbRows) Next(dest []driver.Value) error {
	ok := r.scan.Next()
	if !ok {
		return io.EOF
	}

	for i, col := range r.Columns() {
		val, err := r.scan.GetVal(col)
		if err != nil {
			return err
		}

		dest[i] = val.AnyValue()
	}

	return nil
}
