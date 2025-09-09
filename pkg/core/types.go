package core

import (
	"context"
	"io"
	"time"
)

// DistanceMetric represents the distance calculation method
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

// Vector represents a vector with metadata
type Vector struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

// CreateCollectionRequest represents a collection creation request
type CreateCollectionRequest struct {
	Name       string                 `json:"name"`
	Dimensions int                    `json:"dimensions"`
	Metric     DistanceMetric         `json:"metric"`
	IndexType  IndexType              `json:"index_type"`
	Config     map[string]interface{} `json:"config"`
}

// SearchRequest represents a vector search request
type SearchRequest struct {
	Vector          []float32              `json:"vector"`
	Limit           int                    `json:"limit"`
	Offset          int                    `json:"offset"`
	Filter          *Filter                `json:"filter"`
	IncludeVector   bool                   `json:"include_vector"`
	IncludeMetadata bool                   `json:"include_metadata"`
	SearchParams    map[string]interface{} `json:"search_params"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Results   []*SearchResult `json:"results"`
	Total     int64           `json:"total"`
	TookMS    int64           `json:"took_ms"`
	RequestID string          `json:"request_id"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Vector   []float32              `json:"vector,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Filter represents metadata filtering
type Filter struct {
	And []Filter `json:"and,omitempty"`
	Or  []Filter `json:"or,omitempty"`
	Not *Filter  `json:"not,omitempty"`

	Field    string      `json:"field,omitempty"`
	Operator FilterOp    `json:"operator,omitempty"`
	Value    interface{} `json:"value,omitempty"`
}

// FilterOp represents filter operations
type FilterOp string

const (
	FilterOpEq       FilterOp = "eq"
	FilterOpNe       FilterOp = "ne"
	FilterOpGt       FilterOp = "gt"
	FilterOpGte      FilterOp = "gte"
	FilterOpLt       FilterOp = "lt"
	FilterOpLte      FilterOp = "lte"
	FilterOpIn       FilterOp = "in"
	FilterOpNotIn    FilterOp = "not_in"
	FilterOpContains FilterOp = "contains"
	FilterOpExists   FilterOp = "exists"
)

// CollectionInfo represents collection metadata
type CollectionInfo struct {
	Name        string         `json:"name"`
	Dimensions  int            `json:"dimensions"`
	Metric      DistanceMetric `json:"metric"`
	IndexType   IndexType      `json:"index_type"`
	VectorCount int64          `json:"vector_count"`
	Created     time.Time      `json:"created"`
	Modified    time.Time      `json:"modified"`
}

// HealthStatus represents system health
type HealthStatus struct {
	Status       string `json:"status"`
	Uptime       int64  `json:"uptime"`
	Collections  int    `json:"collections"`
	TotalVectors int64  `json:"total_vectors"`
	MemoryUsage  int64  `json:"memory_usage"`
	DiskUsage    int64  `json:"disk_usage"`
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	Collections     []*CollectionStats `json:"collections"`
	TotalVectors    int64              `json:"total_vectors"`
	TotalSize       int64              `json:"total_size"`
	IndexSize       int64              `json:"index_size"`
	QueriesTotal    int64              `json:"queries_total"`
	QueriesPerSec   float64            `json:"queries_per_sec"`
	AvgQueryLatency float64            `json:"avg_query_latency"`
}

// CollectionStats represents collection statistics
type CollectionStats struct {
	Name         string    `json:"name"`
	VectorCount  int64     `json:"vector_count"`
	Dimensions   int       `json:"dimensions"`
	IndexType    IndexType `json:"index_type"`
	IndexSize    int64     `json:"index_size"`
	LastModified time.Time `json:"last_modified"`
}

// Config represents database configuration
type Config struct {
	DataDir     string        `yaml:"data_dir"`
	Server      ServerConfig  `yaml:"server"`
	Storage     StorageConfig `yaml:"storage"`
	Index       IndexConfig   `yaml:"index"`
	Performance PerfConfig    `yaml:"performance"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	MaxBodySize  int64         `yaml:"max_body_size"`
	CORS         bool          `yaml:"cors"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	PageSize    int  `yaml:"page_size"`
	CacheSize   int  `yaml:"cache_size"`
	SyncWrites  bool `yaml:"sync_writes"`
	Compression bool `yaml:"compression"`
}

// IndexConfig represents index configuration
type IndexConfig struct {
	DefaultType   IndexType      `yaml:"default_type"`
	DefaultMetric DistanceMetric `yaml:"default_metric"`
	HNSWConfig    HNSWConfig     `yaml:"hnsw"`
	FlatConfig    FlatConfig     `yaml:"flat"`
}

// HNSWConfig represents HNSW index configuration
type HNSWConfig struct {
	M              int     `yaml:"m"`
	MaxM           int     `yaml:"max_m"`
	MaxM0          int     `yaml:"max_m0"`
	ML             float64 `yaml:"ml"`
	EfConstruction int     `yaml:"ef_construction"`
	EfSearch       int     `yaml:"ef_search"`
	Seed           int64   `yaml:"seed"`
}

// FlatConfig represents flat index configuration
type FlatConfig struct {
	BatchSize int `yaml:"batch_size"`
}

// PerfConfig represents performance configuration
type PerfConfig struct {
	MaxConcurrency int   `yaml:"max_concurrency"`
	EnableSIMD     bool  `yaml:"enable_simd"`
	MemoryLimit    int64 `yaml:"memory_limit"`
	GCTarget       int   `yaml:"gc_target"`
}

// Database interface represents the main database operations
type Database interface {
	// Lifecycle
	Open(ctx context.Context, config *Config) error
	Close() error
	Health() *HealthStatus

	// Collection management
	CreateCollection(ctx context.Context, req *CreateCollectionRequest) error
	GetCollection(ctx context.Context, name string) (Collection, error)
	ListCollections(ctx context.Context) ([]*CollectionInfo, error)
	DropCollection(ctx context.Context, name string) error

	// Statistics and maintenance
	Stats(ctx context.Context) (*DatabaseStats, error)
	Backup(ctx context.Context, w io.Writer) error
	Restore(ctx context.Context, r io.Reader) error
}

// Collection interface represents vector collection operations
type Collection interface {
	// Metadata
	Name() string
	Dimensions() int
	Metric() DistanceMetric
	Count() (int64, error)

	// Vector operations
	Insert(ctx context.Context, vector *Vector) error
	InsertBatch(ctx context.Context, vectors []*Vector) error
	Get(ctx context.Context, id string) (*Vector, error)
	Delete(ctx context.Context, id string) error

	// Search
	Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error)

	// Maintenance
	Compact(ctx context.Context) error
	Flush(ctx context.Context) error
}
