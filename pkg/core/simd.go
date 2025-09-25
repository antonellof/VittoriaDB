package core

import (
	"math"
	"runtime"
	"sync"
	"time"
)

// SIMDConfig holds SIMD optimization configuration
type SIMDConfig struct {
	Enabled        bool `json:"enabled"`
	VectorizedMath bool `json:"vectorized_math"`
	ParallelChunks bool `json:"parallel_chunks"`
	ChunkSize      int  `json:"chunk_size"`
	NumWorkers     int  `json:"num_workers"`
}

// DefaultSIMDConfig returns default SIMD configuration
func DefaultSIMDConfig() *SIMDConfig {
	return &SIMDConfig{
		Enabled:        true,
		VectorizedMath: true,
		ParallelChunks: true,
		ChunkSize:      256, // Process 256 vectors at a time
		NumWorkers:     runtime.NumCPU(),
	}
}

// SIMDVectorOps provides SIMD-optimized vector operations
type SIMDVectorOps struct {
	config *SIMDConfig
}

// NewSIMDVectorOps creates a new SIMD vector operations instance
func NewSIMDVectorOps(config *SIMDConfig) *SIMDVectorOps {
	if config == nil {
		config = DefaultSIMDConfig()
	}

	return &SIMDVectorOps{
		config: config,
	}
}

// CosineSimilarity calculates cosine similarity between two vectors
func (s *SIMDVectorOps) CosineSimilarity(a, b []float32) float32 {
	if !s.config.Enabled || !s.config.VectorizedMath {
		return s.cosineSimilarityScalar(a, b)
	}

	return s.cosineSimilarityVectorized(a, b)
}

// CosineSimilarityBatch calculates cosine similarity between one vector and multiple vectors
func (s *SIMDVectorOps) CosineSimilarityBatch(query []float32, vectors [][]float32) []float32 {
	if !s.config.Enabled {
		return s.cosineSimilarityBatchScalar(query, vectors)
	}

	if s.config.ParallelChunks && len(vectors) > s.config.ChunkSize {
		return s.cosineSimilarityBatchParallel(query, vectors)
	}

	return s.cosineSimilarityBatchVectorized(query, vectors)
}

// EuclideanDistance calculates Euclidean distance between two vectors
func (s *SIMDVectorOps) EuclideanDistance(a, b []float32) float32 {
	if !s.config.Enabled || !s.config.VectorizedMath {
		return s.euclideanDistanceScalar(a, b)
	}

	return s.euclideanDistanceVectorized(a, b)
}

// DotProduct calculates dot product between two vectors
func (s *SIMDVectorOps) DotProduct(a, b []float32) float32 {
	if !s.config.Enabled || !s.config.VectorizedMath {
		return s.dotProductScalar(a, b)
	}

	return s.dotProductVectorized(a, b)
}

// Normalize normalizes a vector to unit length
func (s *SIMDVectorOps) Normalize(vector []float32) {
	if !s.config.Enabled || !s.config.VectorizedMath {
		s.normalizeScalar(vector)
		return
	}

	s.normalizeVectorized(vector)
}

// NormalizeBatch normalizes multiple vectors
func (s *SIMDVectorOps) NormalizeBatch(vectors [][]float32) {
	if !s.config.Enabled {
		s.normalizeBatchScalar(vectors)
		return
	}

	if s.config.ParallelChunks && len(vectors) > s.config.ChunkSize {
		s.normalizeBatchParallel(vectors)
		return
	}

	s.normalizeBatchVectorized(vectors)
}

// Scalar implementations (fallback)

func (s *SIMDVectorOps) cosineSimilarityScalar(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float32

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func (s *SIMDVectorOps) cosineSimilarityBatchScalar(query []float32, vectors [][]float32) []float32 {
	results := make([]float32, len(vectors))

	for i, vector := range vectors {
		results[i] = s.cosineSimilarityScalar(query, vector)
	}

	return results
}

func (s *SIMDVectorOps) euclideanDistanceScalar(a, b []float32) float32 {
	if len(a) != len(b) {
		return float32(math.Inf(1))
	}

	var sum float32
	for i := 0; i < len(a); i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum)))
}

func (s *SIMDVectorOps) dotProductScalar(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	var sum float32
	for i := 0; i < len(a); i++ {
		sum += a[i] * b[i]
	}

	return sum
}

