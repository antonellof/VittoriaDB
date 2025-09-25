// +build windows

package storage

import (
	"fmt"
	"os"
	"sync"
	"unsafe"
)

// MMapStorage implements a fallback storage for Windows (no mmap support)
type MMapStorage struct {
	file     *os.File
	data     []byte
	size     int64
	readonly bool
	mu       sync.RWMutex
}

// NewMMapStorage creates a new storage (fallback for Windows)
func NewMMapStorage(filepath string, size int64, readonly bool) (*MMapStorage, error) {
	var flag int

	if readonly {
		flag = os.O_RDONLY
	} else {
		flag = os.O_RDWR | os.O_CREATE
	}

	file, err := os.OpenFile(filepath, flag, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Get file info
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := info.Size()

	// If creating new file or file is smaller than requested size, resize it
	if !readonly && fileSize < size {
		if err := file.Truncate(size); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to truncate file: %w", err)
		}
		fileSize = size
	}

	// On Windows, we'll use regular file I/O instead of mmap
	data := make([]byte, fileSize)
	if fileSize > 0 {
		if _, err := file.ReadAt(data, 0); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
	}

	return &MMapStorage{
		file:     file,
		data:     data,
		size:     fileSize,
		readonly: readonly,
	}, nil
}

// Close closes the memory-mapped storage
func (m *MMapStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.file != nil {
		// Write data back to file if not readonly
		if !m.readonly && len(m.data) > 0 {
			if _, err := m.file.WriteAt(m.data, 0); err != nil {
				return fmt.Errorf("failed to write data: %w", err)
			}
		}
		
		if err := m.file.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
		m.file = nil
	}

	m.data = nil
	return nil
}

// Sync syncs the memory-mapped data to disk
func (m *MMapStorage) Sync() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.file == nil {
		return fmt.Errorf("storage is closed")
	}

	// Write data back to file and sync
	if !m.readonly && len(m.data) > 0 {
		if _, err := m.file.WriteAt(m.data, 0); err != nil {
			return fmt.Errorf("failed to write data: %w", err)
		}
	}

	return m.file.Sync()
}

// ReadAt reads data from the specified offset
func (m *MMapStorage) ReadAt(p []byte, offset int64) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if offset < 0 || offset >= m.size {
		return 0, fmt.Errorf("offset out of bounds")
	}

	n := copy(p, m.data[offset:])
	return n, nil
}

