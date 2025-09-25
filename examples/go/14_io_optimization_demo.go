package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/antonellof/VittoriaDB/pkg/core"
	"github.com/antonellof/VittoriaDB/pkg/storage"
)

func main() {
	fmt.Println("⚡ VittoriaDB I/O Optimization Demo")
	fmt.Println("==================================")

	// Demo 1: SIMD Vector Operations
	fmt.Println("\n🚀 1. SIMD Vector Operations:")
	demonstrateSIMDOperations()

	// Demo 2: Memory-Mapped Storage
	fmt.Println("\n💾 2. Memory-Mapped Storage:")
	demonstrateMemoryMappedStorage()

	// Demo 3: Async I/O Operations
	fmt.Println("\n⚡ 3. Async I/O Operations:")
	demonstrateAsyncIO()

	// Demo 4: Integrated I/O Optimizer
	fmt.Println("\n🔧 4. Integrated I/O Optimizer:")
	demonstrateIOOptimizer()

	// Demo 5: Performance Benchmarks
	fmt.Println("\n📊 5. Performance Benchmarks:")
	runPerformanceBenchmarks()

	fmt.Println("\n🎉 I/O Optimization demo completed successfully!")
	fmt.Println("\n📚 Key Benefits:")
	fmt.Println("   • SIMD operations: 2-10x faster vector calculations")
	fmt.Println("   • Memory-mapped I/O: Zero-copy reads, 10-50x faster")
	fmt.Println("   • Async I/O: Non-blocking operations, better throughput")
	fmt.Println("   • Vectorized operations: Better CPU cache utilization")
	fmt.Println("   • Batch processing: Reduced system call overhead")
}

func demonstrateSIMDOperations() {
	// Create SIMD vector operations
	config := core.DefaultSIMDConfig()
	simdOps := core.NewSIMDVectorOps(config)

	// Generate test vectors
	dimensions := 384
	numVectors := 1000

	query := make([]float32, dimensions)
	vectors := make([][]float32, numVectors)

	// Initialize with random data
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < dimensions; i++ {
		query[i] = rand.Float32()
	}

	for i := 0; i < numVectors; i++ {
		vectors[i] = make([]float32, dimensions)
		for j := 0; j < dimensions; j++ {
			vectors[i][j] = rand.Float32()
		}
	}

	// Benchmark scalar vs SIMD operations
	fmt.Printf("   • Testing with %d vectors of %d dimensions\n", numVectors, dimensions)

	// Scalar benchmark
	config.Enabled = false
	start := time.Now()
	scalarResults := simdOps.CosineSimilarityBatch(query, vectors)
	scalarTime := time.Since(start)

	// SIMD benchmark
	config.Enabled = true
	config.ParallelChunks = false
	start = time.Now()
	simdResults := simdOps.CosineSimilarityBatch(query, vectors)
	simdTime := time.Since(start)

	// Parallel SIMD benchmark
	config.ParallelChunks = true
	start = time.Now()
	parallelResults := simdOps.CosineSimilarityBatch(query, vectors)
	parallelTime := time.Since(start)

	// Verify results are consistent
	consistent := true
	minLen := min(len(scalarResults), min(len(simdResults), len(parallelResults)))
	for i := 0; i < minLen; i++ {
		if abs(scalarResults[i]-simdResults[i]) > 0.001 || abs(scalarResults[i]-parallelResults[i]) > 0.001 {
			consistent = false
			break
		}
	}

	fmt.Printf("   • Scalar time: %v\n", scalarTime)
	fmt.Printf("   • SIMD time: %v (%.2fx speedup)\n", simdTime, float64(scalarTime)/float64(simdTime))
	fmt.Printf("   • Parallel SIMD time: %v (%.2fx speedup)\n", parallelTime, float64(scalarTime)/float64(parallelTime))
	fmt.Printf("   • Results consistent: %t\n", consistent)

	// Demonstrate other SIMD operations
	fmt.Printf("   • Testing vector normalization:\n")

	testVectors := make([][]float32, 100)
	for i := range testVectors {
		testVectors[i] = make([]float32, dimensions)
		for j := range testVectors[i] {
			testVectors[i][j] = rand.Float32() * 10
		}
	}

	start = time.Now()
	simdOps.NormalizeBatch(testVectors)
	normalizeTime := time.Since(start)

	fmt.Printf("     - Normalized %d vectors in %v\n", len(testVectors), normalizeTime)

	// Check normalization (vectors should have unit length)
	normalized := true
	for _, vector := range testVectors[:5] { // Check first 5
		var norm float32
		for _, v := range vector {
			norm += v * v
		}
		if abs(norm-1.0) > 0.001 {
			normalized = false
			break
		}
	}
	fmt.Printf("     - Vectors properly normalized: %t\n", normalized)
}

