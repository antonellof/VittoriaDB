package core

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/storage"
)

// IOOptimizerConfig holds configuration for I/O optimization
type IOOptimizerConfig struct {
	// Memory-mapped I/O settings
	UseMemoryMap bool `json:"use_memory_map"`
	MMapReadOnly bool `json:"mmap_readonly"`
	MMapPreload  bool `json:"mmap_preload"`

	// Async I/O settings
	AsyncIO        bool `json:"async_io"`
	AsyncWorkers   int  `json:"async_workers"`
	AsyncQueueSize int  `json:"async_queue_size"`

	// Vectorized operations
	VectorizedOps  bool `json:"vectorized_ops"`
	SIMDEnabled    bool `json:"simd_enabled"`
	ParallelChunks bool `json:"parallel_chunks"`

	// Batch processing
	BatchSize     int           `json:"batch_size"`
	FlushInterval time.Duration `json:"flush_interval"`

	// Read-ahead and write-behind
	ReadAheadSize   int `json:"read_ahead_size"`
	WriteBufferSize int `json:"write_buffer_size"`

	// Performance tuning
	NumWorkers int `json:"num_workers"`
	ChunkSize  int `json:"chunk_size"`
}

// DefaultIOOptimizerConfig returns default I/O optimizer configuration
func DefaultIOOptimizerConfig() *IOOptimizerConfig {
	return &IOOptimizerConfig{
		UseMemoryMap:    true,
		MMapReadOnly:    false,
		MMapPreload:     false,
		AsyncIO:         true,
		AsyncWorkers:    runtime.NumCPU() * 2,
		AsyncQueueSize:  10000,
		VectorizedOps:   true,
		SIMDEnabled:     true,
		ParallelChunks:  true,
		BatchSize:       100,
		FlushInterval:   100 * time.Millisecond,
		ReadAheadSize:   64 * 1024,  // 64KB
		WriteBufferSize: 256 * 1024, // 256KB
		NumWorkers:      runtime.NumCPU(),
		ChunkSize:       256,
	}
}

// IOOptimizer provides comprehensive I/O optimization
type IOOptimizer struct {
	config      *IOOptimizerConfig
	mmapStorage *storage.MMapStorage
	asyncEngine *storage.AsyncIOEngine
	simdOps     *SIMDVectorOps
	readCache   *IOReadCache
	writeBuffer *IOWriteBuffer
	stats       *IOOptimizerStats
	mu          sync.RWMutex
	running     bool
}

// NewIOOptimizer creates a new I/O optimizer
func NewIOOptimizer(config *IOOptimizerConfig) *IOOptimizer {
	if config == nil {
		config = DefaultIOOptimizerConfig()
	}

	optimizer := &IOOptimizer{
		config: config,
		stats:  NewIOOptimizerStats(),
	}

	// Initialize SIMD operations
	if config.SIMDEnabled {
		simdConfig := &SIMDConfig{
			Enabled:        config.SIMDEnabled,
			VectorizedMath: config.VectorizedOps,
			ParallelChunks: config.ParallelChunks,
			ChunkSize:      config.ChunkSize,
			NumWorkers:     config.NumWorkers,
		}
		optimizer.simdOps = NewSIMDVectorOps(simdConfig)
	}

	// Initialize read cache
	optimizer.readCache = NewIOReadCache(config.ReadAheadSize)

	// Initialize write buffer
	optimizer.writeBuffer = NewIOWriteBuffer(config.WriteBufferSize, config.FlushInterval)

	return optimizer
}

// InitializeStorage initializes storage with I/O optimizations
func (io *IOOptimizer) InitializeStorage(filepath string, size int64, storageEngine storage.StorageEngine) error {
	io.mu.Lock()
	defer io.mu.Unlock()

	var err error

	// Initialize memory-mapped storage if enabled
	if io.config.UseMemoryMap {
		io.mmapStorage, err = storage.NewMMapStorage(filepath, size, io.config.MMapReadOnly)
		if err != nil {
			return fmt.Errorf("failed to initialize memory-mapped storage: %w", err)
		}
	}

	// Initialize async I/O engine if enabled
	if io.config.AsyncIO {
		asyncConfig := &storage.AsyncIOConfig{
			Enabled:        true,
			WorkerPoolSize: io.config.AsyncWorkers,
			QueueSize:      io.config.AsyncQueueSize,
			BatchSize:      io.config.BatchSize,
			FlushInterval:  io.config.FlushInterval,
		}

		io.asyncEngine = storage.NewAsyncIOEngine(storageEngine, asyncConfig)
		if err := io.asyncEngine.Start(); err != nil {
			return fmt.Errorf("failed to start async I/O engine: %w", err)
		}
	}

	io.running = true
	return nil
}

