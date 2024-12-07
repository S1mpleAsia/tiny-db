package server

import (
	"fmt"

	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/log"
)

const logFile = "./tinydb.log"

type TinyDB struct {
	fileMgmt 	*file.FileMgmt
	logMgmt		*log.LogMgmt
}

func NewTinyDB(dbDir string, blockSize int) (*TinyDB, error) {
	fileManagement, err := file.NewFileMgmt(dbDir, int64(blockSize))

	if err != nil {
		return nil, fmt.Errorf("file.NewFileMgmt: %w", err)
	}

	logManagement, err := log.NewLogMgmt(fileManagement, logFile)

	if err != nil {
		return nil, fmt.Errorf("log.NewLogMgmt: %w", err)
	}

	return &TinyDB{
		fileMgmt: fileManagement,
		logMgmt: logManagement,
	}, nil
}

func (db *TinyDB) FileMgmt() *file.FileMgmt {
	return db.fileMgmt
}

func (db *TinyDB) LogMgmt() *log.LogMgmt {
	return db.logMgmt
}