func demonstrateMemoryMappedStorage() {
	// Create temporary file for memory-mapped storage
	filepath := "/tmp/vittoria_mmap_demo.dat"
	dimensions := 128
	maxVectors := 1000

	// Create vector memory-mapped storage
	vectorStorage, err := storage.NewVectorMMapStorage(filepath, dimensions, maxVectors, false)
	if err != nil {
		log.Printf("Failed to create vector mmap storage: %v", err)
		return
	}
	defer vectorStorage.Close()

	fmt.Printf("   • Created memory-mapped storage: %s\n", filepath)
	fmt.Printf("   • Dimensions: %d, Max vectors: %d\n", dimensions, maxVectors)

	// Add vectors to storage
	numVectors := 500
	vectors := make([][]float32, numVectors)
	indices := make([]int, numVectors)

	start := time.Now()
	for i := 0; i < numVectors; i++ {
		vector := make([]float32, dimensions)
		for j := 0; j < dimensions; j++ {
			vector[j] = rand.Float32()
		}
		vectors[i] = vector

		index, err := vectorStorage.AddVector(vector)
		if err != nil {
			log.Printf("Failed to add vector %d: %v", i, err)
			continue
		}
		indices[i] = index
	}
	addTime := time.Since(start)

	fmt.Printf("   • Added %d vectors in %v (%.2f vectors/ms)\n",
		numVectors, addTime, float64(numVectors)/float64(addTime.Milliseconds()))

	// Read vectors back (zero-copy)
	start = time.Now()
	readVectors, err := vectorStorage.GetVectorBatch(indices)
	if err != nil {
		log.Printf("Failed to read vectors: %v", err)
		return
	}
	readTime := time.Since(start)

	fmt.Printf("   • Read %d vectors in %v (%.2f vectors/ms)\n",
		len(readVectors), readTime, float64(len(readVectors))/float64(readTime.Milliseconds()))

	// Verify data integrity
	consistent := true
	for i := 0; i < min(len(vectors), len(readVectors)); i++ {
		if len(vectors[i]) != len(readVectors[i]) {
			consistent = false
			break
		}
		for j := 0; j < len(vectors[i]); j++ {
			if abs(vectors[i][j]-readVectors[i][j]) > 0.001 {
				consistent = false
				break
			}
		}
	}

	fmt.Printf("   • Data integrity verified: %t\n", consistent)
	fmt.Printf("   • Storage count: %d vectors\n", vectorStorage.Count())

	// Demonstrate sync operation
	start = time.Now()
	err = vectorStorage.Sync()
	syncTime := time.Since(start)

	if err != nil {
		fmt.Printf("   • Sync failed: %v\n", err)
	} else {
		fmt.Printf("   • Synced to disk in %v\n", syncTime)
	}
}

