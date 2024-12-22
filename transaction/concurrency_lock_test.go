package transaction_test

import (
	"fmt"
	"path"
	"sync"
	"testing"
	"time"

	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/server"
	"s1mpleasia.com/tinydb/transaction"
)

func TestConcurrencySLockTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode, this takes long time for checking timeout.")
	}

	db, err := server.NewTinyDB(path.Join(".", "concurrencytest"), 400, 8)
	if err != nil {
		t.Fatal(err)
	}

	fm := db.FileMgmt()
	lm := db.LogMgmt()
	bm := db.BufferMgmt()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		tx1, err := transaction.NewTransaction(fm, lm, bm)
		if err != nil {
			panic(err)
		}

		blk1 := file.NewBlockId("testfile", 1)
		err = tx1.Pin(blk1)
		if err != nil {
			panic(err)
		}

		fmt.Println("Tx1: Request slock 1")
		_, err = tx1.GetInt(blk1, 0)

		if err != nil {
			fmt.Printf("Tx1: %v, rollback\n", err)
			tx1.Rollback()
			return
		}
		fmt.Println("Tx1: Receive slock 1")

		time.Sleep(15 * time.Second)
		tx1.Commit()
		fmt.Println("Tx1: Commit")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(3 * time.Second)

		tx2, err := transaction.NewTransaction(fm, lm, bm)
		if err != nil {
			panic(err)
		}

		blk1 := file.NewBlockId("testfile", 1)
		err = tx2.Pin(blk1)
		if err != nil {
			panic(err)
		}

		fmt.Println("Tx1: Request xlock 1")
		err = tx2.SetInt(blk1, 0, 0, false)
		if err != nil {
			fmt.Printf("tx2 %v, rollback\n", err)
			tx2.Rollback()
			return
		}

		fmt.Println("tx2 does not reach here")
	}()

	wg.Wait()
}

func TestConcurrencyXLockTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode, this takes long time for checking timeout.")
	}

	db, err := server.NewTinyDB(path.Join(".", "concurrencytest"), 400, 8)
	if err != nil {
		t.Fatal(err)
	}

	fm := db.FileMgmt()
	lm := db.LogMgmt()
	bm := db.BufferMgmt()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		tx1, err := transaction.NewTransaction(fm, lm, bm)
		if err != nil {
			panic(err)
		}

		blk := file.NewBlockId("testfile", 1)
		err = tx1.Pin(blk)
		if err != nil {
			panic(err)
		}

		fmt.Println("tx1: request xlock 1")
		err = tx1.SetInt(blk, 0, 0, false)
		if err != nil {
			fmt.Printf("tx1: %v, rollback", err)
			tx1.Rollback()
			return
		}

		fmt.Println("tx1: receive xlock 1")
		time.Sleep(5 * time.Second)
		tx1.Commit()
		fmt.Println("tx1: commit")
	}()

	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)

		tx2, err := transaction.NewTransaction(fm, lm, bm)
		if err != nil {
			panic(err)
		}

		blk := file.NewBlockId("testfile", 1)
		err = tx2.Pin(blk)
		if err != nil {
			panic(err)
		}

		fmt.Println("tx2: request slock 1")
		_, err = tx2.GetInt(blk, 0)
		if err != nil {
			fmt.Printf("tx1: %v, rollback", err)
			tx2.Rollback()
			return
		}

		fmt.Println("tx2: receive slock 1")
		tx2.Commit()
		fmt.Println("tx2: commit")
	}()

	wg.Wait()
}
