package metadata_test

import (
	"fmt"
	"path"
	"testing"

	"s1mpleasia.com/tinydb/metadata"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/server"
)

func TestTableMgmt(t *testing.T) {
	t.Parallel()

	db, err := server.NewTinyDB(path.Join(".", "tabletest"), 400, 8)
	if err != nil {
		fmt.Println(err)
	}

	tx, err := db.NewTx()
	if err != nil {
		panic(err)
	}

	tm, err := metadata.NewTableMgmt(true, tx)
	if err != nil {
		panic(err)
	}

	sch := record.NewSchema()
	sch.AddIntField("A")
	sch.AddStringField("B", 9)

	if err = tm.CreateTable("mytable", sch, tx); err != nil {
		panic(err)
	}

	layout, err := tm.GetLayout("mytable", tx)
	if err != nil {
		panic(err)
	}

	size := layout.SlotSize()
	sch2 := layout.Schema()

	if size != 30 {
		fmt.Printf("slotSize: expected 30, got %d\n", size)
		return
	}

	fmt.Printf("mytable has slotsize: %d\n", size)
	fmt.Println("Its fields are: ")

	for _, fieldName := range sch2.Fields() {
		var t string = ""

		if sch2.Type(fieldName) == record.INT {
			t = "int"
		} else {
			strlen := sch2.Length(fieldName)
			t = fmt.Sprintf("varchar(%d)", strlen)
		}

		fmt.Printf("%s: %s\n", fieldName, t)
	}

	tx.Commit()
}

func TestCatalog(t *testing.T) {
	t.Parallel()

	db, err := server.NewTinyDB(path.Join(".", "tabletest"), 400, 8)
	if err != nil {
		fmt.Println(err)
	}

	tx, err := db.NewTx()
	if err != nil {
		panic(err)
	}

	tm, err := metadata.NewTableMgmt(true, tx)
	if err != nil {
		panic(err)
	}

	sch := record.NewSchema()
	sch.AddIntField("A")
	sch.AddStringField("B", 9)

	if err = tm.CreateTable("mytable", sch, tx); err != nil {
		panic(err)
	}
	fmt.Println("All tables and their lengths: ")

	layout, err := tm.GetLayout(metadata.TBL_CATALOG_FILE, tx)
	if err != nil {
		panic(err)
	}

	ts, err := query.NewTableScan(tx, metadata.TBL_CATALOG_FILE, layout)
	if err != nil {
		panic(err)
	}

	for ts.Next() {
		tblName, err := ts.GetString(metadata.TBL_CATALOG_TABLE_NAME)
		if err != nil {
			panic(err)
		}
		size, err := ts.GetInt(metadata.TBL_CATALOG_SLOT_SIZE)
		if err != nil {
			panic(err)
		}
		fmt.Printf("table: %s - size: %d\n", tblName, size)
	}

	ts.Close()

	fmt.Println("All fields and their offsets: ")
	layout, err = tm.GetLayout(metadata.FIELD_CATALOG_FILE, tx)
	if err != nil {
		panic(err)
	}

	ts, err = query.NewTableScan(tx, metadata.FIELD_CATALOG_FILE, layout)
	if err != nil {
		panic(err)
	}

	for ts.Next() {
		tblName, err := ts.GetString(metadata.FIELD_CATALOG_TABLE_NAME)
		if err != nil {
			panic(err)
		}

		fieldName, err := ts.GetString(metadata.FIELD_CATALOG_FIELD_NAME)
		if err != nil {
			panic(err)
		}

		offset, err := ts.GetInt(metadata.FIELD_CATALOG_OFFSET)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s.%s - offset: %d\n", tblName, fieldName, offset)
	}
	ts.Close()
}
