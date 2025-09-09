package storage

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"sync"
	"time"
)

// FileStorageEngine implements the StorageEngine interface
type FileStorageEngine struct {
	filepath   string
	file       *os.File
	header     *FileHeader
	cache      PageCache
	wal        WAL
	mu         sync.RWMutex
	nextPageID uint32
	freeList   []uint32
	txCounter  uint64
}

// NewFileStorageEngine creates a new file storage engine
func NewFileStorageEngine(cacheSize int) *FileStorageEngine {
	return &FileStorageEngine{
		cache:     NewLRUCache(cacheSize),
		wal:       NewWAL(),
		freeList:  make([]uint32, 0),
		nextPageID: 1, // Page 0 is reserved for header
	}
}

// Open opens or creates a database file
func (e *FileStorageEngine) Open(filepath string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.filepath = filepath

	// Try to open existing file
	file, err := os.OpenFile(filepath, os.O_RDWR, 0644)
	if err != nil {
		// Create new file if it doesn't exist
		if os.IsNotExist(err) {
			return e.createNewFile(filepath)
		}
		return fmt.Errorf("failed to open file: %w", err)
	}

	e.file = file

	// Read and validate header
	if err := e.readHeader(); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	// Open WAL
	walPath := filepath + ".wal"
	if err := e.wal.Open(walPath); err != nil {
		return fmt.Errorf("failed to open WAL: %w", err)
	}

	// Replay WAL if needed
	if err := e.replayWAL(); err != nil {
		return fmt.Errorf("failed to replay WAL: %w", err)
	}

	return nil
}

// Close closes the storage engine
func (e *FileStorageEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Sync all changes
	if err := e.sync(); err != nil {
		return err
	}

	// Close WAL
	if e.wal != nil {
		if err := e.wal.Close(); err != nil {
			return err
		}
	}

	// Close file
	if e.file != nil {
		if err := e.file.Close(); err != nil {
			return err
		}
	}

	// Clear cache
	if e.cache != nil {
		e.cache.Clear()
	}

	return nil
}

// Sync flushes all pending changes to disk
func (e *FileStorageEngine) Sync() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.sync()
}

func (e *FileStorageEngine) sync() error {
	if e.file == nil {
		return fmt.Errorf("file not open")
	}

	// Update header
	e.header.Modified = time.Now().Unix()
	if err := e.writeHeader(); err != nil {
		return err
	}

	// Sync file
	return e.file.Sync()
}

// ReadPage reads a page from storage
func (e *FileStorageEngine) ReadPage(pageID uint32) (*Page, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Check cache first
	if page, found := e.cache.Get(pageID); found {
		return e.copyPage(page), nil
	}

	// Read from disk
	page, err := e.readPageFromDisk(pageID)
	if err != nil {
		return nil, err
	}

	// Add to cache
	e.cache.Put(pageID, page)

	return e.copyPage(page), nil
}

// WritePage writes a page to storage
func (e *FileStorageEngine) WritePage(page *Page) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Validate page
	if err := e.validatePage(page); err != nil {
		return err
	}

	// Calculate checksum
	page.Checksum = e.calculatePageChecksum(page)

	// Write to WAL first
	walEntry := &WALEntry{
		Sequence:  e.getNextWALSequence(),
		Type:      WALOpUpdate,
		PageID:    page.ID,
		Data:      e.serializePage(page),
		Timestamp: time.Now().Unix(),
	}
	walEntry.Checksum = e.calculateWALChecksum(walEntry)

	if err := e.wal.Append(walEntry); err != nil {
		return fmt.Errorf("failed to write WAL entry: %w", err)
	}

	// Write to disk
	if err := e.writePageToDisk(page); err != nil {
		return err
	}

	// Update cache
	e.cache.Put(page.ID, e.copyPage(page))

	return nil
}