// WriteAt writes data at the specified offset
func (m *MMapStorage) WriteAt(p []byte, offset int64) (int, error) {
	if m.readonly {
		return 0, fmt.Errorf("storage is readonly")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if offset < 0 || offset >= m.size {
		return 0, fmt.Errorf("offset out of bounds")
	}

	n := copy(m.data[offset:], p)
	return n, nil
}

// Size returns the size of the storage
func (m *MMapStorage) Size() int64 {
	return m.size
}

// ReadVector reads a vector from the specified offset
func (m *MMapStorage) ReadVector(offset int64, vectorSize int) ([]float32, error) {
	if vectorSize <= 0 {
		return nil, fmt.Errorf("invalid vector size: %d", vectorSize)
	}

	bytesNeeded := vectorSize * 4 // 4 bytes per float32
	if offset < 0 || offset+int64(bytesNeeded) > m.size {
		return nil, fmt.Errorf("vector read out of bounds")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	vector := make([]float32, vectorSize)
	for i := 0; i < vectorSize; i++ {
		byteOffset := offset + int64(i*4)
		// Convert 4 bytes to float32
		bytes := m.data[byteOffset : byteOffset+4]
		value := *(*float32)(unsafe.Pointer(&bytes[0]))
		vector[i] = value
	}

	return vector, nil
}

// WriteVector writes a vector to the specified offset
func (m *MMapStorage) WriteVector(offset int64, vector []float32) error {
	if m.readonly {
		return fmt.Errorf("storage is readonly")
	}

	vectorSize := len(vector)
	if vectorSize == 0 {
		return fmt.Errorf("empty vector")
	}

	bytesNeeded := vectorSize * 4 // 4 bytes per float32
	if offset < 0 || offset+int64(bytesNeeded) > m.size {
		return fmt.Errorf("vector write out of bounds")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i, value := range vector {
		byteOffset := offset + int64(i*4)
		// Convert float32 to 4 bytes
		valuePtr := (*[4]byte)(unsafe.Pointer(&value))
		copy(m.data[byteOffset:byteOffset+4], valuePtr[:])
	}

	return nil
}

// Resize resizes the storage (Windows fallback implementation)
func (m *MMapStorage) Resize(newSize int64) error {
	if m.readonly {
		return fmt.Errorf("cannot resize readonly storage")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if newSize == m.size {
		return nil
	}

	// Truncate file
	if err := m.file.Truncate(newSize); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	// Resize data buffer
	if newSize > m.size {
		// Extend buffer
		newData := make([]byte, newSize)
		copy(newData, m.data)
		m.data = newData
	} else {
		// Shrink buffer
		m.data = m.data[:newSize]
	}

	m.size = newSize
	return nil
}

// ReadVectorBatch reads multiple vectors efficiently (Windows fallback)
func (m *MMapStorage) ReadVectorBatch(offsets []int64, dimensions int) ([][]float32, error) {
	if dimensions <= 0 {
		return nil, fmt.Errorf("invalid dimensions: %d", dimensions)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	vectors := make([][]float32, len(offsets))
	for i, offset := range offsets {
		vector, err := m.readVectorUnsafe(offset, dimensions)
		if err != nil {
			return nil, fmt.Errorf("failed to read vector at offset %d: %w", offset, err)
		}
		vectors[i] = vector
	}

	return vectors, nil
}

// WriteVectorBatch writes multiple vectors efficiently (Windows fallback)
func (m *MMapStorage) WriteVectorBatch(offsets []int64, vectors [][]float32) error {
	if m.readonly {
		return fmt.Errorf("storage is readonly")
	}

	if len(offsets) != len(vectors) {
		return fmt.Errorf("offsets and vectors length mismatch")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for i, offset := range offsets {
		if err := m.writeVectorUnsafe(offset, vectors[i]); err != nil {
			return fmt.Errorf("failed to write vector at offset %d: %w", offset, err)
		}
	}

	return nil
}

// readVectorUnsafe reads a vector without locking (internal use)
func (m *MMapStorage) readVectorUnsafe(offset int64, vectorSize int) ([]float32, error) {
	if vectorSize <= 0 {
		return nil, fmt.Errorf("invalid vector size: %d", vectorSize)
	}

	bytesNeeded := vectorSize * 4 // 4 bytes per float32
	if offset < 0 || offset+int64(bytesNeeded) > m.size {
		return nil, fmt.Errorf("vector read out of bounds")
	}

	vector := make([]float32, vectorSize)
	for i := 0; i < vectorSize; i++ {
		byteOffset := offset + int64(i*4)
		// Convert 4 bytes to float32
		bytes := m.data[byteOffset : byteOffset+4]
		value := *(*float32)(unsafe.Pointer(&bytes[0]))
		vector[i] = value
	}

	return vector, nil
}

// writeVectorUnsafe writes a vector without locking (internal use)
func (m *MMapStorage) writeVectorUnsafe(offset int64, vector []float32) error {
	vectorSize := len(vector)
	if vectorSize == 0 {
		return fmt.Errorf("empty vector")
	}

	bytesNeeded := vectorSize * 4 // 4 bytes per float32
	if offset < 0 || offset+int64(bytesNeeded) > m.size {
		return fmt.Errorf("vector write out of bounds")
	}

	for i, value := range vector {
		byteOffset := offset + int64(i*4)
		// Convert float32 to 4 bytes
		valuePtr := (*[4]byte)(unsafe.Pointer(&value))
		copy(m.data[byteOffset:byteOffset+4], valuePtr[:])
	}

	return nil
}

// Stats returns statistics about the storage (Windows fallback)
func (m *MMapStorage) Stats() *MMapStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &MMapStats{
		Size:     m.size,
		Readonly: m.readonly,
		IsMapped: false, // Windows fallback doesn't use actual mmap
	}
}

// MMapStats represents memory-mapped storage statistics
type MMapStats struct {
	Size     int64 `json:"size"`
	Readonly bool  `json:"readonly"`
	IsMapped bool  `json:"is_mapped"`
}
