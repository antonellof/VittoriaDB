package index

import (
	"context"
	"io"
)

// Index provides vector similarity search
type Index interface {
	// Lifecycle
	Build(vectors []*IndexVector) error
	Load(r io.Reader) error
	Save(w io.Writer) error

	// Operations
	Add(ctx context.Context, vector *IndexVector) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query []float32, k int, params *SearchParams) ([]*Candidate, error)

	// Metadata
	Size() int
	Dimensions() int
	Type() IndexType

	// Maintenance
	Optimize() error
	Stats() *IndexStats
}

// IndexType represents the type of vector index
type IndexType int

const (
	IndexTypeFlat IndexType = iota
	IndexTypeHNSW
	IndexTypeIVF
)

func (i IndexType) String() string {
	switch i {
	case IndexTypeFlat:
		return "flat"
	case IndexTypeHNSW:
		return "hnsw"
	case IndexTypeIVF:
		return "ivf"
	default:
		return "unknown"
	}
}

// IndexVector represents a vector with ID for indexing
type IndexVector struct {
	ID     string    `json:"id"`
	Vector []float32 `json:"vector"`
}

// Candidate represents a search result candidate
type Candidate struct {
	ID    string  `json:"id"`
	Score float32 `json:"score"`
}

// SearchParams contains search parameters
type SearchParams struct {
	EF          int                    `json:"ef"`           // HNSW search parameter
	NProbes     int                    `json:"n_probes"`     // IVF parameter
	ExactSearch bool                   `json:"exact_search"` // Force exact search
	Params      map[string]interface{} `json:"params"`       // Algorithm-specific
}

// DistanceMetric represents distance calculation methods
type DistanceMetric int

const (
	DistanceMetricCosine DistanceMetric = iota
	DistanceMetricEuclidean
	DistanceMetricDotProduct
	DistanceMetricManhattan
)

func (d DistanceMetric) String() string {
	switch d {
	case DistanceMetricCosine:
		return "cosine"
	case DistanceMetricEuclidean:
		return "euclidean"
	case DistanceMetricDotProduct:
		return "dot_product"
	case DistanceMetricManhattan:
		return "manhattan"
	default:
		return "unknown"
	}
}

// DistanceCalculator provides distance calculations
type DistanceCalculator interface {
	Calculate(a, b []float32) float32
	Name() string
	IsSymmetric() bool
}

// HNSW specific configuration
type HNSWConfig struct {
	M              int     `json:"m"`
	MaxM           int     `json:"max_m"`
	MaxM0          int     `json:"max_m0"`
	ML             float64 `json:"ml"`
	EfConstruction int     `json:"ef_construction"`
	EfSearch       int     `json:"ef_search"`
	Seed           int64   `json:"seed"`
}

// DefaultHNSWConfig returns default HNSW configuration
func DefaultHNSWConfig() *HNSWConfig {
	return &HNSWConfig{
		M:              16,
		MaxM:           16,
		MaxM0:          32,
		ML:             1.0 / 2.303, // 1/ln(2)
		EfConstruction: 200,
		EfSearch:       50,
		Seed:           42,
	}
}

// HNSWIndex extends Index with HNSW-specific methods
type HNSWIndex interface {
	Index
	GetNode(id string) *HNSWNode
	GetConnections(id string, layer int) []string
	SetEfSearch(ef int)
}

// HNSWNode represents a node in the HNSW graph
type HNSWNode struct {
	ID          string           `json:"id"`
	Vector      []float32        `json:"vector"`
	Layer       int              `json:"layer"`
	Connections map[int][]string `json:"connections"`
}

// Flat index configuration
type FlatConfig struct {
	BatchSize int `json:"batch_size"`
}

// DefaultFlatConfig returns default flat index configuration
func DefaultFlatConfig() *FlatConfig {
	return &FlatConfig{
		BatchSize: 1000,
	}
}

// IndexStats represents index statistics
type IndexStats struct {
	IndexType   IndexType `json:"index_type"`
	VectorCount int       `json:"vector_count"`
	Dimensions  int       `json:"dimensions"`
	MemoryUsage int64     `json:"memory_usage"`
	BuildTime   int64     `json:"build_time_ms"`

	// HNSW specific
	MaxLayer  int     `json:"max_layer,omitempty"`
	AvgDegree float64 `json:"avg_degree,omitempty"`

	// Performance metrics
	SearchLatencyP50 float64 `json:"search_latency_p50"`
	SearchLatencyP99 float64 `json:"search_latency_p99"`
	RecallAt10       float64 `json:"recall_at_10"`
}

// Priority queue for search
type PriorityQueue []*QueueItem

type QueueItem struct {
	ID       string
	Distance float32
	Vector   []float32
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Distance < pq[j].Distance
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*QueueItem))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
