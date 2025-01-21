package plan_test

import (
	"fmt"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"s1mpleasia.com/tinydb/plan"
	"s1mpleasia.com/tinydb/query"
	"s1mpleasia.com/tinydb/record"
	"s1mpleasia.com/tinydb/server"
	"s1mpleasia.com/tinydb/testlib"
)

func TestSingleTablePlan(t *testing.T) {
	db, err := server.NewTinyDBWithMetadata(path.Join(".", "single_table_plan_test"))
	require.NoError(t, err, "create DB failed")

	err = testlib.InsertSmallTestData(t, db)
	require.NoError(t, err, "setup data failed")

	fmt.Println("Start test single table plan")
	tx, err := db.NewTx()
	require.NoError(t, err, "create transaction failed")

	p1, err := plan.NewTablePlan(tx, "student", db.MetadataMgmt())
	require.NoError(t, err, "create table plan failed")

	p2, err := plan.NewSelectPlan(p1, query.NewPredicateWithTerm(
		query.NewTerm(
			query.NewExpressionWithField("majorid"),
			query.NewExpressionWithConstant(record.NewConstantWithInt(10)),
		),
	))
	require.NoError(t, err, "create select plan p2 failed")

	p3, err := plan.NewSelectPlan(p2, query.NewPredicateWithTerm(
		query.NewTerm(
			query.NewExpressionWithField("gradyear"),
			query.NewExpressionWithConstant(record.NewConstantWithInt(2020)),
		),
	))
	require.NoError(t, err, "create select plan p3 failed")

	p4, err := plan.NewProjectPlan(p3, []string{"sname", "majorid", "gradyear"})
	require.NoError(t, err, "create project plan p4 failed")

	printStats(1, p1)
	printStats(2, p2)
	printStats(3, p3)
	printStats(4, p4)

	tx.Commit()
}

func TestMultipleTablePlan(t *testing.T) {
	db, err := server.NewTinyDBWithMetadata(path.Join(".", "multiple_table_plan_test"))
	require.NoError(t, err, "create DB failed")

	err = testlib.InsertSmallTestData(t, db)
	require.NoError(t, err, "setup data failed")

	tx, err := db.NewTx()
	require.NoError(t, err, "create transaction failed")

	t.Log("NewTablePlan(student)")
	p1, err := plan.NewTablePlan(tx, "student", db.MetadataMgmt())
	require.NoError(t, err, "create table plan p1 failed")

	t.Log("NewTablePlan(dept)")
	p2, err := plan.NewTablePlan(tx, "dept", db.MetadataMgmt())
	require.NoError(t, err, "create table plan p2 failed")

	t.Log("NewProductPlan(p1, p2)")
	p3, err := plan.NewProductPlan(p1, p2)
	require.NoError(t, err, "create product plan p3 failed")

	t.Log("NewSelectPlan(p3, majorid = did)")
	p4, err := plan.NewSelectPlan(p3, query.NewPredicateWithTerm(
		query.NewTerm(
			query.NewExpressionWithField("majorid"),
			query.NewExpressionWithField("did"),
		),
	))
	require.NoError(t, err, "create select plan p4 failed")

	t.Log("print")
	printStats(1, p1)
	printStats(2, p2)
	printStats(3, p3)
	printStats(4, p4)

	tx.Commit()
}

func printStats(n int, p plan.Plan) {
	planName := fmt.Sprintf("p%d", n)
	fmt.Println("Here are the stats for plan " + planName)
	fmt.Printf("\tR(%s): %d\n", planName, p.RecordsOutput())
	fmt.Printf("\tB(%s): %d\n", planName, p.BlockAccessed())
	fmt.Println()
}
