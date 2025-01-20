package metadata_test

import (
	"fmt"
	"math/rand"
	"path"
	"testing"

	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/server"
)

func TestMetadata(t *testing.T) {
	db, err := server.NewTinyDBWithMetadata(path.Join(".", "tabletest"))
	if err != nil {
		panic(err)
	}

	mm := db.MetadataMgmt()
	tx, err := db.NewTx()
	if err != nil {
		panic(err)
	}

	sch := record.NewSchema()
	sch.AddIntField("A")
	sch.AddStringField("B", 9)

	// Part 1: Table metadata
	err = mm.CreateTable("mytable", sch, tx)
	if err != nil {
		panic(err)
	}

	layout, err := mm.GetLayout("mytable", tx)
	if err != nil {
		panic(err)
	}

	size := layout.SlotSize()
	sch2 := layout.Schema()

	fmt.Printf("mytable has slot size: %d\n", size)
	fmt.Println("its fields are: ")

	for _, fieldName := range sch2.Fields() {
		fieldType := ""
		if sch2.Type(fieldName) == record.INT {
			fieldType = "int"
		} else {
			strlen := sch2.Length(fieldName)
			fieldType = fmt.Sprintf("varchar(%d)", strlen)
		}
		fmt.Printf("%s: %s\n", fieldName, fieldType)
	}

	// Part 2: Statistics metadata
	ts, err := query.NewTableScan(tx, "mytable", layout)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 50; i++ {
		ts.Insert()

		n := rand.Int31n(50)
		ts.SetInt("A", n)
		ts.SetString("B", fmt.Sprintf("rec%d", n))
	}

	si, _ := mm.GetStatInfo("mytable", layout, tx)

	fmt.Printf("B(mytable) = %d\n", si.BlockAccessed())
	fmt.Printf("R(mytable) = %d\n", si.RecordsOutput())
	fmt.Printf("V(mytable, A) = %d\n", si.DistinctValues("A"))
	fmt.Printf("V(mytable, A) = %d\n", si.DistinctValues("B"))

	// Part 3: View metadata
	viewDef := "select B from mytable where A = 1"
	mm.CreateView("viewA", viewDef, tx)

	v, _ := mm.GetViewDef("viewA", tx)
	fmt.Printf("view def = %s\n", v)

	// Part 4: Index metadata
	err = mm.CreateIndex("indexA", "mytable", "A", tx)
	if err != nil {
		panic(err)
	}

	err = mm.CreateIndex("indexB", "mytable", "B", tx)
	if err != nil {
		panic(err)
	}

	indexMap, err := mm.GetIndexInfo("mytable", tx)
	if err != nil {
		panic(err)
	}

	ii := indexMap["A"]
	fmt.Printf("R(indexA) = %d\n", ii.RecordsOutput())
	fmt.Printf("V(indexA,A) = %d\n", ii.DistinctValues("A"))
	fmt.Printf("V(indexA,B) = %d\n", ii.DistinctValues("B"))

	ii = indexMap["B"]
	fmt.Printf("R(indexB) = %d\n", ii.RecordsOutput())
	fmt.Printf("V(indexB,A) = %d\n", ii.DistinctValues("A"))
	fmt.Printf("V(indexB,B) = %d\n", ii.DistinctValues("B"))

	tx.Commit()
}