// AllocatePage allocates a new page
func (e *FileStorageEngine) AllocatePage() (uint32, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var pageID uint32

	// Try to reuse a free page
	if len(e.freeList) > 0 {
		pageID = e.freeList[len(e.freeList)-1]
		e.freeList = e.freeList[:len(e.freeList)-1]
	} else {
		// Allocate new page
		pageID = e.nextPageID
		e.nextPageID++
		e.header.PageCount++
	}

	// Create empty page
	page := &Page{
		ID:   pageID,
		Type: PageTypeVectorLeaf,
		Size: 0,
		Data: make([]byte, PageSize-32), // Reserve space for header
	}

	// Write page
	if err := e.WritePage(page); err != nil {
		return 0, err
	}

	return pageID, nil
}

// FreePage marks a page as free
func (e *FileStorageEngine) FreePage(pageID uint32) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Add to free list
	e.freeList = append(e.freeList, pageID)

	// Remove from cache
	e.cache.Remove(pageID)

	// Write WAL entry
	walEntry := &WALEntry{
		Sequence:  e.getNextWALSequence(),
		Type:      WALOpDelete,
		PageID:    pageID,
		Timestamp: time.Now().Unix(),
	}
	walEntry.Checksum = e.calculateWALChecksum(walEntry)

	return e.wal.Append(walEntry)
}

// BeginTx begins a new transaction
func (e *FileStorageEngine) BeginTx() (Transaction, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.txCounter++
	return &FileTransaction{
		engine: e,
		id:     e.txCounter,
		pages:  make(map[uint32]*Page),
	}, nil
}

// Compact performs database compaction
func (e *FileStorageEngine) Compact() error {
	// TODO: Implement compaction
	return nil
}

// Stats returns storage statistics
func (e *FileStorageEngine) Stats() *StorageStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cacheStats := e.cache.Stats()

	return &StorageStats{
		TotalPages:   uint64(e.header.PageCount),
		UsedPages:    uint64(e.header.PageCount) - uint64(len(e.freeList)),
		FreePages:    uint64(len(e.freeList)),
		PageSize:     PageSize,
		FileSize:     int64(e.header.PageCount) * PageSize,
		CacheHitRate: cacheStats.HitRate,
	}
}

// Private methods

func (e *FileStorageEngine) createNewFile(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	e.file = file

	// Create header
	e.header = &FileHeader{
		Magic:     [8]byte{'V', 'I', 'T', 'T', 'O', 'R', 'I', 'A'},
		Version:   1,
		PageSize:  PageSize,
		PageCount: 1, // Header page
		RootPage:  0,
		FreeList:  0,
		Created:   time.Now().Unix(),
		Modified:  time.Now().Unix(),
	}

	// Write header
	if err := e.writeHeader(); err != nil {
		return err
	}

	// Open WAL
	walPath := filepath + ".wal"
	return e.wal.Open(walPath)
}

func (e *FileStorageEngine) readHeader() error {
	// Read header from page 0
	headerData := make([]byte, PageSize)
	if _, err := e.file.ReadAt(headerData, 0); err != nil {
		return err
	}

	// Parse header
	buf := bytes.NewReader(headerData)
	e.header = &FileHeader{}

	if err := binary.Read(buf, binary.LittleEndian, e.header); err != nil {
		return err
	}

	// Validate magic
	expectedMagic := [8]byte{'V', 'I', 'T', 'T', 'O', 'R', 'I', 'A'}
	if e.header.Magic != expectedMagic {
		return fmt.Errorf("invalid file format")
	}

	// Update next page ID
	e.nextPageID = uint32(e.header.PageCount)

	return nil
}

func (e *FileStorageEngine) writeHeader() error {
	// Calculate checksum
	e.header.Checksum = e.calculateHeaderChecksum()

	// Serialize header
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, e.header); err != nil {
		return err
	}

	// Pad to page size
	headerData := make([]byte, PageSize)
	copy(headerData, buf.Bytes())

	// Write to disk
	_, err := e.file.WriteAt(headerData, 0)
	return err
}