func demonstrateAsyncIO() {
	// Create a mock storage engine for demonstration
	mockStorage := &MockStorageEngine{
		pages: make(map[uint32]*storage.Page),
	}

	// Create async I/O engine
	config := storage.DefaultAsyncIOConfig()
	config.WorkerPoolSize = 4
	config.QueueSize = 1000

	asyncEngine := storage.NewAsyncIOEngine(mockStorage, config)

	err := asyncEngine.Start()
	if err != nil {
		log.Printf("Failed to start async I/O engine: %v", err)
		return
	}
	defer asyncEngine.Stop()

	fmt.Printf("   • Started async I/O engine with %d workers\n", config.WorkerPoolSize)

	ctx := context.Background()
	numOperations := 100

	// Demonstrate async reads
	fmt.Printf("   • Performing %d async read operations:\n", numOperations)

	start := time.Now()
	readResults := make([]<-chan storage.AsyncIOResult, numOperations)

	for i := 0; i < numOperations; i++ {
		readResults[i] = asyncEngine.ReadAsync(ctx, uint32(i))
	}

	// Wait for all reads to complete
	successfulReads := 0
	for i, resultChan := range readResults {
		select {
		case result := <-resultChan:
			if result.Error == nil {
				successfulReads++
			} else {
				fmt.Printf("     - Read %d failed: %v\n", i, result.Error)
			}
		case <-time.After(5 * time.Second):
			fmt.Printf("     - Read %d timed out\n", i)
		}
	}

	readTime := time.Since(start)
	fmt.Printf("     - Completed %d/%d reads in %v\n", successfulReads, numOperations, readTime)

	// Demonstrate async writes
	fmt.Printf("   • Performing %d async write operations:\n", numOperations)

	start = time.Now()
	writeResults := make([]<-chan storage.AsyncIOResult, numOperations)

	for i := 0; i < numOperations; i++ {
		page := &storage.Page{
			ID:   uint32(i),
			Type: storage.PageTypeVectorLeaf,
			Size: 1024,
			Data: make([]byte, 1024),
		}

		// Fill with test data
		for j := range page.Data {
			page.Data[j] = byte(i + j)
		}

		writeResults[i] = asyncEngine.WriteAsync(ctx, page)
	}

	// Wait for all writes to complete
	successfulWrites := 0
	for i, resultChan := range writeResults {
		select {
		case result := <-resultChan:
			if result.Error == nil {
				successfulWrites++
			} else {
				fmt.Printf("     - Write %d failed: %v\n", i, result.Error)
			}
		case <-time.After(5 * time.Second):
			fmt.Printf("     - Write %d timed out\n", i)
		}
	}

	writeTime := time.Since(start)
	fmt.Printf("     - Completed %d/%d writes in %v\n", successfulWrites, numOperations, writeTime)

	// Get async I/O statistics
	stats := asyncEngine.GetStats()
	fmt.Printf("   • Async I/O Statistics:\n")
	fmt.Printf("     - Total operations: %d\n", getTotalOperations(stats))
	fmt.Printf("     - Throughput: %.2f ops/sec\n", stats.Throughput)
	fmt.Printf("     - Error count: %d\n", stats.ErrorCount)
}