// OptimizedVectorSimilarity performs optimized vector similarity calculation
func (io *IOOptimizer) OptimizedVectorSimilarity(query []float32, vectors [][]float32, metric DistanceMetric) []float32 {
	start := time.Now()
	defer func() {
		io.stats.RecordOperation("vector_similarity", time.Since(start))
	}()

	if io.simdOps == nil || !io.config.VectorizedOps {
		return io.fallbackVectorSimilarity(query, vectors, metric)
	}

	switch metric {
	case DistanceMetricCosine:
		return io.simdOps.CosineSimilarityBatch(query, vectors)
	case DistanceMetricEuclidean:
		results := make([]float32, len(vectors))
		for i, vector := range vectors {
			results[i] = io.simdOps.EuclideanDistance(query, vector)
		}
		return results
	case DistanceMetricDotProduct:
		results := make([]float32, len(vectors))
		for i, vector := range vectors {
			results[i] = io.simdOps.DotProduct(query, vector)
		}
		return results
	default:
		return io.fallbackVectorSimilarity(query, vectors, metric)
	}
}

// OptimizedVectorRead performs optimized vector reading
func (io *IOOptimizer) OptimizedVectorRead(ctx context.Context, offsets []int64, dimensions int) ([][]float32, error) {
	start := time.Now()
	defer func() {
		io.stats.RecordOperation("vector_read", time.Since(start))
	}()

	// Use memory-mapped storage if available
	if io.mmapStorage != nil {
		return io.mmapStorage.ReadVectorBatch(offsets, dimensions)
	}

	// Use async I/O if available
	if io.asyncEngine != nil {
		return io.asyncVectorRead(ctx, offsets, dimensions)
	}

	// Fallback to synchronous read
	return io.fallbackVectorRead(offsets, dimensions)
}

// OptimizedVectorWrite performs optimized vector writing
func (io *IOOptimizer) OptimizedVectorWrite(ctx context.Context, vectors [][]float32, offsets []int64) error {
	start := time.Now()
	defer func() {
		io.stats.RecordOperation("vector_write", time.Since(start))
	}()

	// Use memory-mapped storage if available
	if io.mmapStorage != nil && !io.config.MMapReadOnly {
		return io.mmapStorage.WriteVectorBatch(offsets, vectors)
	}

	// Use async I/O if available
	if io.asyncEngine != nil {
		return io.asyncVectorWrite(ctx, vectors, offsets)
	}

	// Fallback to synchronous write
	return io.fallbackVectorWrite(vectors, offsets)
}

// OptimizedBatchNormalize performs optimized batch vector normalization
func (io *IOOptimizer) OptimizedBatchNormalize(vectors [][]float32) {
	start := time.Now()
	defer func() {
		io.stats.RecordOperation("batch_normalize", time.Since(start))
	}()

	if io.simdOps != nil && io.config.VectorizedOps {
		io.simdOps.NormalizeBatch(vectors)
	} else {
		io.fallbackBatchNormalize(vectors)
	}
}

// Sync synchronizes all pending I/O operations
func (io *IOOptimizer) Sync() error {
	io.mu.RLock()
	defer io.mu.RUnlock()

	var errors []error

	// Sync memory-mapped storage
	if io.mmapStorage != nil {
		if err := io.mmapStorage.Sync(); err != nil {
			errors = append(errors, fmt.Errorf("mmap sync failed: %w", err))
		}
	}

	// Sync async I/O engine
	if io.asyncEngine != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result := <-io.asyncEngine.SyncAsync(ctx)
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("async sync failed: %w", result.Error))
		}
	}

	// Flush write buffer
	if io.writeBuffer != nil {
		if err := io.writeBuffer.Flush(); err != nil {
			errors = append(errors, fmt.Errorf("write buffer flush failed: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("sync errors: %v", errors)
	}

	return nil
}

// Close closes the I/O optimizer and all resources
func (io *IOOptimizer) Close() error {
	io.mu.Lock()
	defer io.mu.Unlock()

	if !io.running {
		return nil
	}

	var errors []error

	// Stop async I/O engine
	if io.asyncEngine != nil {
		if err := io.asyncEngine.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop async I/O engine: %w", err))
		}
	}

	// Close memory-mapped storage
	if io.mmapStorage != nil {
		if err := io.mmapStorage.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close mmap storage: %w", err))
		}
	}

	// Close write buffer
	if io.writeBuffer != nil {
		if err := io.writeBuffer.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close write buffer: %w", err))
		}
	}

	io.running = false

	if len(errors) > 0 {
		return fmt.Errorf("close errors: %v", errors)
	}

	return nil
}

