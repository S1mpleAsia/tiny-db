package plan_test

import (
	"path"
	"s1mpleasia.com/tinydb/plan"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/server"
	"s1mpleasia.com/tinydb/testlib"
	"testing"
)

func TestGroupByPlan(t *testing.T) {
	tinyDB, err := server.NewTinyDBWithMetadata(path.Join(".", "sort_plan_test"))
	if err != nil {
		t.Fatalf("failed to create tinyDB: %v", err)
	}

	err = testlib.InsertSmallTestData(t, tinyDB)
	if err != nil {
		t.Fatalf("failed to setup test data: %v", err)
	}

	tx, err := tinyDB.NewTx()
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}

	p, err := plan.NewTablePlan(tx, "student", tinyDB.MetadataMgmt())
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	groupByPlan, err := plan.NewGroupByPlan(
		tx,
		p,
		[]string{"majorid"},
		[]query.AggregateFn{
			query.NewMaxFn("gradyear"),
			query.NewMinFn("gradyear"),
			query.NewCountFn(""),
		},
	)
	if err != nil {
		t.Fatalf("failed to create GroupByPlan: %v", err)
	}

	groupByScan, err := groupByPlan.Open()
	if err != nil {
		t.Fatalf("failed to open GroupByPlan: %v", err)
	}

	err = groupByScan.BeforeFirst()
	if err != nil {
		t.Fatalf("failed to call BeforeFirst: %v", err)
	}

	expects := map[int32]struct {
		minGradYear int32
		maxGradYear int32
		count       int32
	}{
		10: {2021, 2022, 4},
		20: {2019, 2022, 4},
		30: {2020, 2021, 2},
	}

	for {
		next := groupByScan.Next()
		if !next {
			break
		}

		majorId, err := groupByScan.GetInt("majorid")
		if err != nil {
			t.Fatalf("failed to get majorid: %v", err)
		}

		minGradYear, err := groupByScan.GetInt("min(gradyear)")
		if err != nil {
			t.Fatalf("failed to get min gradyear: %v", err)
		}

		maxGradYear, err := groupByScan.GetInt("max(gradyear)")
		if err != nil {
			t.Fatalf("failed to get max gradyear: %v", err)
		}

		count, err := groupByScan.GetInt("count()")
		if err != nil {
			t.Fatalf("failed to get count: %v", err)
		}

		t.Logf("majorid: %d, minGradYear: %d, maxGradYear: %d, count: %d", majorId, minGradYear, maxGradYear, count)

		expect, ok := expects[majorId]
		if !ok {
			t.Errorf("unexpected majorid: %d", majorId)
		}

		if minGradYear != expect.minGradYear {
			t.Errorf("unexpected minGradYear of majorId %d: %d, expect: %d", majorId, minGradYear, expect.minGradYear)
		}

		if maxGradYear != expect.maxGradYear {
			t.Errorf("unexpected maxGradYear of majorId %d: %d, expect: %d", majorId, maxGradYear, expect.maxGradYear)
		}

		if count != expect.count {
			t.Errorf("unexpected count of majorid %d: %d, expect: %d", majorId, count, expect.count)
		}
	}

	groupByScan.Close()
	tx.Commit()
}