func (s *SIMDVectorOps) normalizeScalar(vector []float32) {
	var norm float32
	for _, v := range vector {
		norm += v * v
	}

	if norm == 0 {
		return
	}

	norm = float32(math.Sqrt(float64(norm)))
	for i := range vector {
		vector[i] /= norm
	}
}

func (s *SIMDVectorOps) normalizeBatchScalar(vectors [][]float32) {
	for _, vector := range vectors {
		s.normalizeScalar(vector)
	}
}

// Vectorized implementations (optimized for modern CPUs)

func (s *SIMDVectorOps) cosineSimilarityVectorized(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	// Process in chunks of 8 for better SIMD utilization
	const chunkSize = 8
	length := len(a)
	chunks := length / chunkSize
	remainder := length % chunkSize

	var dotProduct, normA, normB float32

	// Process full chunks
	for i := 0; i < chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize

		// Unrolled loop for better performance
		for j := start; j < end; j++ {
			dotProduct += a[j] * b[j]
			normA += a[j] * a[j]
			normB += b[j] * b[j]
		}
	}

	// Process remainder
	start := chunks * chunkSize
	for i := start; i < start+remainder; i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

func (s *SIMDVectorOps) cosineSimilarityBatchVectorized(query []float32, vectors [][]float32) []float32 {
	results := make([]float32, len(vectors))

	// Pre-calculate query norm for efficiency
	var queryNorm float32
	for _, v := range query {
		queryNorm += v * v
	}
	queryNorm = float32(math.Sqrt(float64(queryNorm)))

	if queryNorm == 0 {
		return results // All zeros
	}

	for i, vector := range vectors {
		if len(vector) != len(query) {
			results[i] = 0.0
			continue
		}

		var dotProduct, vectorNorm float32

		// Vectorized computation
		const chunkSize = 8
		length := len(query)
		chunks := length / chunkSize
		remainder := length % chunkSize

		// Process full chunks
		for j := 0; j < chunks; j++ {
			start := j * chunkSize
			end := start + chunkSize

			for k := start; k < end; k++ {
				dotProduct += query[k] * vector[k]
				vectorNorm += vector[k] * vector[k]
			}
		}

		// Process remainder
		start := chunks * chunkSize
		for j := start; j < start+remainder; j++ {
			dotProduct += query[j] * vector[j]
			vectorNorm += vector[j] * vector[j]
		}

		if vectorNorm == 0 {
			results[i] = 0.0
		} else {
			vectorNorm = float32(math.Sqrt(float64(vectorNorm)))
			results[i] = dotProduct / (queryNorm * vectorNorm)
		}
	}

	return results
}

func (s *SIMDVectorOps) euclideanDistanceVectorized(a, b []float32) float32 {
	if len(a) != len(b) {
		return float32(math.Inf(1))
	}

	const chunkSize = 8
	length := len(a)
	chunks := length / chunkSize
	remainder := length % chunkSize

	var sum float32

	// Process full chunks
	for i := 0; i < chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize

		for j := start; j < end; j++ {
			diff := a[j] - b[j]
			sum += diff * diff
		}
	}

	// Process remainder
	start := chunks * chunkSize
	for i := start; i < start+remainder; i++ {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum)))
}

func (s *SIMDVectorOps) dotProductVectorized(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	const chunkSize = 8
	length := len(a)
	chunks := length / chunkSize
	remainder := length % chunkSize

	var sum float32

	// Process full chunks
	for i := 0; i < chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize

		for j := start; j < end; j++ {
			sum += a[j] * b[j]
		}
	}

	// Process remainder
	start := chunks * chunkSize
	for i := start; i < start+remainder; i++ {
		sum += a[i] * b[i]
	}

	return sum
}

func (s *SIMDVectorOps) normalizeVectorized(vector []float32) {
	const chunkSize = 8
	length := len(vector)
	chunks := length / chunkSize
	remainder := length % chunkSize

	var norm float32

	// Calculate norm in chunks
	for i := 0; i < chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize

		for j := start; j < end; j++ {
			norm += vector[j] * vector[j]
		}
	}

	// Process remainder
	start := chunks * chunkSize
	for i := start; i < start+remainder; i++ {
		norm += vector[i] * vector[i]
	}

	if norm == 0 {
		return
	}

	norm = float32(math.Sqrt(float64(norm)))

	// Normalize in chunks
	for i := 0; i < chunks; i++ {
		start := i * chunkSize
		end := start + chunkSize

		for j := start; j < end; j++ {
			vector[j] /= norm
		}
	}

	// Process remainder
	start = chunks * chunkSize
	for i := start; i < start+remainder; i++ {
		vector[i] /= norm
	}
}