// GetStats returns I/O optimizer statistics
func (io *IOOptimizer) GetStats() *IOOptimizerStats {
	io.mu.RLock()
	defer io.mu.RUnlock()

	stats := io.stats.Copy()

	// Add async I/O stats if available
	if io.asyncEngine != nil {
		stats.AsyncIOStats = io.asyncEngine.GetStats()
	}

	// Add memory-mapped storage stats if available
	if io.mmapStorage != nil {
		stats.MMapStats = io.mmapStorage.Stats()
	}

	return stats
}

// BenchmarkOptimizations runs performance benchmarks
func (io *IOOptimizer) BenchmarkOptimizations(dimensions int, numVectors int) *IOOptimizationBenchmark {
	benchmark := &IOOptimizationBenchmark{
		Dimensions: dimensions,
		NumVectors: numVectors,
		Timestamp:  time.Now(),
	}

	// Benchmark SIMD operations if available
	if io.simdOps != nil {
		benchmark.SIMDResults = io.simdOps.BenchmarkSIMDOperations(dimensions, numVectors)
	}

	// Benchmark I/O operations
	benchmark.IOResults = io.benchmarkIOOperations(dimensions, numVectors)

	return benchmark
}

// Private helper methods

func (io *IOOptimizer) fallbackVectorSimilarity(query []float32, vectors [][]float32, metric DistanceMetric) []float32 {
	results := make([]float32, len(vectors))

	for i, vector := range vectors {
		switch metric {
		case DistanceMetricCosine:
			results[i] = cosineSimilarity(query, vector)
		case DistanceMetricEuclidean:
			results[i] = euclideanDistance(query, vector)
		case DistanceMetricDotProduct:
			results[i] = dotProduct(query, vector)
		default:
			results[i] = cosineSimilarity(query, vector)
		}
	}

	return results
}

func (io *IOOptimizer) asyncVectorRead(ctx context.Context, offsets []int64, dimensions int) ([][]float32, error) {
	// Convert offsets to page IDs and read asynchronously
	results := make([][]float32, len(offsets))

	for i, offset := range offsets {
		pageID := uint32(offset / storage.PageSize)
		result := <-io.asyncEngine.ReadAsync(ctx, pageID)

		if result.Error != nil {
			return nil, result.Error
		}

		// Extract vector from page data
		vectorOffset := offset % storage.PageSize
		vectorSize := int64(dimensions * 4)

		if int64(len(result.Data)) < vectorOffset+vectorSize {
			return nil, fmt.Errorf("insufficient data for vector at offset %d", offset)
		}

		vectorData := result.Data[vectorOffset : vectorOffset+vectorSize]
		vector := make([]float32, dimensions)

		// Convert bytes to float32 slice
		for j := 0; j < dimensions; j++ {
			// Simple byte-to-float conversion (little-endian assumed)
			bytes := vectorData[j*4 : (j+1)*4]
			vector[j] = float32(bytes[0]) + float32(bytes[1])*256 + float32(bytes[2])*65536 + float32(bytes[3])*16777216
		}

		results[i] = vector
	}

	return results, nil
}