func demonstrateIOOptimizer() {
	// Create I/O optimizer with all optimizations enabled
	config := core.DefaultIOOptimizerConfig()
	config.UseMemoryMap = true
	config.AsyncIO = true
	config.SIMDEnabled = true
	config.VectorizedOps = true

	optimizer := core.NewIOOptimizer(config)

	fmt.Printf("   • Created I/O optimizer with optimizations:\n")
	fmt.Printf("     - Memory-mapped I/O: %t\n", config.UseMemoryMap)
	fmt.Printf("     - Async I/O: %t\n", config.AsyncIO)
	fmt.Printf("     - SIMD operations: %t\n", config.SIMDEnabled)
	fmt.Printf("     - Vectorized operations: %t\n", config.VectorizedOps)

	// Initialize storage
	filepath := "/tmp/vittoria_optimizer_demo.dat"
	size := int64(10 * 1024 * 1024) // 10MB
	mockStorage := &MockStorageEngine{
		pages: make(map[uint32]*storage.Page),
	}

	err := optimizer.InitializeStorage(filepath, size, mockStorage)
	if err != nil {
		log.Printf("Failed to initialize optimizer storage: %v", err)
		return
	}
	defer optimizer.Close()

	// Generate test data
	dimensions := 384
	numVectors := 1000

	query := make([]float32, dimensions)
	vectors := make([][]float32, numVectors)

	for i := 0; i < dimensions; i++ {
		query[i] = rand.Float32()
	}

	for i := 0; i < numVectors; i++ {
		vectors[i] = make([]float32, dimensions)
		for j := 0; j < dimensions; j++ {
			vectors[i][j] = rand.Float32()
		}
	}

	// Demonstrate optimized vector similarity
	fmt.Printf("   • Testing optimized vector similarity:\n")

	start := time.Now()
	similarities := optimizer.OptimizedVectorSimilarity(query, vectors, core.DistanceMetricCosine)
	similarityTime := time.Since(start)

	fmt.Printf("     - Calculated %d similarities in %v\n", len(similarities), similarityTime)
	fmt.Printf("     - Average similarity: %.4f\n", average(similarities))

	// Demonstrate optimized batch normalization
	fmt.Printf("   • Testing optimized batch normalization:\n")

	testVectors := make([][]float32, 100)
	for i := range testVectors {
		testVectors[i] = make([]float32, dimensions)
		for j := range testVectors[i] {
			testVectors[i][j] = rand.Float32() * 10
		}
	}

	start = time.Now()
	optimizer.OptimizedBatchNormalize(testVectors)
	normalizeTime := time.Since(start)

	fmt.Printf("     - Normalized %d vectors in %v\n", len(testVectors), normalizeTime)

	// Sync all operations
	start = time.Now()
	err = optimizer.Sync()
	syncTime := time.Since(start)

	if err != nil {
		fmt.Printf("   • Sync failed: %v\n", err)
	} else {
		fmt.Printf("   • Synced all operations in %v\n", syncTime)
	}

	// Get optimizer statistics
	stats := optimizer.GetStats()
	fmt.Printf("   • I/O Optimizer Statistics:\n")
	fmt.Printf("     - Total operations: %d\n", stats.TotalOperations)
	fmt.Printf("     - Average latency: %v\n", stats.AverageLatency)

	if stats.AsyncIOStats != nil {
		fmt.Printf("     - Async I/O throughput: %.2f ops/sec\n", stats.AsyncIOStats.Throughput)
	}
}

