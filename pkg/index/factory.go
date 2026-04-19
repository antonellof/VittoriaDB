package index

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// CreateIndex creates an index of the specified type
func CreateIndex(indexType IndexType, dimensions int, metric DistanceMetric, config map[string]interface{}) (Index, error) {
	switch indexType {
	case IndexTypeFlat:
		flatConfig := DefaultFlatConfig()
		if config != nil {
			if batchSize, ok := config["batch_size"].(int); ok {
				flatConfig.BatchSize = batchSize
			}
		}
		return NewFlatIndex(dimensions, metric, flatConfig), nil

	case IndexTypeHNSW:
		hnswConfig := DefaultHNSWConfig()
		if config != nil {
			if m, ok := config["m"].(int); ok {
				hnswConfig.M = m
				hnswConfig.MaxM = m
			}
			if maxM0, ok := config["max_m0"].(int); ok {
				hnswConfig.MaxM0 = maxM0
			}
			if efConstruction, ok := config["ef_construction"].(int); ok {
				hnswConfig.EfConstruction = efConstruction
			}
			if efSearch, ok := config["ef_search"].(int); ok {
				hnswConfig.EfSearch = efSearch
			}
			if ml, ok := config["ml"].(float64); ok {
				hnswConfig.ML = ml
			}
			if seed, ok := config["seed"].(int64); ok {
				hnswConfig.Seed = seed
			}
		}
		return NewHNSWIndex(dimensions, metric, hnswConfig), nil

	case IndexTypeIVF:
		return nil, fmt.Errorf("IVF index not implemented yet")

	default:
		return nil, fmt.Errorf("unknown index type: %s", indexType.String())
	}
}

// ParseIndexType parses an index type string
func ParseIndexType(s string) (IndexType, error) {
	switch s {
	case "flat":
		return IndexTypeFlat, nil
	case "hnsw":
		return IndexTypeHNSW, nil
	case "ivf":
		return IndexTypeIVF, nil
	default:
		return IndexTypeFlat, fmt.Errorf("unknown index type: %s", s)
	}
}

// ParseDistanceMetric parses a distance metric string
func ParseDistanceMetric(s string) (DistanceMetric, error) {
	switch s {
	case "cosine":
		return DistanceMetricCosine, nil
	case "euclidean":
		return DistanceMetricEuclidean, nil
	case "dot_product":
		return DistanceMetricDotProduct, nil
	case "manhattan":
		return DistanceMetricManhattan, nil
	default:
		return DistanceMetricCosine, fmt.Errorf("unknown distance metric: %s", s)
	}
}

// RecommendedConfig returns recommended configuration for different use cases
func RecommendedConfig(useCase string, dimensions int, expectedSize int) map[string]interface{} {
	config := make(map[string]interface{})

	switch useCase {
	case "small":
		// Small datasets (< 10k vectors) - use flat index
		config["index_type"] = "flat"
		config["batch_size"] = 1000

	case "medium":
		// Medium datasets (10k - 100k vectors) - use HNSW with moderate parameters
		config["index_type"] = "hnsw"
		config["m"] = 16
		config["ef_construction"] = 200
		config["ef_search"] = 50

	case "large":
		// Large datasets (> 100k vectors) - use HNSW with higher parameters
		config["index_type"] = "hnsw"
		config["m"] = 32
		config["ef_construction"] = 400
		config["ef_search"] = 100

	case "high_precision":
		// High precision requirements - use HNSW with high parameters
		config["index_type"] = "hnsw"
		config["m"] = 48
		config["ef_construction"] = 500
		config["ef_search"] = 200

	case "fast_build":
		// Fast build time - use HNSW with lower parameters
		config["index_type"] = "hnsw"
		config["m"] = 8
		config["ef_construction"] = 100
		config["ef_search"] = 32

	default:
		// Default configuration
		if expectedSize < 10000 {
			config["index_type"] = "flat"
			config["batch_size"] = 1000
		} else {
			config["index_type"] = "hnsw"
			config["m"] = 16
			config["ef_construction"] = 200
			config["ef_search"] = 50
		}
	}

	return config
}

// EstimateMemoryUsage estimates memory usage for different index configurations
func EstimateMemoryUsage(indexType IndexType, dimensions int, vectorCount int, config map[string]interface{}) int64 {
	vectorMemory := int64(vectorCount) * int64(dimensions) * 4 // 4 bytes per float32

	switch indexType {
	case IndexTypeFlat:
		// Flat index has minimal overhead
		return vectorMemory + int64(vectorCount)*64 // 64 bytes overhead per vector

	case IndexTypeHNSW:
		// HNSW has connection overhead
		m := 16
		if config != nil {
			if mVal, ok := config["m"].(int); ok {
				m = mVal
			}
		}

		// Estimate average connections per vector
		avgConnections := float64(m) * 1.5                                   // Rough estimate
		connectionMemory := int64(float64(vectorCount) * avgConnections * 8) // 8 bytes per connection

		return vectorMemory + connectionMemory + int64(vectorCount)*128 // 128 bytes overhead per node

	case IndexTypeIVF:
		// IVF not implemented yet
		return vectorMemory

	default:
		return vectorMemory
	}
}

