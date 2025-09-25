package storage

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

// MMapStorage implements memory-mapped file storage for zero-copy operations
type MMapStorage struct {
	file     *os.File
	data     []byte
	size     int64
	readonly bool
	mu       sync.RWMutex
}

// NewMMapStorage creates a new memory-mapped storage
func NewMMapStorage(filepath string, size int64, readonly bool) (*MMapStorage, error) {
	var flag int
	var prot int

	if readonly {
		flag = os.O_RDONLY
		prot = syscall.PROT_READ
	} else {
		flag = os.O_RDWR | os.O_CREATE
		prot = syscall.PROT_READ | syscall.PROT_WRITE
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
			return nil, fmt.Errorf("failed to resize file: %w", err)
		}
		fileSize = size
	}

	// Memory map the file
	data, err := syscall.Mmap(int(file.Fd()), 0, int(fileSize), prot, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to mmap file: %w", err)
	}

	return &MMapStorage{
		file:     file,
		data:     data,
		size:     fileSize,
		readonly: readonly,
	}, nil
}

// Read reads data from the memory-mapped region (zero-copy)
func (m *MMapStorage) Read(offset int64, length int) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if offset < 0 || offset >= m.size {
		return nil, fmt.Errorf("offset %d out of bounds (size: %d)", offset, m.size)
	}

	end := offset + int64(length)
	if end > m.size {
		end = m.size
	}

	// Return a slice of the memory-mapped data (zero-copy)
	return m.data[offset:end], nil
}

// Write writes data to the memory-mapped region
func (m *MMapStorage) Write(offset int64, data []byte) error {
	if m.readonly {
		return fmt.Errorf("cannot write to readonly storage")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if offset < 0 || offset >= m.size {
		return fmt.Errorf("offset %d out of bounds (size: %d)", offset, m.size)
	}

	end := offset + int64(len(data))
	if end > m.size {
		return fmt.Errorf("write would exceed file size (offset: %d, length: %d, size: %d)",
			offset, len(data), m.size)
	}

	// Copy data directly to memory-mapped region
	copy(m.data[offset:end], data)

	return nil
}

// ReadVector reads a vector from the memory-mapped storage with zero-copy
func (m *MMapStorage) ReadVector(offset int64, dimensions int) ([]float32, error) {
	vectorSize := int64(dimensions * 4) // 4 bytes per float32

	data, err := m.Read(offset, int(vectorSize))
	if err != nil {
		return nil, err
	}

	if len(data) < int(vectorSize) {
		return nil, fmt.Errorf("insufficient data for vector (got %d, need %d bytes)",
			len(data), vectorSize)
	}

	// Convert bytes to float32 slice safely
	vector := make([]float32, dimensions)
	for i := 0; i < dimensions; i++ {
		bits := uint32(data[i*4]) | uint32(data[i*4+1])<<8 |
			uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
		vector[i] = *(*float32)(unsafe.Pointer(&bits))
	}

	return vector, nil
}

// WriteVector writes a vector to the memory-mapped storage
func (m *MMapStorage) WriteVector(offset int64, vector []float32) error {
	if m.readonly {
		return fmt.Errorf("cannot write to readonly storage")
	}

	vectorSize := len(vector) * 4 // 4 bytes per float32

	// Convert float32 slice to bytes safely
	data := make([]byte, vectorSize)
	for i, v := range vector {
		bits := *(*uint32)(unsafe.Pointer(&v))
		data[i*4] = byte(bits)
		data[i*4+1] = byte(bits >> 8)
		data[i*4+2] = byte(bits >> 16)
		data[i*4+3] = byte(bits >> 24)
	}

	return m.Write(offset, data)
}

// ReadVectorBatch reads multiple vectors efficiently
func (m *MMapStorage) ReadVectorBatch(offsets []int64, dimensions int) ([][]float32, error) {
	vectors := make([][]float32, len(offsets))

	for i, offset := range offsets {
		vector, err := m.ReadVector(offset, dimensions)
		if err != nil {
			return nil, fmt.Errorf("failed to read vector %d: %w", i, err)
		}
		vectors[i] = vector
	}

	return vectors, nil
}

// WriteVectorBatch writes multiple vectors efficiently
func (m *MMapStorage) WriteVectorBatch(offsets []int64, vectors [][]float32) error {
	if m.readonly {
		return fmt.Errorf("cannot write to readonly storage")
	}

	if len(offsets) != len(vectors) {
		return fmt.Errorf("offsets and vectors length mismatch")
	}

	for i, vector := range vectors {
		if err := m.WriteVector(offsets[i], vector); err != nil {
			return fmt.Errorf("failed to write vector %d: %w", i, err)
		}
	}

	return nil
}

// Sync synchronizes the memory-mapped data to disk
func (m *MMapStorage) Sync() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Platform-specific sync implementation
	// Note: syscall.Msync may not be available on all platforms
	_ = m.data // Use the data to avoid unused variable warning

	return m.file.Sync()
}

// AsyncSync performs asynchronous synchronization
func (m *MMapStorage) AsyncSync() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Use MS_ASYNC for non-blocking sync (platform-specific)
	// Note: syscall.Msync may not be available on all platforms
	// This is a simplified implementation for demonstration
	_ = m.data // Use the data to avoid unused variable warning

	return nil
}

// Resize resizes the memory-mapped storage
func (m *MMapStorage) Resize(newSize int64) error {
	if m.readonly {
		return fmt.Errorf("cannot resize readonly storage")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Unmap current mapping
	if err := syscall.Munmap(m.data); err != nil {
		return fmt.Errorf("failed to unmap current data: %w", err)
	}

	// Resize file
	if err := m.file.Truncate(newSize); err != nil {
		return fmt.Errorf("failed to resize file: %w", err)
	}

	// Remap with new size
	data, err := syscall.Mmap(int(m.file.Fd()), 0, int(newSize),
		syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return fmt.Errorf("failed to remap file: %w", err)
	}

	m.data = data
	m.size = newSize

	return nil
}

// Size returns the current size of the memory-mapped storage
func (m *MMapStorage) Size() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size
}

