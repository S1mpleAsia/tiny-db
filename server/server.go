package server

import (
	"fmt"

	"s1mpleasia.com/tinydb/buffer"
	"s1mpleasia.com/tinydb/file"
	"s1mpleasia.com/tinydb/log"
	"s1mpleasia.com/tinydb/metadata"
	"s1mpleasia.com/tinydb/transaction"
)

const BLOCK_SIZE = 400
const BUFFER_SIZE = 8
const logFile = "./tinydb.log"

type TinyDB struct {
	fileMgmt     *file.FileMgmt
	logMgmt      *log.LogMgmt
	bufferMgmt   *buffer.BufferMgmt
	metadataMgmt *metadata.MetadataMgmt
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
		fileMgmt:   fileManagement,
		logMgmt:    logManagement,
		bufferMgmt: bufferManagement,
	}, nil
}

func NewTinyDBWithMetadata(dirName string) (*TinyDB, error) {
	return newTinyDBWithMetadata(dirName, BUFFER_SIZE)
}

func newTinyDBWithMetadata(dirName string, bufferSize int) (*TinyDB, error) {
	db, err := NewTinyDB(dirName, BLOCK_SIZE, bufferSize)
	if err != nil {
		return nil, fmt.Errorf("SimpleDB: %w", err)
	}

	tx, err := db.NewTx()
	if err != nil {
		return nil, fmt.Errorf("db.NewTx: %w", err)
	}

	isNew := db.fileMgmt.IsNew()

	if isNew {
		fmt.Printf("creating new database: %q\n", dirName)
	} else {
		fmt.Printf("recovering existing database: %q\n", dirName)
		tx.Recover()
	}

	metadataMgmt, err := metadata.NewMetadataMgmt(isNew, tx)
	if err != nil {
		return nil, fmt.Errorf("metadata.NewManager: %w", err)
	}
	db.metadataMgmt = metadataMgmt

	tx.Commit()
	return db, nil
}

func (db *TinyDB) NewTx() (*transaction.Transaction, error) {
	return transaction.NewTransaction(db.fileMgmt, db.logMgmt, db.bufferMgmt)
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

func (db *TinyDB) MetadataMgmt() *metadata.MetadataMgmt {
	return db.metadataMgmt
}
