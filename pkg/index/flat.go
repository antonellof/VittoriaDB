package index

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"
)

// FlatIndex implements a brute-force flat index
type FlatIndex struct {
	vectors    []*IndexVector
	dimensions int
	metric     DistanceMetric
	calculator DistanceCalculator
	mu         sync.RWMutex
	config     *FlatConfig
	stats      *IndexStats
}

// NewFlatIndex creates a new flat index
func NewFlatIndex(dimensions int, metric DistanceMetric, config *FlatConfig) *FlatIndex {
	if config == nil {
		config = DefaultFlatConfig()
	}

	return &FlatIndex{
		vectors:    make([]*IndexVector, 0),
		dimensions: dimensions,
		metric:     metric,
		calculator: NewDistanceCalculator(metric),
		config:     config,
		stats: &IndexStats{
			IndexType:   IndexTypeFlat,
			Dimensions:  dimensions,
			VectorCount: 0,
		},
	}
}

// Build builds the index from a set of vectors
func (idx *FlatIndex) Build(vectors []*IndexVector) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	startTime := time.Now()

	// Validate vectors
	for i, vector := range vectors {
		if len(vector.Vector) != idx.dimensions {
			return fmt.Errorf("vector %d has wrong dimensions: expected %d, got %d",
				i, idx.dimensions, len(vector.Vector))
		}
	}

	// Copy vectors
	idx.vectors = make([]*IndexVector, len(vectors))
	for i, vector := range vectors {
		idx.vectors[i] = &IndexVector{
			ID:     vector.ID,
			Vector: make([]float32, len(vector.Vector)),
		}
		copy(idx.vectors[i].Vector, vector.Vector)
	}

	// Update stats
	idx.stats.VectorCount = len(idx.vectors)
	idx.stats.BuildTime = time.Since(startTime).Milliseconds()

	return nil
}

// Load loads the index from a reader
func (idx *FlatIndex) Load(r io.Reader) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	decoder := json.NewDecoder(r)

	var data struct {
		Vectors    []*IndexVector `json:"vectors"`
		Dimensions int            `json:"dimensions"`
		Metric     DistanceMetric `json:"metric"`
		Stats      *IndexStats    `json:"stats"`
	}

	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode flat index: %w", err)
	}

	// Validate dimensions and metric
	if data.Dimensions != idx.dimensions {
		return fmt.Errorf("dimension mismatch: expected %d, got %d",
			idx.dimensions, data.Dimensions)
	}
	if data.Metric != idx.metric {
		return fmt.Errorf("metric mismatch: expected %s, got %s",
			idx.metric.String(), data.Metric.String())
	}

	idx.vectors = data.Vectors
	idx.stats = data.Stats

	return nil
}

// Save saves the index to a writer
func (idx *FlatIndex) Save(w io.Writer) error {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	encoder := json.NewEncoder(w)

	data := struct {
		Vectors    []*IndexVector `json:"vectors"`
		Dimensions int            `json:"dimensions"`
		Metric     DistanceMetric `json:"metric"`
		Stats      *IndexStats    `json:"stats"`
	}{
		Vectors:    idx.vectors,
		Dimensions: idx.dimensions,
		Metric:     idx.metric,
		Stats:      idx.stats,
	}

	return encoder.Encode(data)
}

// Add adds a vector to the index
func (idx *FlatIndex) Add(ctx context.Context, vector *IndexVector) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Validate vector
	if len(vector.Vector) != idx.dimensions {
		return fmt.Errorf("vector has wrong dimensions: expected %d, got %d",
			idx.dimensions, len(vector.Vector))
	}

	// Check for duplicate ID
	for _, existing := range idx.vectors {
		if existing.ID == vector.ID {
			return fmt.Errorf("vector with ID %s already exists", vector.ID)
		}
	}

	// Add vector
	newVector := &IndexVector{
		ID:     vector.ID,
		Vector: make([]float32, len(vector.Vector)),
	}
	copy(newVector.Vector, vector.Vector)

	idx.vectors = append(idx.vectors, newVector)
	idx.stats.VectorCount = len(idx.vectors)

	return nil
}

// Delete removes a vector from the index
func (idx *FlatIndex) Delete(ctx context.Context, id string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Find and remove vector
	for i, vector := range idx.vectors {
		if vector.ID == id {
			// Remove by swapping with last element
			idx.vectors[i] = idx.vectors[len(idx.vectors)-1]
			idx.vectors = idx.vectors[:len(idx.vectors)-1]
			idx.stats.VectorCount = len(idx.vectors)
			return nil
		}
	}

	return fmt.Errorf("vector with ID %s not found", id)
}

// Search performs k-nearest neighbor search
func (idx *FlatIndex) Search(ctx context.Context, query []float32, k int, params *SearchParams) ([]*Candidate, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	startTime := time.Now()

	// Validate query
	if len(query) != idx.dimensions {
		return nil, fmt.Errorf("query vector has wrong dimensions: expected %d, got %d",
			idx.dimensions, len(query))
	}

	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	if k > len(idx.vectors) {
		k = len(idx.vectors)
	}

	// Calculate distances for all vectors
	candidates := make([]*Candidate, 0, len(idx.vectors))

	for _, vector := range idx.vectors {
		distance := idx.calculator.Calculate(query, vector.Vector)
		candidates = append(candidates, &Candidate{
			ID:    vector.ID,
			Score: distance,
		})
	}

	// Sort by distance (ascending for distance, descending for similarity)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score < candidates[j].Score
	})

	// Return top-k results
	if k > len(candidates) {
		k = len(candidates)
	}

	results := candidates[:k]

	// Update search latency stats (simplified)
	latency := time.Since(startTime).Seconds() * 1000 // Convert to milliseconds
	idx.stats.SearchLatencyP50 = latency              // Simplified - would need proper percentile calculation
	idx.stats.SearchLatencyP99 = latency

	return results, nil
}

// Size returns the number of vectors in the index
func (idx *FlatIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.vectors)
}

// Dimensions returns the vector dimensions
func (idx *FlatIndex) Dimensions() int {
	return idx.dimensions
}

// Type returns the index type
func (idx *FlatIndex) Type() IndexType {
	return IndexTypeFlat
}

// Optimize optimizes the index (no-op for flat index)
func (idx *FlatIndex) Optimize() error {
	// Flat index doesn't need optimization
	return nil
}

// Stats returns index statistics
func (idx *FlatIndex) Stats() *IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Calculate memory usage
	vectorMemory := int64(len(idx.vectors)) * int64(idx.dimensions) * 4 // 4 bytes per float32
	idMemory := int64(0)
	for _, vector := range idx.vectors {
		idMemory += int64(len(vector.ID))
	}

	stats := *idx.stats
	stats.MemoryUsage = vectorMemory + idMemory
	stats.VectorCount = len(idx.vectors)

	return &stats
}