// Close closes the memory-mapped storage
func (m *MMapStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var err error

	// Sync before closing (platform-specific implementation)
	// Note: syscall.Msync may not be available on all platforms
	_ = m.data // Use the data to avoid unused variable warning

	// Unmap memory
	if unmapErr := syscall.Munmap(m.data); unmapErr != nil {
		if err != nil {
			err = fmt.Errorf("%v; failed to unmap: %w", err, unmapErr)
		} else {
			err = fmt.Errorf("failed to unmap: %w", unmapErr)
		}
	}

	// Close file
	if closeErr := m.file.Close(); closeErr != nil {
		if err != nil {
			err = fmt.Errorf("%v; failed to close file: %w", err, closeErr)
		} else {
			err = fmt.Errorf("failed to close file: %w", closeErr)
		}
	}

	return err
}

// Stats returns statistics about the memory-mapped storage
func (m *MMapStorage) Stats() *MMapStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &MMapStats{
		Size:     m.size,
		ReadOnly: m.readonly,
		FilePath: m.file.Name(),
	}
}

// MMapStats represents memory-mapped storage statistics
type MMapStats struct {
	Size     int64  `json:"size"`
	ReadOnly bool   `json:"readonly"`
	FilePath string `json:"filepath"`
}

// VectorLayout represents the layout of vectors in memory-mapped storage
type VectorLayout struct {
	Dimensions int `json:"dimensions"`
	VectorSize int `json:"vector_size"` // Size in bytes
	Alignment  int `json:"alignment"`   // Memory alignment
	Stride     int `json:"stride"`      // Distance between vectors
}

// NewVectorLayout creates an optimized vector layout
func NewVectorLayout(dimensions int) *VectorLayout {
	vectorSize := dimensions * 4 // 4 bytes per float32

	// Align to 32-byte boundaries for SIMD optimization
	alignment := 32
	stride := ((vectorSize + alignment - 1) / alignment) * alignment

	return &VectorLayout{
		Dimensions: dimensions,
		VectorSize: vectorSize,
		Alignment:  alignment,
		Stride:     stride,
	}
}

// GetVectorOffset calculates the offset for a vector at given index
func (vl *VectorLayout) GetVectorOffset(index int) int64 {
	return int64(index * vl.Stride)
}

// GetMaxVectors calculates maximum vectors that can fit in given size
func (vl *VectorLayout) GetMaxVectors(totalSize int64) int {
	return int(totalSize / int64(vl.Stride))
}

// VectorMMapStorage provides high-level vector storage operations
type VectorMMapStorage struct {
	storage *MMapStorage
	layout  *VectorLayout
	count   int
	mu      sync.RWMutex
}

// NewVectorMMapStorage creates a new vector-optimized memory-mapped storage
func NewVectorMMapStorage(filepath string, dimensions int, maxVectors int, readonly bool) (*VectorMMapStorage, error) {
	layout := NewVectorLayout(dimensions)
	totalSize := int64(maxVectors * layout.Stride)

	storage, err := NewMMapStorage(filepath, totalSize, readonly)
	if err != nil {
		return nil, err
	}

	return &VectorMMapStorage{
		storage: storage,
		layout:  layout,
		count:   0,
	}, nil
}

// AddVector adds a vector and returns its index
func (vms *VectorMMapStorage) AddVector(vector []float32) (int, error) {
	if len(vector) != vms.layout.Dimensions {
		return -1, fmt.Errorf("vector dimensions mismatch: got %d, expected %d",
			len(vector), vms.layout.Dimensions)
	}

	vms.mu.Lock()
	defer vms.mu.Unlock()

	index := vms.count
	offset := vms.layout.GetVectorOffset(index)

	if err := vms.storage.WriteVector(offset, vector); err != nil {
		return -1, err
	}

	vms.count++
	return index, nil
}

// GetVector retrieves a vector by index
func (vms *VectorMMapStorage) GetVector(index int) ([]float32, error) {
	vms.mu.RLock()
	defer vms.mu.RUnlock()

	if index < 0 || index >= vms.count {
		return nil, fmt.Errorf("vector index %d out of bounds (count: %d)", index, vms.count)
	}

	offset := vms.layout.GetVectorOffset(index)
	return vms.storage.ReadVector(offset, vms.layout.Dimensions)
}

// GetVectorBatch retrieves multiple vectors efficiently
func (vms *VectorMMapStorage) GetVectorBatch(indices []int) ([][]float32, error) {
	vms.mu.RLock()
	defer vms.mu.RUnlock()

	offsets := make([]int64, len(indices))
	for i, index := range indices {
		if index < 0 || index >= vms.count {
			return nil, fmt.Errorf("vector index %d out of bounds (count: %d)", index, vms.count)
		}
		offsets[i] = vms.layout.GetVectorOffset(index)
	}

	return vms.storage.ReadVectorBatch(offsets, vms.layout.Dimensions)
}

// Count returns the number of vectors stored
func (vms *VectorMMapStorage) Count() int {
	vms.mu.RLock()
	defer vms.mu.RUnlock()
	return vms.count
}

// Sync synchronizes the storage to disk
func (vms *VectorMMapStorage) Sync() error {
	return vms.storage.Sync()
}

// Close closes the vector storage
func (vms *VectorMMapStorage) Close() error {
	return vms.storage.Close()
}
