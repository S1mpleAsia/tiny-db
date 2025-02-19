package index_test

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

func TestIndexRetrieval(t *testing.T) {
	t.Log("testing index retrieval")

	db, err := server.NewTinyDBWithMetadata(path.Join(".", "index_retrieval_test"))
	require.NoError(t, err, "create DB failed")

	err = testlib.InsertSmallTestData(t, db)
	require.NoError(t, err, "setup data failed")

	fmt.Println("Start test index retrieval")

	tx, err := db.NewTx()
	require.NoError(t, err, "create transaction failed")

	mdm := db.MetadataMgmt()

	studentPlan, err := plan.NewTablePlan(tx, "student", mdm)
	require.NoError(t, err, "failed to open table scan")

	scan, err := studentPlan.Open()
	require.NoError(t, err, "failed to open table scan")

	studentScan, ok := scan.(*query.TableScan)
	if !ok {
		t.Fatalf("scan is not a table scan")
	}

	indexInfoMap, err := mdm.GetIndexInfo("student", tx)
	require.NoError(t, err, "failed to get index info")

	majorIdIndexInfo, ok := indexInfoMap["majorId"]
	if !ok {
		t.Fatalf("no index on majorId")
	}

	indexInfo := majorIdIndexInfo.Open()

	// Retrieve all index records having a dataval of 20
	majorIdIndexVal := record.NewConstantWithInt(20)

	err = indexInfo.BeforeFirst(majorIdIndexVal)
	require.NoError(t, err, "falied to call BeforeFirst")

	for {
		ok, err = indexInfo.Next()
		require.NoError(t, err, "failed to call Next")

		if !ok {
			break
		}

		rid, err := indexInfo.GetDataRID()
		require.NoError(t, err, "failed to get data rid")

		studentScan.MoveToRID(rid)

		studentName, err := studentScan.GetString("SName")
		require.NoError(t, err, "failed to get student name")

		fmt.Printf("student name: %s\n", studentName)
	}

	indexInfo.Close()
	studentScan.Close()
	tx.Commit()
}
