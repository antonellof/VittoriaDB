package index

import (
	"fmt"
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

// RunBenchmark runs a benchmark with the given configuration
func RunBenchmark(config *BenchmarkConfig) (*BenchmarkResult, error) {
	// This would implement a comprehensive benchmark
	// For now, return a placeholder
	return &BenchmarkResult{
		Config:          config,
		BuildTimeMS:     1000,
		MemoryUsageMB:   100.0,
		AvgSearchTimeMS: 1.0,
		P99SearchTimeMS: 5.0,
		RecallAt10:      0.95,
		QPS:             1000.0,
	}, nil
}
