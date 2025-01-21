package driver_test

import (
	"context"
	"database/sql"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDriver(t *testing.T) {
	t.Log("testing driver")

	db, err := sql.Open("tinydb", path.Join(".", "playerdb"))
	require.NoError(t, err, "failed to open db")

	defer db.Close()

	tx2 := beginTx(t, db)
	ctx := context.TODO()

	createTable(t, tx2, "create table player (player_id int, name varchar(10), birth_year int, country varchar(10), point int)", ctx)
	commit(t, tx2)

	tx3 := beginTx(t, db)
	updateTable(t, tx3, "insert into player (player_id, name, birth_year, country, point) values (1, 'Nobak', 1987, 'Serbia', 11055)", ctx)
	updateTable(t, tx3, "insert into player (player_id, name, birth_year, country, point) values (2, 'Carlos', 2003, 'Spain', 8855)", ctx)
	updateTable(t, tx3, "insert into player (player_id, name, birth_year, country, point) values (3, 'Daniil', 1996, 'Russia', 7555)", ctx)
	updateTable(t, tx3, "insert into player (player_id, name, birth_year, country, point) values (4, 'Jannik', 2001, 'Italy', 6490)", ctx)

	commit(t, tx3)

	tx4 := beginTx(t, db)
	points3 := queryPlayer(t, tx4, ctx)
	expected3 := []int{11055, 8855, 7555, 6490}

	for i, s := range points3 {
		if s != expected3[i] {
			t.Errorf("expected: %d, but got: %d", expected3[i], points3[i])
		}
	}

	updateTable(t, tx4, "update player set point = 8360 where player_id = 1", ctx)
	updateTable(t, tx4, "update player set point = 8130 where player_id = 2", ctx)
	updateTable(t, tx4, "update player set point = 6445 where player_id = 3", ctx)
	updateTable(t, tx4, "update player set point = 9890 where player_id = 4", ctx)
	commit(t, tx4)

	tx5 := beginTx(t, db)
	points4 := queryPlayer(t, tx5, ctx)
	expected4 := []int{8360, 8130, 6445, 9890}

	for i, s := range points4 {
		if s != expected4[i] {
			t.Errorf("expected: %d, but got: %d", expected4[i], points4[i])
		}
	}

	updateTable(t, tx5, "update player set point = 0", ctx)
	points5 := queryPlayer(t, tx5, ctx)
	expected5 := []int{0, 0, 0, 0}

	for i, s := range points5 {
		if s != expected5[i] {
			t.Errorf("expected: %d, but got: %d", expected5[i], points5[i])
		}
	}

	rollback(t, tx5)

	tx6 := beginTx(t, db)
	points6 := queryPlayer(t, tx6, ctx)
	expected6 := []int{8360, 8130, 6445, 9890}

	for i, s := range points6 {
		if s != expected6[i] {
			t.Errorf("expected: %d, but got: %d", expected5[i], points5[i])
		}
	}

	// rows6 := delete(t, tx6, )
}

func beginTx(t *testing.T, db *sql.DB) *sql.Tx {
	tx, err := db.Begin()
	require.NoError(t, err, "failed to begin transaction")

	return tx
}

func commit(t *testing.T, tx *sql.Tx) {
	err := tx.Commit()
	require.NoError(t, err, "failed to commit")
}

func rollback(t *testing.T, tx *sql.Tx) {
	err := tx.Rollback()
	require.NoError(t, err, "failed to rollback")
}

func createTable(t *testing.T, tx *sql.Tx, query string, ctx context.Context) {
	stmt, err := tx.Prepare(query)
	require.NoError(t, err, "failed to prepare statement")

	_, err = stmt.QueryContext(ctx)
	require.NoError(t, err, "failed to create table")

	t.Log("created table")
}

func updateTable(t *testing.T, tx *sql.Tx, query string, ctx context.Context) {
	stmt, err := tx.Prepare(query)
	require.NoError(t, err, "failed to prepare statement")

	result, err := stmt.ExecContext(ctx)
	require.NoError(t, err, "failed to create table")

	rows, err := result.RowsAffected()
	require.NoError(t, err, "failed to get rows")

	t.Logf("Updated %d rows\n", rows)
}

func queryPlayer(t *testing.T, tx *sql.Tx, ctx context.Context) []int {
	stmt, err := tx.Prepare("select player_id, name, birth_year, country, point from player")
	require.NoError(t, err, "failed to prepare statement")

	result, err := stmt.QueryContext(ctx)
	require.NoError(t, err, "failed to create table")

	points := []int{}

	for result.Next() {
		var playerID int
		var name string
		var birthYear int
		var country string
		var point int

		err = result.Scan(&playerID, &name, &birthYear, &country, &point)
		require.NoError(t, err, "failed to scan")

		t.Logf("player {player_id: %d, name: %s, birth_year: %d, country: %s, point: %d}", playerID, name, birthYear, country, point)
		points = append(points, point)
	}

	return points
}
