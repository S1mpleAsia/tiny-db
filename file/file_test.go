package file_test

import (
	"path"
	"testing"

	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/server"
)

func TestFile(t *testing.T) {
	t.Parallel()

	dir := "./test"

	db, err := server.NewTinyDB(path.Join(dir, "filetest"), 400)
	if err != nil {
		t.Fatalf("NewTinyDB: %v", err)
	}

	fm := db.FileMgmt()

	p1 := file.NewPage(fm.BlockSize()) // 400 bytes
	pos1 := 88
	strVal := "abcdefghijklm"
	p1.SetString(pos1, strVal)

	size := file.MaxLength(len(strVal))
	pos2 := pos1 + size
	intVar := int32(345)
	p1.SetInt(pos2, intVar)

	blk := file.NewBlockId("testfile", 2)
	err = fm.Write(blk, p1)
	if err != nil {
		t.Fatalf("fm.Write: %v", err)
	}

	p2 := file.NewPage(fm.BlockSize())
	err = fm.Read(blk, p2)

	if err != nil {
		t.Fatalf("fm.Read: %v", err)
	}

	if p2.GetInt(pos2) != intVar {
		t.Errorf("expected %d, got %d", intVar, p2.GetInt(pos2))
	}

	if p2.GetString(pos1) != strVal {
		t.Errorf("expected %q, got %q", strVal, p2.GetString(pos1))
	}

}