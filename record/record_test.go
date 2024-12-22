package record_test

import (
	"fmt"
	"math/rand"
	"path"
	"strconv"
	"testing"

	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/server"
)

func TestRecord(t *testing.T) {
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
		fmt.Printf("%q has offset %v\n", fieldName, offset)
	}

	block := tx.Append("testfile")
	err = tx.Pin(block)
	if err != nil {
		t.Fatal(err)
	}

	rp, err := record.NewRecordPage(tx, block, layout)
	if err != nil {
		t.Fatal(err)
	}

	rp.Format()
	fmt.Println("Filling the page with random records")
	slot, err := rp.InsertAfter(-1)
	if err != nil {
		t.Fatal(err)
	}

	for slot >= 0 {
		n := rand.Intn(50)
		err = rp.SetInt(slot, "A", int32(n))
		if err != nil {
			t.Fatal(err)
		}

		err = rp.SetString(slot, "B", "rec"+strconv.Itoa(n))
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("insert into slot %d: {A: %d, B: %q}\n", slot, n, "rec"+strconv.Itoa(n))
		slot, err = rp.InsertAfter(slot)
		if err != nil {
			t.Fatal(err)
		}
	}

	fmt.Println("deleted these records with A-values < 25")

	count := 0
	slot, err = rp.NextAfter(-1)
	if err != nil {
		t.Fatal(err)
	}

	for slot >= 0 {
		a, _ := rp.GetInt(slot, "A")
		b, err := rp.GetString(slot, "B")
		if err != nil {
			t.Fatal(err)
		}

		if a < 25 {
			count++
			fmt.Printf("deleted slot %d: {A: %d, B: %q}\n", slot, a, b)
			rp.Delete(slot)
		}

		slot, err = rp.NextAfter(slot)
		if err != nil {
			t.Fatal(err)
		}
	}

	fmt.Printf("%d values under 25 were deleted.\n", count)
	fmt.Println("here is the remaining records")

	slot, err = rp.NextAfter(-1)
	if err != nil {
		t.Fatal(err)
	}

	for slot >= 0 {
		a, _ := rp.GetInt(slot, "A")
		b, err := rp.GetString(slot, "B")
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("slot %d: {A: %d, B: %q}\n", slot, a, b)
		slot, err = rp.NextAfter(slot)
		if err != nil {
			t.Fatal(err)
		}
	}

	tx.Unpin(block)
	tx.Commit()
}
