package log_test

import (
	"fmt"
	"path"
	"strconv"
	"testing"

	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/log"
	"s1mpleasia.com/tinydb/server"
)

func TestLog(t *testing.T) {
	t.Parallel()
	dir := "."

	db, err := server.NewTinyDB(path.Join(dir, "filetest"), 400, 3)

	if err != nil {
		t.Fatalf("NewTinyDB: %v", err)
	}

	lm := db.LogMgmt()
	fmt.Printf("Initial empty log file: ")
	output := printLogRecords(lm)

	if output != genWant(0) {
		t.Fatalf("Got %v, want %v", output, genWant(0))
	}
	
	fmt.Println("Done")

	createRecords(t, lm, 1, 35)
	fmt.Println("The log file now has these records:")
	
	output = printLogRecords(lm)
	if output != genWant(35) {
		t.Fatalf("Got %v, want %v", output, genWant(35))
	}

	createRecords(t, lm, 36, 70)
	lm.Flush(65)

	fmt.Println("The log file now has these records:")
	output = printLogRecords(lm)
	if output != genWant(70) {
		t.Fatalf("Got %v, want %v", output, genWant(70))
	}
}

func printLogRecords(logMgmt *log.LogMgmt) string {
	iter := logMgmt.Iterator()

	res := ""
	sentinel := 0


	for iter.HasNext() {
		rec := iter.Next()
		p := file.NewPageWith(rec)
		s := p.GetString(0)
		npos := file.MaxLength(len(s))
		val := p.GetInt(npos)
		output := fmt.Sprintf("[%s %d]\n", s, val)
		fmt.Print(output)
		res += output

		sentinel++

		if sentinel > 100 {
			panic("Too many records")
		}
	}

	return res
}


func createRecords(t *testing.T, lm *log.LogMgmt, start, end int) {
	fmt.Println("Creating records:")

	for i := start; i <= end; i++ {
		rec := createLogRecord("record" + strconv.Itoa(i), i+100)
		lsn, err := lm.Append(rec)

		if err != nil {
			t.Fatalf("Append: %v", err)
		}

		fmt.Println(strconv.Itoa(lsn))
	}

	fmt.Println("")
}

// Create a log record having 2 value: a string and an integer
func createLogRecord(s string, n int) []byte {
	spos := 0
	npos := spos + file.MaxLength(len(s))
	b := make([]byte, npos + file.INT_32_BITS)
	p := file.NewPageWith(b)
	p.SetString(spos, s)
	p.SetInt(npos, int32(n))
	
	return b
}

func genWant(n int) string {
	want := ""

	for i := n; i > 0; i-- {
		want += fmt.Sprintf("[%s %d]\n", "record" + strconv.Itoa(i), i+100)
	}

	return want
}