// BenchmarkConfig represents benchmark configuration
type BenchmarkConfig struct {
	IndexType   IndexType              `json:"index_type"`
	Dimensions  int                    `json:"dimensions"`
	VectorCount int                    `json:"vector_count"`
	QueryCount  int                    `json:"query_count"`
	K           int                    `json:"k"`
	Config      map[string]interface{} `json:"config"`
}

// BenchmarkResult represents benchmark results
type BenchmarkResult struct {
	Config          *BenchmarkConfig `json:"config"`
	BuildTimeMS     int64            `json:"build_time_ms"`
	MemoryUsageMB   float64          `json:"memory_usage_mb"`
	AvgSearchTimeMS float64          `json:"avg_search_time_ms"`
	P99SearchTimeMS float64          `json:"p99_search_time_ms"`
	RecallAt10      float64          `json:"recall_at_10"`
	QPS             float64          `json:"qps"`
}

// RunBenchmark builds an index with synthetic random unit vectors, runs search
// queries, and returns timing statistics. Recall is not computed (left at 0).
func RunBenchmark(config *BenchmarkConfig) (*BenchmarkResult, error) {
	if config == nil {
		return nil, fmt.Errorf("benchmark config is nil")
	}
	if config.VectorCount <= 0 || config.Dimensions <= 0 {
		return nil, fmt.Errorf("vector_count and dimensions must be positive")
	}

	idx, err := CreateIndex(config.IndexType, config.Dimensions, DistanceMetricCosine, config.Config)
	if err != nil {
		return nil, err
	}

	seed := int64(42)
	if config.Config != nil {
		if s, ok := config.Config["seed"].(int64); ok {
			seed = s
		}
	}
	r := rand.New(rand.NewSource(seed))

	vectors := make([]*IndexVector, config.VectorCount)
	for i := 0; i < config.VectorCount; i++ {
		v := make([]float32, config.Dimensions)
		var norm float32
		for j := 0; j < config.Dimensions; j++ {
			x := r.Float32()*2 - 1
			v[j] = x
			norm += x * x
		}
		if norm > 0 {
			inv := float32(1.0 / math.Sqrt(float64(norm)))
			for j := range v {
				v[j] *= inv
			}
		}
		vectors[i] = &IndexVector{ID: fmt.Sprintf("bench-%d", i), Vector: v}
	}

	buildStart := time.Now()
	if err := idx.Build(vectors); err != nil {
		return nil, err
	}
	buildMs := time.Since(buildStart).Milliseconds()

	q := config.QueryCount
	if q <= 0 {
		q = 100
	}
	k := config.K
	if k <= 0 {
		k = 10
	}

	ctx := context.Background()
	latencies := make([]time.Duration, 0, q)
	var searchTotal time.Duration
	queryRand := rand.New(rand.NewSource(seed + 1))

	for i := 0; i < q; i++ {
		qv := make([]float32, config.Dimensions)
		var norm float32
		for j := 0; j < config.Dimensions; j++ {
			x := queryRand.Float32()*2 - 1
			qv[j] = x
			norm += x * x
		}
		if norm > 0 {
			inv := float32(1.0 / math.Sqrt(float64(norm)))
			for j := range qv {
				qv[j] *= inv
			}
		}
		t0 := time.Now()
		if _, err := idx.Search(ctx, qv, k, &SearchParams{}); err != nil {
			return nil, err
		}
		d := time.Since(t0)
		latencies = append(latencies, d)
		searchTotal += d
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p99Idx := (len(latencies) * 99 / 100)
	if p99Idx >= len(latencies) {
		p99Idx = len(latencies) - 1
	}
	p99Ms := float64(latencies[p99Idx]) / float64(time.Millisecond)
	avgMs := float64(searchTotal) / float64(q) / float64(time.Millisecond)
	qps := float64(q) / searchTotal.Seconds()

	memMB := 0.0
	if st := idx.Stats(); st != nil && st.MemoryUsage > 0 {
		memMB = float64(st.MemoryUsage) / (1024 * 1024)
	}

	return &BenchmarkResult{
		Config:          config,
		BuildTimeMS:     buildMs,
		MemoryUsageMB:   memMB,
		AvgSearchTimeMS: avgMs,
		P99SearchTimeMS: p99Ms,
		RecallAt10:      0,
		QPS:             qps,
	}, nil
}
