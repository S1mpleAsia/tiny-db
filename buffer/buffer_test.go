package buffer_test

import (
	"fmt"
	"path"
	"testing"

	"s1mpleasia.com/tinydb/buffer"
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/server"
)

func TestBuffer(t *testing.T) {
	t.Parallel()
	
	tempDir := "."
	db, err := server.NewTinyDB(path.Join(tempDir, "buffertest"), 400, 3)

	if err != nil {
		t.Fatalf("NewTinyDB: %v", err)
	}

	bm := db.BufferMgmt()

	buff1, err := bm.Pin(file.NewBlockId("testfile", 1))

	if err != nil {
		t.Fatalf("bm.Pin(1): %v", err)
	}

	p := buff1.Contents()
	n := p.GetInt(80)
	p.SetInt(80, n+1)
	buff1.SetModified(1, 0) // Set placeholder value for txNum and lsn

	fmt.Printf("The new value is %d\n", (n+1))
	bm.Unpin(buff1)

	// Buff 2
	buff2, err := bm.Pin(file.NewBlockId("testfile", 2))

	if err != nil {
		t.Fatalf("bm.Pin(2): %v", err)
	}

	// Buff 3
	_, err = bm.Pin(file.NewBlockId("testfile", 3))

	if err != nil {
		t.Fatalf("bm.Pin(3): %v", err)
	}

	// Buff 4
	_, err = bm.Pin(file.NewBlockId("testfile", 4))

	if err != nil {
		t.Fatalf("bm.Pin(4): %v", err)
	}

	bm.Unpin(buff2)

	buff2, err = bm.Pin(file.NewBlockId("testfile", 1))
	if err != nil {
		t.Fatalf("bm.Pin(1): %v", err)
	}

	p2 := buff2.Contents()
	p2.SetInt(80, 9999)
	buff2.SetModified(1, 0)
	bm.Unpin(buff2)
}

func TestBufferMgmt(t *testing.T) {
	t.Parallel()

	tempDir := "."
	db, err := server.NewTinyDB(path.Join(tempDir, "buffertest"), 400, 3)
	if err != nil {
		t.Fatalf("NewTinyDB: %v", err)
	}

	bm := db.BufferMgmt()

	buffers := make([]*buffer.Buffer, 6)

	block0 := file.NewBlockId("testfile", 0)
	block1 := file.NewBlockId("testfile", 1)
	block2 := file.NewBlockId("testfile", 2)
	block3 := file.NewBlockId("testfile", 3)


	// Buff 0
	buffers[0], err = bm.Pin(block0)

	if err != nil {
		t.Fatalf("bm.Pin(0): %v", err)
	}

	// Buff 1
	buffers[1], err = bm.Pin(block1)

	if err != nil {
		t.Fatalf("bm.Pin(1): %v", err)
	}

	// Buff 2
	buffers[2], err = bm.Pin(block2)

	if err != nil {
		t.Fatalf("bm.Pin(2): %v", err)
	}

	bm.Unpin(buffers[1])
	buffers[1] = nil

	// Buff 3
	buffers[3], err = bm.Pin(block0)

	if err != nil {
		t.Fatalf("bm.Pin(3): %v", err)
	}

	// Buff 4
	buffers[4], err = bm.Pin(block1)

	if err != nil {
		t.Fatalf("bm.Pin(4): %v", err)
	}

	fmt.Printf("Available buffers: %d\n", bm.Available())
	
	fmt.Println("Attempting to pin block 3...")
	buffers[5], err = bm.Pin(block3)

	if err != nil {
		fmt.Println(err)
	}

	bm.Unpin(buffers[2])
	buffers[2] = nil

	buffers[5], err = bm.Pin(block3)

	if err != nil {
		fmt.Println("%w", err)
	}

	fmt.Println("Final buffer allocation")
	for i := 0; i < 5; i++ {
		if buffers[i] != nil {
			fmt.Printf("buff[%d] pinned to block %v", i, buffers[i].Block())
		}
	}

}
