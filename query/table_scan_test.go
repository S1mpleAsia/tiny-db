package query_test

import (
	"fmt"
	"math/rand"
	"path"
	"testing"

	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/server"
)

func TestTableScan(t *testing.T) {
	t.Parallel()

	db, err := server.NewTinyDB(path.Join(".", "recordtest"), 400, 8)
	if err != nil {
		fmt.Println(err)
	}

	tx, err := db.NewTx()
	if err != nil {
		panic(err)
	}

	schema := record.NewSchema()
	schema.AddIntField("A")
	schema.AddStringField("B", 9)
	layout := record.NewLayoutFromSchema(schema)

	for _, fieldName := range layout.Schema().Fields() {
		offset := layout.Offset(fieldName)
		fmt.Printf("%s has offset %d\n", fieldName, offset)
	}

	fmt.Println("filling the table with 50 random records")
	ts, err := query.NewTableScan(tx, "T", layout)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 50; i++ {
		ts.Insert()
		n := rand.Int31n(50)
		ts.SetInt("A", n)
		ts.SetString("B", fmt.Sprintf("rec%d", n))
		fmt.Printf("inserting into slot %s: {%d, %s}\n", ts.GetRID().String(), n, fmt.Sprintf("rec%d", n))
	}

	fmt.Println("deleting records with A-values < 25")
	cnt := 0
	if err = ts.BeforeFirst(); err != nil {
		panic(err)
	}

	for ts.Next() {
		a, _ := ts.GetInt("A")
		b, err := ts.GetString("B")

		if err != nil {
			panic(err)
		}

		if a < 25 {
			cnt++
			fmt.Printf("deleting record from slot %s: {%d, %s}\n", ts.GetRID().String(), a, b)
			ts.Delete()
		}
	}

	fmt.Printf("%d values under 25 were deleted\n", cnt)

	ts.BeforeFirst()

	for ts.Next() {
		a, _ := ts.GetInt("A")
		b, err := ts.GetString("B")

		if err != nil {
			panic(err)
		}

		fmt.Printf("slot %s: {%d, %s}\n", ts.GetRID().String(), a, b)
	}

	ts.Close()
	tx.Commit()
}
