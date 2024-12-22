package transaction_test

import (
	"path"
	"testing"

	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/server"
	"s1mpleasia.com/tinydb/transaction"
)

func TestTransaction(t *testing.T) {
	db, err := server.NewTinyDB(path.Join(".", "txtest"), 400, 8)

	if err != nil {
		t.Fatal(err)
	}

	fm := db.FileMgmt()
	lm := db.LogMgmt()
	bm := db.BufferMgmt()

	tx1, err := transaction.NewTransaction(fm, lm, bm)
	if err != nil {
		t.Fatal(err)
	}

	block := file.NewBlockId("testfile", 1)
	tx1.Pin(block)
	tx1.SetInt(block, 80, 1, false) // Do not log initial block values
	tx1.SetString(block, 40, "one", false)
	tx1.Commit()

	tx2, err := transaction.NewTransaction(fm, lm, bm)
	if err != nil {
		t.Fatal(err)
	}

	tx2.Pin(block)
	ival, _ := tx2.GetInt(block, 80)
	sval, _ := tx2.GetString(block, 40)

	if ival != 1 {
		t.Fatalf("Expected 1, got %d", ival)
	}

	if sval != "one" {
		t.Fatalf("Expecteed one, got %s", sval)
	}

	t.Logf("Initital value at location 80 = %d\n", ival)
	t.Logf("Initital value at location 40 = %s\n", sval)

	newIVal := ival + 1
	newSVal := sval + "!"
	tx2.SetInt(block, 80, newIVal, true)
	tx2.SetString(block, 40, newSVal, true)
	tx2.Commit()

	tx3, err := transaction.NewTransaction(fm, lm, bm)
	if err != nil {
		t.Fatal(err)
	}
	tx3.Pin(block)

	ival, _ = tx3.GetInt(block, 80)
	sval, _ = tx3.GetString(block, 40)

	if ival != 2 {
		t.Fatalf("expected one!, got %d", ival)

	}

	if sval != "one!" {
		t.Fatalf("expected one!, got %s", sval)
	}

	t.Logf("Initial value at location 80 = %d\n", ival)
	t.Logf("Initial value at location 40 = %s\n", sval)

	tx3.SetInt(block, 80, 9999, true)
	ival, _ = tx3.GetInt(block, 80)

	t.Logf("pre-rollback value at location 80 = %d\n", ival)
	tx3.Rollback()

	tx4, err := transaction.NewTransaction(fm, lm, bm)
	if err != nil {
		t.Fatal(err)
	}

	tx4.Pin(block)
	ival, _ = tx4.GetInt(block, 80)
	t.Logf("post-rollback value at location 80 = %d\n", ival)
	if ival != 2 {
		t.Fatalf("expected 2, got %d", ival)
	}
	tx4.Commit()
}