func (e *FileStorageEngine) readPageFromDisk(pageID uint32) (*Page, error) {
	offset := int64(pageID) * PageSize
	pageData := make([]byte, PageSize)

	if _, err := e.file.ReadAt(pageData, offset); err != nil {
		return nil, err
	}

	return e.deserializePage(pageData), nil
}

func (e *FileStorageEngine) writePageToDisk(page *Page) error {
	offset := int64(page.ID) * PageSize
	pageData := e.serializePage(page)

	// Pad to page size
	if len(pageData) < PageSize {
		padded := make([]byte, PageSize)
		copy(padded, pageData)
		pageData = padded
	}

	_, err := e.file.WriteAt(pageData, offset)
	return err
}

func (e *FileStorageEngine) serializePage(page *Page) []byte {
	buf := new(bytes.Buffer)

	// Write page header
	binary.Write(buf, binary.LittleEndian, page.ID)
	binary.Write(buf, binary.LittleEndian, page.Type)
	binary.Write(buf, binary.LittleEndian, page.Size)
	binary.Write(buf, binary.LittleEndian, page.Flags)
	binary.Write(buf, binary.LittleEndian, page.LSN)
	binary.Write(buf, binary.LittleEndian, page.Checksum)

	// Write data
	buf.Write(page.Data)

	return buf.Bytes()
}

func (e *FileStorageEngine) deserializePage(data []byte) *Page {
	buf := bytes.NewReader(data)

	page := &Page{}
	binary.Read(buf, binary.LittleEndian, &page.ID)
	binary.Read(buf, binary.LittleEndian, &page.Type)
	binary.Read(buf, binary.LittleEndian, &page.Size)
	binary.Read(buf, binary.LittleEndian, &page.Flags)
	binary.Read(buf, binary.LittleEndian, &page.LSN)
	binary.Read(buf, binary.LittleEndian, &page.Checksum)

	// Read remaining data
	remaining := make([]byte, len(data)-24) // 24 bytes for header
	buf.Read(remaining)
	page.Data = remaining

	return page
}

func (e *FileStorageEngine) validatePage(page *Page) error {
	if page == nil {
		return fmt.Errorf("page is nil")
	}

	if len(page.Data) > PageSize-32 {
		return fmt.Errorf("page data too large")
	}

	return nil
}

func (e *FileStorageEngine) calculatePageChecksum(page *Page) uint32 {
	data := e.serializePage(&Page{
		ID:    page.ID,
		Type:  page.Type,
		Size:  page.Size,
		Flags: page.Flags,
		LSN:   page.LSN,
		Data:  page.Data,
	})
	return crc32.ChecksumIEEE(data)
}

func (e *FileStorageEngine) calculateHeaderChecksum() uint32 {
	buf := new(bytes.Buffer)
	header := *e.header
	header.Checksum = 0 // Exclude checksum from calculation

	binary.Write(buf, binary.LittleEndian, &header)
	return crc32.ChecksumIEEE(buf.Bytes())
}

func (e *FileStorageEngine) calculateWALChecksum(entry *WALEntry) uint32 {
	buf := new(bytes.Buffer)
	walEntry := *entry
	walEntry.Checksum = 0 // Exclude checksum from calculation

	binary.Write(buf, binary.LittleEndian, &walEntry)
	return crc32.ChecksumIEEE(buf.Bytes())
}

func (e *FileStorageEngine) copyPage(page *Page) *Page {
	data := make([]byte, len(page.Data))
	copy(data, page.Data)

	return &Page{
		ID:       page.ID,
		Type:     page.Type,
		Size:     page.Size,
		Flags:    page.Flags,
		LSN:      page.LSN,
		Checksum: page.Checksum,
		Data:     data,
	}
}

func (e *FileStorageEngine) getNextWALSequence() uint64 {
	// TODO: Implement proper WAL sequence tracking
	return uint64(time.Now().UnixNano())
}

func (e *FileStorageEngine) replayWAL() error {
	// TODO: Implement WAL replay
	return nil
}
