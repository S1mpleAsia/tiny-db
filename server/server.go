package server

import (
	"fmt"

	"s1mpleasia.com/tinydb/buffer"
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/log"
)

const logFile = "./tinydb.log"

type TinyDB struct {
	fileMgmt 	*file.FileMgmt
	logMgmt		*log.LogMgmt
	bufferMgmt	*buffer.BufferMgmt
}

func NewTinyDB(dbDir string, blockSize int, bufferSize int) (*TinyDB, error) {
	fileManagement, err := file.NewFileMgmt(dbDir, int64(blockSize))

	if err != nil {
		return nil, fmt.Errorf("file.NewFileMgmt: %w", err)
	}

	logManagement, err := log.NewLogMgmt(fileManagement, logFile)

	if err != nil {
		return nil, fmt.Errorf("log.NewLogMgmt: %w", err)
	}

	bufferManagement := buffer.NewBufferMgmt(fileManagement, logManagement, bufferSize)

	return &TinyDB{
		fileMgmt: fileManagement,
		logMgmt: logManagement,
		bufferMgmt: bufferManagement,
	}, nil
}

func (db *TinyDB) FileMgmt() *file.FileMgmt {
	return db.fileMgmt
}

func (db *TinyDB) LogMgmt() *log.LogMgmt {
	return db.logMgmt
}

func (db *TinyDB) BufferMgmt() *buffer.BufferMgmt {
	return db.bufferMgmt
}
