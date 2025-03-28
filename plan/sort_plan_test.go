package plan_test

import (
	"path"
	"s1mpleasia.com/tinydb/plan"
	"s1mpleasia.com/tinydb/server"
	"s1mpleasia.com/tinydb/testlib"
	"testing"
)

func TestSortPlan(t *testing.T) {
	tinyDB, err := server.NewTinyDBWithMetadata(path.Join(".", "sort_plan_test"))
	if err != nil {
		t.Fatalf("failed to create tinyDB: %v", err)
	}

	err = testlib.InsertSmallTestData(t, tinyDB)
	if err != nil {
		t.Fatalf("failed to insert small test data: %v", err)
	}

	tx, err := tinyDB.NewTx()
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	p, err := plan.NewTablePlan(tx, "student", tinyDB.MetadataMgmt())
	if err != nil {
		t.Fatalf("failed to create table plan: %v", err)
	}

	sortPlan, err := plan.NewSortPlan(p, tx, []string{"gradyear", "sname"})
	if err != nil {
		t.Fatalf("failed to create sort plan: %v", err)
	}

	sortScan, err := sortPlan.Open()
	if err != nil {
		t.Fatalf("failed to open sorting plan: %v", err)
	}

	err = sortScan.BeforeFirst()
	if err != nil {
		t.Fatalf("failed to execute sortScan.BeforeFirst(): %v", err)
	}

	next := sortScan.Next()
	if !next {
		t.Fatal("No record")
	}

	gradyear, err := sortScan.GetInt("gradyear")
	if err != nil {
		t.Fatalf("failed to get gradyear: %v", err)
	}

	sname, err := sortScan.GetString("sname")
	if err != nil {
		t.Fatalf("failed to get sname: %v", err)
	}

	for {
		t.Logf("gradyear: %d, sname: %s", gradyear, sname)
		next = sortScan.Next()
		if !next {
			break
		}

		currentGradYear, err := sortScan.GetInt("gradyear")
		if err != nil {
			t.Fatalf("failed to get gradyear: %v", err)
		}

		currentSname, err := sortScan.GetString("sname")
		if err != nil {
			t.Fatalf("failed to get sname: %v", err)
		}

		if gradyear > currentGradYear {
			t.Fatalf("gradyear is not sorted: %d > %d", gradyear, currentGradYear)
		}

		if gradyear == currentGradYear && sname > currentSname {
			t.Fatalf("sname is not sorted: %v > %v", sname, currentSname)
		}

		gradyear = currentGradYear
		sname = currentSname
	}

	sortScan.Close()
	tx.Commit()
}
