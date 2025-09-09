package storage

import (
	"fmt"
	"sync"
)

// FileTransaction implements the Transaction interface
type FileTransaction struct {
	engine   *FileStorageEngine
	id       uint64
	pages    map[uint32]*Page
	mu       sync.RWMutex
	active   bool
	readOnly bool
}

// ReadPage reads a page within the transaction
func (tx *FileTransaction) ReadPage(pageID uint32) (*Page, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if !tx.active {
		return nil, fmt.Errorf("transaction is not active")
	}

	// Check if page is in transaction cache
	if page, found := tx.pages[pageID]; found {
		return tx.copyPage(page), nil
	}

	// Read from storage engine
	page, err := tx.engine.ReadPage(pageID)
	if err != nil {
		return nil, err
	}

	// Cache in transaction
	tx.pages[pageID] = tx.copyPage(page)

	return page, nil
}

// WritePage writes a page within the transaction
func (tx *FileTransaction) WritePage(page *Page) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if !tx.active {
		return fmt.Errorf("transaction is not active")
	}

	if tx.readOnly {
		return fmt.Errorf("transaction is read-only")
	}

	// Store in transaction cache
	tx.pages[page.ID] = tx.copyPage(page)

	return nil
}

// Commit commits the transaction
func (tx *FileTransaction) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if !tx.active {
		return fmt.Errorf("transaction is not active")
	}

	// Write all pages to storage
	for _, page := range tx.pages {
		if err := tx.engine.WritePage(page); err != nil {
			// Rollback on error
			tx.active = false
			return fmt.Errorf("failed to commit page %d: %w", page.ID, err)
		}
	}

	// Mark transaction as inactive
	tx.active = false

	return nil
}

// Rollback rolls back the transaction
func (tx *FileTransaction) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if !tx.active {
		return fmt.Errorf("transaction is not active")
	}

	// Clear transaction cache
	tx.pages = make(map[uint32]*Page)

	// Mark transaction as inactive
	tx.active = false

	return nil
}

// copyPage creates a deep copy of a page
func (tx *FileTransaction) copyPage(page *Page) *Page {
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