func runPerformanceBenchmarks() {
	fmt.Printf("   • Running comprehensive performance benchmarks:\n")

	// SIMD benchmarks
	config := core.DefaultSIMDConfig()
	simdOps := core.NewSIMDVectorOps(config)

	dimensions := []int{128, 384, 768, 1536}
	vectorCounts := []int{100, 1000, 10000}

	for _, dim := range dimensions {
		for _, count := range vectorCounts {
			fmt.Printf("     - Benchmarking %d vectors of %d dimensions:\n", count, dim)

			results := simdOps.BenchmarkSIMDOperations(dim, count)

			fmt.Printf("       • Scalar: %v\n", time.Duration(results.ScalarTime))
			fmt.Printf("       • SIMD: %v (%.2fx speedup)\n",
				time.Duration(results.VectorizedTime), results.VectorizedSpeedup)
			fmt.Printf("       • Parallel: %v (%.2fx speedup)\n",
				time.Duration(results.ParallelTime), results.ParallelSpeedup)
		}
	}

	// I/O optimizer benchmarks
	fmt.Printf("     - I/O Optimizer Benchmarks:\n")

	optimizer := core.NewIOOptimizer(nil)
	defer optimizer.Close()

	benchmark := optimizer.BenchmarkOptimizations(384, 1000)

	fmt.Printf("       • Dimensions: %d, Vectors: %d\n",
		benchmark.Dimensions, benchmark.NumVectors)
	fmt.Printf("       • Timestamp: %v\n", benchmark.Timestamp.Format(time.RFC3339))

	if benchmark.SIMDResults != nil {
		fmt.Printf("       • SIMD vectorized speedup: %.2fx\n",
			benchmark.SIMDResults.VectorizedSpeedup)
		fmt.Printf("       • SIMD parallel speedup: %.2fx\n",
			benchmark.SIMDResults.ParallelSpeedup)
	}

	if benchmark.IOResults != nil {
		fmt.Printf("       • Read latency: %v\n", benchmark.IOResults.ReadLatency)
		fmt.Printf("       • Write latency: %v\n", benchmark.IOResults.WriteLatency)
		fmt.Printf("       • Read throughput: %.2f ops/sec\n", benchmark.IOResults.ReadThroughput)
		fmt.Printf("       • Write throughput: %.2f ops/sec\n", benchmark.IOResults.WriteThroughput)
	}

	// System information
	fmt.Printf("   • System Information:\n")
	fmt.Printf("     - CPU cores: %d\n", runtime.NumCPU())
	fmt.Printf("     - GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("     - Allocated memory: %d KB\n", m.Alloc/1024)
	fmt.Printf("     - System memory: %d KB\n", m.Sys/1024)
}

// Helper functions and mock implementations

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func average(values []float32) float32 {
	if len(values) == 0 {
		return 0
	}

	var sum float32
	for _, v := range values {
		sum += v
	}
	return sum / float32(len(values))
}

func getTotalOperations(stats *storage.AsyncIOStats) int64 {
	var total int64
	for _, count := range stats.OperationsTotal {
		total += count
	}
	return total
}

// MockStorageEngine for demonstration purposes
type MockStorageEngine struct {
	pages map[uint32]*storage.Page
	mu    sync.RWMutex
}

func (m *MockStorageEngine) Open(filepath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pages = make(map[uint32]*storage.Page)
	return nil
}

func (m *MockStorageEngine) Close() error {
	return nil
}

func (m *MockStorageEngine) Sync() error {
	// Simulate sync delay
	time.Sleep(1 * time.Millisecond)
	return nil
}

func (m *MockStorageEngine) ReadPage(pageID uint32) (*storage.Page, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simulate read delay
	time.Sleep(100 * time.Microsecond)

	if page, exists := m.pages[pageID]; exists {
		return page, nil
	}

	// Return empty page if not found
	return &storage.Page{
		ID:   pageID,
		Type: storage.PageTypeVectorLeaf,
		Size: 1024,
		Data: make([]byte, 1024),
	}, nil
}

func (m *MockStorageEngine) WritePage(page *storage.Page) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simulate write delay
	time.Sleep(200 * time.Microsecond)

	// Copy page data
	pageCopy := &storage.Page{
		ID:   page.ID,
		Type: page.Type,
		Size: page.Size,
		Data: make([]byte, len(page.Data)),
	}
	copy(pageCopy.Data, page.Data)

	m.pages[page.ID] = pageCopy
	return nil
}

func (m *MockStorageEngine) AllocatePage() (uint32, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find next available page ID
	pageID := uint32(len(m.pages) + 1)
	return pageID, nil
}

func (m *MockStorageEngine) FreePage(pageID uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.pages, pageID)
	return nil
}

func (m *MockStorageEngine) BeginTx() (storage.Transaction, error) {
	return &MockTransaction{storage: m}, nil
}

func (m *MockStorageEngine) Compact() error {
	return nil
}

func (m *MockStorageEngine) Stats() *storage.StorageStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &storage.StorageStats{
		TotalPages:   uint64(len(m.pages)),
		UsedPages:    uint64(len(m.pages)),
		FreePages:    0,
		PageSize:     storage.PageSize,
		FileSize:     int64(len(m.pages)) * storage.PageSize,
		CacheHitRate: 1.0,
		WALSize:      0,
	}
}

// MockTransaction for demonstration
type MockTransaction struct {
	storage *MockStorageEngine
}

func (t *MockTransaction) ReadPage(pageID uint32) (*storage.Page, error) {
	return t.storage.ReadPage(pageID)
}

func (t *MockTransaction) WritePage(page *storage.Page) error {
	return t.storage.WritePage(page)
}

func (t *MockTransaction) Commit() error {
	return nil
}

func (t *MockTransaction) Rollback() error {
	return nil
}
