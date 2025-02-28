package file

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"strings"
	"unicode/utf16"
)

const (
	INT_32_BITS = 4
	UTF_16_SIZE = 2
)

/*
- A BlockId object identifies a specified block by its name and logical block number
- BlockNum is 0-indexed
*/
type BlockId struct {
	fileName string
	blockNum int64
}

func NewBlockId(fileName string, blockNum int64) *BlockId {
	return &BlockId{
		fileName: fileName,
		blockNum: blockNum,
	}
}

func (block *BlockId) FileName() string {
	return block.fileName
}

func (block *BlockId) BlockNumber() int64 {
	return block.blockNum
}

func (b *BlockId) Equals(other *BlockId) bool {
	return b.fileName == other.fileName && b.blockNum == other.blockNum
}

/*
A Page object holds the content of a disk block.
A page can hold int, string or blobs value type (etc. arbitrary arrays of bytes). It's corresponding for a page in RAM memory
*/
type Page struct {
	buffer []byte
}

func NewPage(blockSize int64) *Page {
	return &Page{
		buffer: make([]byte, blockSize),
	}
}

func NewPageWith(buffer []byte) *Page {
	return &Page{
		buffer: buffer,
	}
}

func (p *Page) GetInt(offset int) int32 {
	return int32(binary.LittleEndian.Uint32(p.buffer[offset : offset+INT_32_BITS]))
}

func (p *Page) SetInt(offset int, val int32) {
	binary.LittleEndian.PutUint32(p.buffer[offset:offset+INT_32_BITS], uint32(val))
}

func (p *Page) GetBytes(offset int) []byte {
	length := p.GetInt(offset)
	return p.buffer[offset+INT_32_BITS : offset+INT_32_BITS+int(length)]
}

/*
	Save a blob as two values:

-	First the number of bytes in the specified blob
-	Second is the bytes themselves

+------------+-----------+
|  	  4B	 |	  ...	 |
+------------+-----------+

	len(val)		 val
*/
func (p *Page) SetBytes(offset int, val []byte) {
	p.SetInt(offset, int32(len(val)))
	copy(p.buffer[offset+INT_32_BITS:], val)
}

/*
Save a string as two values:
-	First 4 bytes is store the size of the string
-	Second is the string themselves using UTF-16 (16 bit per character)

+------------+-----------+
|  	  4B	 |	  ...	 |
+------------+-----------+

	len(val)		 val
*/
func (p *Page) GetString(offset int) string {
	length := int(p.GetInt(offset)) / UTF_16_SIZE

	runes := make([]uint16, length)
	for i := range length {
		runes[i] = p.getUint16(offset + INT_32_BITS + i*UTF_16_SIZE)
	}

	return string(utf16.Decode(runes))
}

func (p *Page) SetString(offset int, val string) {
	runes := utf16.Encode([]rune(val))
	p.SetInt(offset, int32(len(runes)*UTF_16_SIZE))

	for i, r := range runes {
		p.setUint16(offset+INT_32_BITS+i*UTF_16_SIZE, r)
	}

}

func (p *Page) getUint16(offset int) uint16 {
	return binary.LittleEndian.Uint16(p.buffer[offset : offset+UTF_16_SIZE])
}

func (p *Page) setUint16(offset int, val uint16) {
	binary.LittleEndian.PutUint16(p.buffer[offset:offset+UTF_16_SIZE], val)
}

func MaxLength(length int) int {
	return INT_32_BITS + length*UTF_16_SIZE
}

/*
FileMgmt responsible for implement methods that read and write page to disk blocks
*/
type FileMgmt struct {
	dbDir     string
	blockSize int64
	isNew     bool
	files     map[string]*os.File
}

func NewFileMgmt(dbDir string, blockSize int64) (*FileMgmt, error) {
	isNew := false
	// If not exists, create dbDir recursively
	if _, err := os.Stat(dbDir); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("os Stat: %w", err)
		}
		isNew = true

		err := os.MkdirAll(dbDir, 0o700)
		if err != nil {
			return nil, fmt.Errorf("os.MkdirAll: %w", err)
		}
	}

	files, err := os.ReadDir(dbDir)

	if err != nil {
		return nil, fmt.Errorf("os.ReadDir: %w", err)
	}

	// Remove temporary table
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "temp") {
			continue
		}

		if err = os.Remove(file.Name()); err != nil {
			return nil, fmt.Errorf("os.Remove: %w", err)
		}
	}

	return &FileMgmt{
		dbDir:     dbDir,
		blockSize: blockSize,
		isNew:     isNew,
		files:     make(map[string]*os.File),
	}, nil
}

func (fm *FileMgmt) IsNew() bool {
	return fm.isNew
}

func (fm *FileMgmt) BlockSize() int64 {
	return fm.blockSize
}

// Read from disk block to page in memory
func (fm *FileMgmt) Read(block *BlockId, p *Page) error {
	f, err := fm.openFile(block.fileName)

	if err != nil {
		return fmt.Errorf("fm.openFile: %w", err)
	}

	_, err = f.Seek(block.blockNum*fm.blockSize, 0)
	if err != nil {
		return fmt.Errorf("f.Seek: %w", err)
	}

	_, err = f.Read(p.buffer)
	if err != nil {
		return fmt.Errorf("f.Read: %w", err)
	}

	return nil
}

// Write value from page on RAM -> block on disk
func (fm *FileMgmt) Write(block *BlockId, p *Page) error {
	f, err := fm.openFile(block.fileName)

	if err != nil {
		return fmt.Errorf("fm.openFile: %w", err)
	}

	_, err = f.Seek(block.blockNum*fm.blockSize, 0)
	if err != nil {
		return fmt.Errorf("f.Seek: %w", err)
	}

	_, err = f.Write(p.buffer)
	if err != nil {
		return fmt.Errorf("f.Write: %w", err)
	}

	return nil
}

func (fm *FileMgmt) Append(fileName string) (*BlockId, error) {
	newBlockNum, err := fm.Length(fileName)

	if err != nil {
		return nil, fmt.Errorf("fm.Length: %w", err)
	}

	block := NewBlockId(fileName, newBlockNum)
	b := make([]byte, fm.blockSize)

	f, err := fm.openFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("fm.openFile: %w", err)
	}

	f.Seek(block.blockNum*fm.blockSize, 0)
	f.Write(b)

	return block, nil
}

/* Length in number Blocks */
func (fm *FileMgmt) Length(fileName string) (int64, error) {
	f, err := fm.openFile(fileName)

	if err != nil {
		return 0, fmt.Errorf("fm.openFile: %w", err)
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("f.Stat: %w", err)
	}

	return fileInfo.Size() / fm.blockSize, nil
}

func (fm *FileMgmt) openFile(fileName string) (*os.File, error) {
	if f, ok := fm.files[fileName]; ok {
		return f, nil
	}

	f, err := os.OpenFile(path.Join(fm.dbDir, fileName), os.O_RDWR|os.O_CREATE, 0o600)

	if err != nil {
		return nil, fmt.Errorf("os.OpenFile: %w", err)
	}

	fm.files[fileName] = f

	return f, nil
}
