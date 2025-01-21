package driver

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"s1mpleasia.com/tinydb/server"
)

var _ driver.Driver = (*TinyDBDriver)(nil)

type TinyDBDriver struct {
}

func init() {
	sql.Register("tinydb", &TinyDBDriver{})
}

func (d *TinyDBDriver) Open(name string) (driver.Conn, error) {
	fmt.Println("Opening database")
	db, err := server.NewTinyDBWithMetadata(name)
	if err != nil {
		return nil, err
	}

	conn := NewConnection(db, db.Planner())
	fmt.Println("database opended")

	return conn, nil
}
