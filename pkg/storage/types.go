package storage

import (
	"context"
	"io"
	"time"
)

// PageSize represents the size of a storage page (4KB)
const PageSize = 4096

// PageType represents different types of pages
type PageType uint8

const (
	PageTypeHeader     PageType = 1
	PageTypeSchema     PageType = 2
	PageTypeVectorLeaf PageType = 3
	PageTypeVectorNode PageType = 4
	PageTypeIndex      PageType = 5
	PageTypeFreeList   PageType = 6
	PageTypeWAL        PageType = 7
)

// Page represents a storage page
type Page struct {
	ID       uint32    `json:"id"`
	Type     PageType  `json:"type"`
	Size     uint16    `json:"size"`
	Flags    uint16    `json:"flags"`
	LSN      uint64    `json:"lsn"`      // Log Sequence Number
	Checksum uint32    `json:"checksum"`
	Data     []byte    `json:"data"`
}

// FileHeader represents the database file header
type FileHeader struct {
	Magic      [8]byte  // "VITTORIA"
	Version    uint32   // File format version
	PageSize   uint32   // Page size (4096)
	PageCount  uint64   // Total number of pages
	RootPage   uint32   // Root page ID
	FreeList   uint32   // Free list page ID
	WALOffset  uint64   // WAL file offset
	Created    int64    // Creation timestamp
	Modified   int64    // Last modification timestamp
	Checksum   uint32   // Header checksum
}

// WALEntry represents a write-ahead log entry
type WALEntry struct {
	Sequence  uint64    `json:"sequence"`
	Type      WALOpType `json:"type"`
	PageID    uint32    `json:"page_id"`
	Data      []byte    `json:"data"`
	Timestamp int64     `json:"timestamp"`
	Checksum  uint32    `json:"checksum"`
}

// WALOpType represents WAL operation types
type WALOpType uint8

const (
	WALOpInsert WALOpType = 1
	WALOpUpdate WALOpType = 2
	WALOpDelete WALOpType = 3
	WALOpCommit WALOpType = 4
)

// StorageStats represents storage statistics
type StorageStats struct {
	TotalPages   uint64  `json:"total_pages"`
	UsedPages    uint64  `json:"used_pages"`
	FreePages    uint64  `json:"free_pages"`
	PageSize     uint32  `json:"page_size"`
	FileSize     int64   `json:"file_size"`
	CacheHitRate float64 `json:"cache_hit_rate"`
	WALSize      int64   `json:"wal_size"`
}

// StorageEngine handles persistent storage
type StorageEngine interface {
	// Lifecycle
	Open(filepath string) error
	Close() error
	Sync() error

	// Page operations
	ReadPage(pageID uint32) (*Page, error)
	WritePage(page *Page) error
	AllocatePage() (uint32, error)
	FreePage(pageID uint32) error

	// Transactions
	BeginTx() (Transaction, error)

	// Maintenance
	Compact() error
	Stats() *StorageStats
}

// Transaction provides ACID operations
type Transaction interface {
	ReadPage(pageID uint32) (*Page, error)
	WritePage(page *Page) error
	Commit() error
	Rollback() error
}

// WAL interface for write-ahead logging
type WAL interface {
	Open(filepath string) error
	Close() error
	Append(entry *WALEntry) error
	Replay(handler func(*WALEntry) error) error
	Checkpoint(pageID uint32) error
	Truncate(beforeSeq uint64) error
}

// PageCache interface for page caching
type PageCache interface {
	Get(pageID uint32) (*Page, bool)
	Put(pageID uint32, page *Page)
	Remove(pageID uint32)
	Clear()
	Stats() CacheStats
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size     int     `json:"size"`
	Capacity int     `json:"capacity"`
	Hits     uint64  `json:"hits"`
	Misses   uint64  `json:"misses"`
	HitRate  float64 `json:"hit_rate"`
}

// BTree interface for B-tree operations
type BTree interface {
	Insert(key []byte, value []byte) error
	Get(key []byte) ([]byte, error)
	Delete(key []byte) error
	Scan(startKey []byte, endKey []byte, handler func(key, value []byte) error) error
	Stats() BTreeStats
}

// BTreeStats represents B-tree statistics
type BTreeStats struct {
	Height    int    `json:"height"`
	NodeCount int    `json:"node_count"`
	KeyCount  int64  `json:"key_count"`
	FillRate  float64 `json:"fill_rate"`
}
