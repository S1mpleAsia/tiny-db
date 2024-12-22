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

func TestConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	db, err := server.NewTinyDB(path.Join(".", "concurrencytest"), 400, 8)
	if err != nil {
		t.Fatal(err)
	}

	fm := db.FileMgmt()
	lm := db.LogMgmt()
	bm := db.BufferMgmt()

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		tx1, err := transaction.NewTransaction(fm, lm, bm)
		if err != nil {
			panic(err)
		}

		blk1 := file.NewBlockId("testfile", 1)
		blk2 := file.NewBlockId("testfile", 2)

		err = tx1.Pin(blk1)
		if err != nil {
			panic(err)
		}

		err = tx1.Pin(blk2)
		if err != nil {
			panic(err)
		}

		fmt.Println("tx1: request slock 1")
		_, err = tx1.GetInt(blk1, 0)
		if err != nil {
			fmt.Printf("tx1: %v, rollback", err)
			tx1.Rollback()
			return
		}

		fmt.Println("tx1: receive slock 1")
		time.Sleep(1 * time.Second)

		fmt.Println("tx1: request slock 2")
		_, err = tx1.GetInt(blk2, 0)
		if err != nil {
			fmt.Printf("tx1: %v, rollback\n", err)
			tx1.Rollback()
			return
		}

		fmt.Println("tx1: receive slock 2")

		tx1.Commit()
		fmt.Println("tx1: Commit")
	}()

	go func() {
		defer wg.Done()

		tx2, err := transaction.NewTransaction(fm, lm, bm)
		if err != nil {
			panic(err)
		}

		blk1 := file.NewBlockId("testfile", 1)
		blk2 := file.NewBlockId("testfile", 2)

		err = tx2.Pin(blk1)
		if err != nil {
			panic(err)
		}

		err = tx2.Pin(blk2)
		if err != nil {
			panic(err)
		}

		fmt.Println("tx2: request xlock 2")
		err = tx2.SetInt(blk2, 0, 0, false)
		if err != nil {
			fmt.Printf("tx2: %v, rollback\n", err)
			tx2.Rollback()
			return
		}

		fmt.Println("tx2: receive xlock 2")
		time.Sleep(1 * time.Second)

		fmt.Println("tx2: request slock 1")
		_, err = tx2.GetInt(blk1, 0)
		if err != nil {
			fmt.Printf("tx2: %v, rollback\n", err)
			tx2.Rollback()
			return
		}

		fmt.Println("tx2: receive slock 1")
		tx2.Commit()
		fmt.Println("tx2: commit")
	}()

	go func() {
		defer wg.Done()

		tx3, err := transaction.NewTransaction(fm, lm, bm)
		if err != nil {
			panic(err)
		}

		blk1 := file.NewBlockId("testfile", 1)
		blk2 := file.NewBlockId("testfile", 2)

		err = tx3.Pin(blk1)
		if err != nil {
			panic(err)
		}

		err = tx3.Pin(blk2)
		if err != nil {
			panic(err)
		}

		time.Sleep(500 * time.Millisecond)
		fmt.Println("tx3: request xlock 1")
		err = tx3.SetInt(blk1, 0, 0, false)
		if err != nil {
			fmt.Printf("tx3: %v, rollback\n", err)
			tx3.Rollback()
			return
		}

		fmt.Println("tx3: receive xlock 1")
		time.Sleep(1 * time.Second)

		fmt.Println("tx3: request slock 2")
		_, err = tx3.GetInt(blk2, 0)
		if err != nil {
			fmt.Printf("tx3: %v, rollback\n", err)
			tx3.Rollback()
			return
		}

		fmt.Println("tx3: receive slock 2")
		tx3.Commit()
		fmt.Println("tx3: commit")
	}()

	wg.Wait()
}
