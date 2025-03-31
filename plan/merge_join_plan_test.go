package plan_test

import (
	"path"
	"s1mpleasia.com/tinydb/plan"
	"s1mpleasia.com/tinydb/server"
	"s1mpleasia.com/tinydb/testlib"
	"testing"
)

func TestMergeJoinPlan(t *testing.T) {
	tinyDB, err := server.NewTinyDBWithMetadata(path.Join(".", "merge_join_plan_test"))
	if err != nil {
		t.Fatalf("failed to create tinyDB: %v", err)
	}

	err = testlib.InsertSmallTestData(t, tinyDB)
	if err != nil {
		t.Fatalf("failed to insert small test data: %v", err)
	}

	tx, err := tinyDB.NewTx()
	if err != nil {
		t.Fatalf("failed to create tx: %v", err)
	}

	p1, err := plan.NewTablePlan(tx, "dept", tinyDB.MetadataMgmt())
	if err != nil {
		t.Fatalf("failed to create plan of dept: %v", err)
	}

	p2, err := plan.NewTablePlan(tx, "student", tinyDB.MetadataMgmt())
	if err != nil {
		t.Fatalf("failed to create plan of student: %v", err)
	}

	mergeJoinPlan, err := plan.NewMergeJoinPlan(tx, p1, p2, "did", "majorid")
	if err != nil {
		t.Fatalf("failed to create merge join plan: %v", err)
	}

	mergeJoinScan, err := mergeJoinPlan.Open()
	if err != nil {
		t.Fatalf("failed to open merge join scan: %v", err)
	}

	err = mergeJoinScan.BeforeFirst()
	if err != nil {
		t.Fatalf("mergeJoinScan.BeforeFirst(): %v", err)
	}

	expects := map[string]string{
		"joe": "compsci",
		"amy": "math",
		"max": "compsci",
		"sue": "math",
		"bob": "drama",
		"kim": "math",
		"art": "drama",
		"pat": "compsci",
		"lee": "compsci",
		"dan": "math",
	}

	for {
		next := mergeJoinScan.Next()

		if !next {
			break
		}

		sname, err := mergeJoinScan.GetString("sname")
		if err != nil {
			t.Fatalf("failed to get sname: %v", err)
		}

		dname, err := mergeJoinScan.GetString("dname")
		if err != nil {
			t.Fatalf("failed to get dname: %v", err)
		}

		t.Logf("sname: %s, dname: %s", sname, dname)
		expect, ok := expects[sname]
		if !ok {
			t.Fatalf("failed to find expect for sname: %s", sname)
		}

		if dname != expect {
			t.Errorf("expect: %s, got: %s", expect, dname)
		}
	}

	mergeJoinScan.Close()
	tx.Commit()
}