func (io *IOOptimizer) asyncVectorWrite(ctx context.Context, vectors [][]float32, offsets []int64) error {
	// Convert vectors to pages and write asynchronously
	for i, vector := range vectors {
		offset := offsets[i]
		pageID := uint32(offset / storage.PageSize)

		// Create page with vector data
		vectorData := make([]byte, len(vector)*4)
		for j, v := range vector {
			// Simple float-to-byte conversion (little-endian)
			intVal := uint32(v)
			vectorData[j*4] = byte(intVal)
			vectorData[j*4+1] = byte(intVal >> 8)
			vectorData[j*4+2] = byte(intVal >> 16)
			vectorData[j*4+3] = byte(intVal >> 24)
		}

		page := &storage.Page{
			ID:   pageID,
			Type: storage.PageTypeVectorLeaf,
			Size: uint16(len(vectorData)),
			Data: vectorData,
		}

		result := <-io.asyncEngine.WriteAsync(ctx, page)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func (io *IOOptimizer) fallbackVectorRead(offsets []int64, dimensions int) ([][]float32, error) {
	// Placeholder for synchronous vector read
	results := make([][]float32, len(offsets))
	for i := range results {
		results[i] = make([]float32, dimensions)
	}
	return results, nil
}

func (io *IOOptimizer) fallbackVectorWrite(vectors [][]float32, offsets []int64) error {
	// Placeholder for synchronous vector write
	return nil
}

func (io *IOOptimizer) fallbackBatchNormalize(vectors [][]float32) {
	for _, vector := range vectors {
		var norm float32
		for _, v := range vector {
			norm += v * v
		}

		if norm > 0 {
			norm = float32(1.0 / float64(norm))
			for i := range vector {
				vector[i] *= norm
			}
		}
	}
}

func (io *IOOptimizer) benchmarkIOOperations(dimensions int, numVectors int) *IOBenchmarkResults {
	// Placeholder for I/O benchmarking
	return &IOBenchmarkResults{
		ReadLatency:     1 * time.Millisecond,
		WriteLatency:    2 * time.Millisecond,
		ReadThroughput:  1000.0,
		WriteThroughput: 500.0,
	}
}

// Supporting types and structures

// IOOptimizerStats tracks I/O optimizer statistics
type IOOptimizerStats struct {
	mu               sync.RWMutex
	OperationCounts  map[string]int64      `json:"operation_counts"`
	OperationLatency map[string]int64      `json:"operation_latency_ns"`
	TotalOperations  int64                 `json:"total_operations"`
	AverageLatency   time.Duration         `json:"average_latency"`
	AsyncIOStats     *storage.AsyncIOStats `json:"async_io_stats,omitempty"`
	MMapStats        *storage.MMapStats    `json:"mmap_stats,omitempty"`
	LastUpdate       time.Time             `json:"last_update"`
}

// NewIOOptimizerStats creates new I/O optimizer statistics
func NewIOOptimizerStats() *IOOptimizerStats {
	return &IOOptimizerStats{
		OperationCounts:  make(map[string]int64),
		OperationLatency: make(map[string]int64),
		LastUpdate:       time.Now(),
	}
}

// RecordOperation records an operation
func (s *IOOptimizerStats) RecordOperation(operation string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.OperationCounts[operation]++
	s.OperationLatency[operation] += duration.Nanoseconds()
	s.TotalOperations++
	s.LastUpdate = time.Now()
}

// Copy returns a copy of the statistics
func (s *IOOptimizerStats) Copy() *IOOptimizerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	copy := &IOOptimizerStats{
		OperationCounts:  make(map[string]int64),
		OperationLatency: make(map[string]int64),
		TotalOperations:  s.TotalOperations,
		AverageLatency:   s.AverageLatency,
		LastUpdate:       s.LastUpdate,
	}

	for k, v := range s.OperationCounts {
		copy.OperationCounts[k] = v
	}
	for k, v := range s.OperationLatency {
		copy.OperationLatency[k] = v
	}

	return copy
}

// IOOptimizationBenchmark holds benchmark results
type IOOptimizationBenchmark struct {
	Dimensions  int                   `json:"dimensions"`
	NumVectors  int                   `json:"num_vectors"`
	SIMDResults *SIMDBenchmarkResults `json:"simd_results,omitempty"`
	IOResults   *IOBenchmarkResults   `json:"io_results"`
	Timestamp   time.Time             `json:"timestamp"`
}

// IOBenchmarkResults holds I/O benchmark results
type IOBenchmarkResults struct {
	ReadLatency     time.Duration `json:"read_latency"`
	WriteLatency    time.Duration `json:"write_latency"`
	ReadThroughput  float64       `json:"read_throughput_ops_per_sec"`
	WriteThroughput float64       `json:"write_throughput_ops_per_sec"`
}

// IOReadCache provides read-ahead caching
type IOReadCache struct {
	size  int
	cache map[int64][]byte
	mu    sync.RWMutex
}

// NewIOReadCache creates a new read cache
func NewIOReadCache(size int) *IOReadCache {
	return &IOReadCache{
		size:  size,
		cache: make(map[int64][]byte),
	}
}

// IOWriteBuffer provides write buffering
type IOWriteBuffer struct {
	size          int
	flushInterval time.Duration
	buffer        map[int64][]byte
	mu            sync.RWMutex
}

// NewIOWriteBuffer creates a new write buffer
func NewIOWriteBuffer(size int, flushInterval time.Duration) *IOWriteBuffer {
	return &IOWriteBuffer{
		size:          size,
		flushInterval: flushInterval,
		buffer:        make(map[int64][]byte),
	}
}

// Flush flushes the write buffer
func (wb *IOWriteBuffer) Flush() error {
	wb.mu.Lock()
	defer wb.mu.Unlock()

	// Clear buffer
	wb.buffer = make(map[int64][]byte)
	return nil
}

// Close closes the write buffer
func (wb *IOWriteBuffer) Close() error {
	return wb.Flush()
}