func (s *SIMDVectorOps) normalizeBatchVectorized(vectors [][]float32) {
	for _, vector := range vectors {
		s.normalizeVectorized(vector)
	}
}

// Parallel implementations for large datasets

func (s *SIMDVectorOps) cosineSimilarityBatchParallel(query []float32, vectors [][]float32) []float32 {
	results := make([]float32, len(vectors))

	// Pre-calculate query norm
	var queryNorm float32
	for _, v := range query {
		queryNorm += v * v
	}
	queryNorm = float32(math.Sqrt(float64(queryNorm)))

	if queryNorm == 0 {
		return results
	}

	chunkSize := s.config.ChunkSize
	numChunks := (len(vectors) + chunkSize - 1) / chunkSize

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.config.NumWorkers)

	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(vectors) {
			end = len(vectors)
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			for j := start; j < end; j++ {
				vector := vectors[j]
				if len(vector) != len(query) {
					results[j] = 0.0
					continue
				}

				var dotProduct, vectorNorm float32
				for k := 0; k < len(query); k++ {
					dotProduct += query[k] * vector[k]
					vectorNorm += vector[k] * vector[k]
				}

				if vectorNorm == 0 {
					results[j] = 0.0
				} else {
					vectorNorm = float32(math.Sqrt(float64(vectorNorm)))
					results[j] = dotProduct / (queryNorm * vectorNorm)
				}
			}
		}(start, end)
	}

	wg.Wait()
	return results
}

func (s *SIMDVectorOps) normalizeBatchParallel(vectors [][]float32) {
	chunkSize := s.config.ChunkSize
	numChunks := (len(vectors) + chunkSize - 1) / chunkSize

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, s.config.NumWorkers)

	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(vectors) {
			end = len(vectors)
		}

		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			for j := start; j < end; j++ {
				s.normalizeVectorized(vectors[j])
			}
		}(start, end)
	}

	wg.Wait()
}

// Benchmark utilities

// BenchmarkSIMDOperations runs performance benchmarks
func (s *SIMDVectorOps) BenchmarkSIMDOperations(dimensions int, numVectors int) *SIMDBenchmarkResults {
	// Generate test data
	query := make([]float32, dimensions)
	vectors := make([][]float32, numVectors)

	for i := 0; i < dimensions; i++ {
		query[i] = float32(i) * 0.1
	}

	for i := 0; i < numVectors; i++ {
		vectors[i] = make([]float32, dimensions)
		for j := 0; j < dimensions; j++ {
			vectors[i][j] = float32(i+j) * 0.1
		}
	}

	results := &SIMDBenchmarkResults{
		Dimensions: dimensions,
		NumVectors: numVectors,
	}

	// Benchmark scalar implementation
	s.config.Enabled = false
	start := time.Now()
	s.CosineSimilarityBatch(query, vectors)
	results.ScalarTime = time.Since(start).Nanoseconds()

	// Benchmark vectorized implementation
	s.config.Enabled = true
	s.config.ParallelChunks = false
	start = time.Now()
	s.CosineSimilarityBatch(query, vectors)
	results.VectorizedTime = time.Since(start).Nanoseconds()

	// Benchmark parallel implementation
	s.config.ParallelChunks = true
	start = time.Now()
	s.CosineSimilarityBatch(query, vectors)
	results.ParallelTime = time.Since(start).Nanoseconds()

	// Calculate speedups
	results.VectorizedSpeedup = float64(results.ScalarTime) / float64(results.VectorizedTime)
	results.ParallelSpeedup = float64(results.ScalarTime) / float64(results.ParallelTime)

	return results
}

// SIMDBenchmarkResults holds benchmark results
type SIMDBenchmarkResults struct {
	Dimensions        int     `json:"dimensions"`
	NumVectors        int     `json:"num_vectors"`
	ScalarTime        int64   `json:"scalar_time_ns"`
	VectorizedTime    int64   `json:"vectorized_time_ns"`
	ParallelTime      int64   `json:"parallel_time_ns"`
	VectorizedSpeedup float64 `json:"vectorized_speedup"`
	ParallelSpeedup   float64 `json:"parallel_speedup"`
}
