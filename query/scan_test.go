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

func TestScan(t *testing.T) {
	db, err := server.NewTinyDBWithMetadata(path.Join(".", "scantest"))
	if err != nil {
		t.Fatalf("failed to create tinydb: %v", err)
	}

	tx, err := db.NewTx()
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	sch := record.NewSchema()
	sch.AddIntField("A")
	sch.AddStringField("B", 9)
	layout := record.NewLayoutFromSchema(sch)

	s1, err := record.NewTableScan(tx, "my_table", layout)
	if err != nil {
		t.Fatalf("failed to create table scan: %v", err)
	}

	n := 200
	fmt.Printf("Inserting %d random records\n", n)

	for i := 0; i < n; i++ {
		if err := s1.Insert(); err != nil {
			t.Fatalf("failed to insert record: %v", err)
		}

		k := rand.Int31n(50)
		if err := s1.SetInt("A", k); err != nil {
			t.Fatalf("failed to setInt: %v", err)
		}

		if err := s1.SetString("B", fmt.Sprintf("rec%d", k)); err != nil {
			t.Fatalf("failed to setString: %v", err)
		}

		fmt.Printf("Inserted record %d: A: %d, B: %s\n", i+1, k, fmt.Sprintf("rec%d", k))
	}

	s1.Close()

	// Selecting all records where A = 10
	s2, err := record.NewTableScan(tx, "my_table", layout)
	if err != nil {
		t.Fatalf("failed to create table scan: %v", err)
	}

	c := record.NewConstantWithInt(10)
	lhs := query.NewExpressionWithField("A")
	rhs := query.NewExpressionWithConstant(c)
	term := query.NewTerm(lhs, rhs)
	pred := query.NewPredicateWithTerm(term)

	fmt.Printf("the predicate is %v\n", pred)

	s3 := query.NewSelectScan(s2, pred)

	fields := []string{"B"}
	s4 := query.NewProjectScan(s3, fields)

	for s4.Next() {
		b, err := s4.GetString("B")
		if err != nil {
			t.Fatalf("GetString failed: %v", err)
		}

		fmt.Printf("B: %s\n", b)

		if b != "rec10" {
			t.Fatalf("expected rec10, got %s", b)
		}
	}

	s4.Close()
	tx.Commit()
}

func TestJoinAndSelect(t *testing.T) {
	db, err := server.NewTinyDBWithMetadata(path.Join(".", "scantest"))
	if err != nil {
		t.Fatalf("failed to create tinydb: %v", err)
	}

	tx, err := db.NewTx()
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	sch := record.NewSchema()
	sch.AddIntField("A")
	sch.AddStringField("B", 9)
	layout := record.NewLayoutFromSchema(sch)

	ts1, err := record.NewTableScan(tx, "T1", layout)
	if err != nil {
		t.Fatalf("failed to create table scan: %v", err)
	}

	if err := ts1.BeforeFirst(); err != nil {
		t.Fatalf("failed to call BeforeFirst: %v", err)
	}

	n := 200
	fmt.Printf("Inserting %d random records into T1\n", n)

	for i := 0; i < n; i++ {
		if err := ts1.Insert(); err != nil {
			t.Fatalf("failed to insert record: %v", err)
		}

		if err := ts1.SetInt("A", int32(i)); err != nil {
			t.Fatalf("failed to setInt: %v", err)
		}

		if err := ts1.SetString("B", fmt.Sprintf("%d", i)); err != nil {
			t.Fatalf("failed to setString: %v", err)
		}

		fmt.Printf("Inserted record %d: A: %d, B: %s\n", i+1, i, fmt.Sprintf("%d", i))
	}

	ts1.Close()

	sch2 := record.NewSchema()
	sch2.AddIntField("C")
	sch2.AddStringField("D", 9)
	layout2 := record.NewLayoutFromSchema(sch2)

	ts2, err := record.NewTableScan(tx, "T2", layout2)
	if err != nil {
		t.Fatalf("failed to create table scan: %v", err)
	}

	if err := ts2.BeforeFirst(); err != nil {
		t.Fatalf("failed to call BeforeFirst: %v", err)
	}

	fmt.Printf("Inserting %d random records into T2\n", n)

	for i := 0; i < n; i++ {
		if err := ts2.Insert(); err != nil {
			t.Fatalf("failed to insert record: %v", err)
		}

		if err := ts2.SetInt("C", int32(n-i-1)); err != nil {
			t.Fatalf("failed to setInt: %v", err)
		}

		if err := ts1.SetString("D", fmt.Sprintf("%d", n-i-1)); err != nil {
			t.Fatalf("failed to setString: %v", err)
		}

		fmt.Printf("Inserted record %d: A: %d, B: %s\n", i+1, n-i-1, fmt.Sprintf("%d", n-i-1))
	}

	ts2.Close()

	s1, _ := record.NewTableScan(tx, "T1", layout)
	s2, _ := record.NewTableScan(tx, "T2", layout2)

	s3, _ := query.NewProductScan(s1, s2)

	lhs := query.NewExpressionWithField("A")
	rhs := query.NewExpressionWithField("C")
	term := query.NewTerm(lhs, rhs)
	pred := query.NewPredicateWithTerm(term)

	fmt.Printf("the predicate is %v\n", pred)
	s4 := query.NewSelectScan(s3, pred)

	fields := []string{"B", "D"}
	s5 := query.NewProjectScan(s4, fields)

	for s5.Next() {
		b, _ := s5.GetString("B")
		d, _ := s5.GetString("D")

		fmt.Printf("B: %s, D: %s\n", b, d)

		if b != d {
			t.Fatalf("expected B=D, got B=%s, D=%s", b, d)
		}
	}

	s5.Close()
	tx.Commit()
}
