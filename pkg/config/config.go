package config

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// VittoriaConfig represents the unified configuration for VittoriaDB
type VittoriaConfig struct {
	// Server configuration
	Server ServerConfig `yaml:"server" json:"server" env:"VITTORIA_SERVER"`

	// Storage configuration
	Storage StorageConfig `yaml:"storage" json:"storage" env:"VITTORIA_STORAGE"`

	// Search configuration
	Search SearchConfig `yaml:"search" json:"search" env:"VITTORIA_SEARCH"`

	// Embeddings configuration
	Embeddings EmbeddingsConfig `yaml:"embeddings" json:"embeddings" env:"VITTORIA_EMBEDDINGS"`

	// Performance configuration
	Performance PerformanceConfig `yaml:"performance" json:"performance" env:"VITTORIA_PERF"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging" json:"logging" env:"VITTORIA_LOGGING"`

	// Data directory (overrides individual data dirs)
	DataDir string `yaml:"data_dir" json:"data_dir" env:"VITTORIA_DATA_DIR"`

	// Configuration metadata
	Version string `yaml:"version" json:"version"`
	Source  string `yaml:"-" json:"-"` // Where config was loaded from
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Host         string        `yaml:"host" json:"host" env:"HOST"`
	Port         int           `yaml:"port" json:"port" env:"PORT"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout" env:"READ_TIMEOUT"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" env:"WRITE_TIMEOUT"`
	MaxBodySize  int64         `yaml:"max_body_size" json:"max_body_size" env:"MAX_BODY_SIZE"`
	CORS         bool          `yaml:"cors" json:"cors" env:"CORS"`
	TLS          TLSConfig     `yaml:"tls" json:"tls"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled  bool   `yaml:"enabled" json:"enabled" env:"TLS_ENABLED"`
	CertFile string `yaml:"cert_file" json:"cert_file" env:"TLS_CERT_FILE"`
	KeyFile  string `yaml:"key_file" json:"key_file" env:"TLS_KEY_FILE"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Engine      string       `yaml:"engine" json:"engine" env:"ENGINE"` // "file", "memory"
	PageSize    int          `yaml:"page_size" json:"page_size" env:"PAGE_SIZE"`
	CacheSize   int          `yaml:"cache_size" json:"cache_size" env:"CACHE_SIZE"`
	SyncWrites  bool         `yaml:"sync_writes" json:"sync_writes" env:"SYNC_WRITES"`
	WAL         WALConfig    `yaml:"wal" json:"wal"`
	Backup      BackupConfig `yaml:"backup" json:"backup"`
	Compression bool         `yaml:"compression" json:"compression" env:"COMPRESSION"` // For future use
}

// WALConfig represents Write-Ahead Log configuration
type WALConfig struct {
	Enabled       bool          `yaml:"enabled" json:"enabled" env:"WAL_ENABLED"`
	SyncInterval  time.Duration `yaml:"sync_interval" json:"sync_interval" env:"WAL_SYNC_INTERVAL"`
	MaxSize       int64         `yaml:"max_size" json:"max_size" env:"WAL_MAX_SIZE"`
	CheckpointAge time.Duration `yaml:"checkpoint_age" json:"checkpoint_age" env:"WAL_CHECKPOINT_AGE"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	Enabled   bool          `yaml:"enabled" json:"enabled" env:"BACKUP_ENABLED"`
	Interval  time.Duration `yaml:"interval" json:"interval" env:"BACKUP_INTERVAL"`
	Retention int           `yaml:"retention" json:"retention" env:"BACKUP_RETENTION"`
	Directory string        `yaml:"directory" json:"directory" env:"BACKUP_DIRECTORY"`
}

// SearchConfig represents search configuration
type SearchConfig struct {
	// Parallel search settings
	Parallel ParallelSearchConfig `yaml:"parallel" json:"parallel"`

	// Cache settings
	Cache SearchCacheConfig `yaml:"cache" json:"cache"`

	// Index settings
	Index IndexConfig `yaml:"index" json:"index"`

	// Default search parameters
	DefaultLimit int     `yaml:"default_limit" json:"default_limit" env:"DEFAULT_LIMIT"`
	MaxLimit     int     `yaml:"max_limit" json:"max_limit" env:"MAX_LIMIT"`
	MinScore     float32 `yaml:"min_score" json:"min_score" env:"MIN_SCORE"`
}

// ParallelSearchConfig holds configuration for parallel search
type ParallelSearchConfig struct {
	Enabled               bool `yaml:"enabled" json:"enabled" env:"PARALLEL_ENABLED"`
	MaxWorkers            int  `yaml:"max_workers" json:"max_workers" env:"PARALLEL_MAX_WORKERS"`
	BatchSize             int  `yaml:"batch_size" json:"batch_size" env:"PARALLEL_BATCH_SIZE"`
	UseCache              bool `yaml:"use_cache" json:"use_cache" env:"PARALLEL_USE_CACHE"`
	PreloadVectors        bool `yaml:"preload_vectors" json:"preload_vectors" env:"PARALLEL_PRELOAD"`
	MinVectorsForParallel int  `yaml:"min_vectors_for_parallel" json:"min_vectors_for_parallel" env:"PARALLEL_MIN_VECTORS"`
}

// SearchCacheConfig holds configuration for search caching
type SearchCacheConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled" env:"CACHE_ENABLED"`
	MaxEntries      int           `yaml:"max_entries" json:"max_entries" env:"CACHE_MAX_ENTRIES"`
	TTL             time.Duration `yaml:"ttl" json:"ttl" env:"CACHE_TTL"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval" env:"CACHE_CLEANUP_INTERVAL"`
}

// IndexConfig represents index configuration
type IndexConfig struct {
	DefaultType   string     `yaml:"default_type" json:"default_type" env:"INDEX_DEFAULT_TYPE"`
	DefaultMetric string     `yaml:"default_metric" json:"default_metric" env:"INDEX_DEFAULT_METRIC"`
	HNSW          HNSWConfig `yaml:"hnsw" json:"hnsw"`
	Flat          FlatConfig `yaml:"flat" json:"flat"`
	IVF           IVFConfig  `yaml:"ivf" json:"ivf"`
}

// HNSWConfig represents HNSW index configuration
type HNSWConfig struct {
	M              int     `yaml:"m" json:"m" env:"HNSW_M"`
	MaxM           int     `yaml:"max_m" json:"max_m" env:"HNSW_MAX_M"`
	MaxM0          int     `yaml:"max_m0" json:"max_m0" env:"HNSW_MAX_M0"`
	ML             float64 `yaml:"ml" json:"ml" env:"HNSW_ML"`
	EfConstruction int     `yaml:"ef_construction" json:"ef_construction" env:"HNSW_EF_CONSTRUCTION"`
	EfSearch       int     `yaml:"ef_search" json:"ef_search" env:"HNSW_EF_SEARCH"`
	Seed           int64   `yaml:"seed" json:"seed" env:"HNSW_SEED"`
}

// FlatConfig represents flat index configuration
type FlatConfig struct {
	BatchSize int `yaml:"batch_size" json:"batch_size" env:"FLAT_BATCH_SIZE"`
}

// IVFConfig represents IVF index configuration
type IVFConfig struct {
	NClusters int `yaml:"n_clusters" json:"n_clusters" env:"IVF_N_CLUSTERS"`
	NProbe    int `yaml:"n_probe" json:"n_probe" env:"IVF_N_PROBE"`
}

// EmbeddingsConfig represents embeddings configuration
type EmbeddingsConfig struct {
	// Default vectorizer settings
	Default VectorizerConfig `yaml:"default" json:"default"`

	// Batch processing settings
	Batch BatchProcessorConfig `yaml:"batch" json:"batch"`

	// Processing settings
	Processing ProcessingConfig `yaml:"processing" json:"processing"`

	// Provider-specific settings
	OpenAI               OpenAIConfig               `yaml:"openai" json:"openai"`
	HuggingFace          HuggingFaceConfig          `yaml:"huggingface" json:"huggingface"`
	Ollama               OllamaConfig               `yaml:"ollama" json:"ollama"`
	SentenceTransformers SentenceTransformersConfig `yaml:"sentence_transformers" json:"sentence_transformers"`
}

// VectorizerConfig represents vectorizer configuration
type VectorizerConfig struct {
	Type       string                 `yaml:"type" json:"type" env:"VECTORIZER_TYPE"`
	Model      string                 `yaml:"model" json:"model" env:"VECTORIZER_MODEL"`
	Dimensions int                    `yaml:"dimensions" json:"dimensions" env:"VECTORIZER_DIMENSIONS"`
	Options    map[string]interface{} `yaml:"options" json:"options"`
}

// BatchProcessorConfig represents batch processing configuration
type BatchProcessorConfig struct {
	Enabled          bool          `yaml:"enabled" json:"enabled" env:"BATCH_ENABLED"`
	DefaultBatchSize int           `yaml:"default_batch_size" json:"default_batch_size" env:"BATCH_DEFAULT_SIZE"`
	MaxBatchSize     int           `yaml:"max_batch_size" json:"max_batch_size" env:"BATCH_MAX_SIZE"`
	MaxRetries       int           `yaml:"max_retries" json:"max_retries" env:"BATCH_MAX_RETRIES"`
	RetryDelay       time.Duration `yaml:"retry_delay" json:"retry_delay" env:"BATCH_RETRY_DELAY"`
	Timeout          time.Duration `yaml:"timeout" json:"timeout" env:"BATCH_TIMEOUT"`
	EnableFallback   bool          `yaml:"enable_fallback" json:"enable_fallback" env:"BATCH_ENABLE_FALLBACK"`
	CollectStats     bool          `yaml:"collect_stats" json:"collect_stats" env:"BATCH_COLLECT_STATS"`
}

// ProcessingConfig represents text processing configuration
type ProcessingConfig struct {
	ChunkSize    int               `yaml:"chunk_size" json:"chunk_size" env:"PROCESSING_CHUNK_SIZE"`
	ChunkOverlap int               `yaml:"chunk_overlap" json:"chunk_overlap" env:"PROCESSING_CHUNK_OVERLAP"`
	MinChunkSize int               `yaml:"min_chunk_size" json:"min_chunk_size" env:"PROCESSING_MIN_CHUNK_SIZE"`
	MaxChunkSize int               `yaml:"max_chunk_size" json:"max_chunk_size" env:"PROCESSING_MAX_CHUNK_SIZE"`
	Strategy     string            `yaml:"strategy" json:"strategy" env:"PROCESSING_STRATEGY"`
	Metadata     map[string]string `yaml:"metadata" json:"metadata"`
}

// Provider-specific configurations
type OpenAIConfig struct {
	APIKey     string          `yaml:"api_key" json:"-" env:"OPENAI_API_KEY"`
	BaseURL    string          `yaml:"base_url" json:"base_url" env:"OPENAI_BASE_URL"`
	Model      string          `yaml:"model" json:"model" env:"OPENAI_MODEL"`
	Timeout    time.Duration   `yaml:"timeout" json:"timeout" env:"OPENAI_TIMEOUT"`
	MaxRetries int             `yaml:"max_retries" json:"max_retries" env:"OPENAI_MAX_RETRIES"`
	RateLimit  RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
}

type HuggingFaceConfig struct {
	APIKey    string          `yaml:"api_key" json:"-" env:"HUGGINGFACE_API_KEY"`
	BaseURL   string          `yaml:"base_url" json:"base_url" env:"HUGGINGFACE_BASE_URL"`
	Model     string          `yaml:"model" json:"model" env:"HUGGINGFACE_MODEL"`
	Timeout   time.Duration   `yaml:"timeout" json:"timeout" env:"HUGGINGFACE_TIMEOUT"`
	RateLimit RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
}

type OllamaConfig struct {
	BaseURL string        `yaml:"base_url" json:"base_url" env:"OLLAMA_BASE_URL"`
	Model   string        `yaml:"model" json:"model" env:"OLLAMA_MODEL"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" env:"OLLAMA_TIMEOUT"`
}

type SentenceTransformersConfig struct {
	CacheDir    string `yaml:"cache_dir" json:"cache_dir" env:"SENTENCE_TRANSFORMERS_CACHE_DIR"`
	DeviceMap   string `yaml:"device_map" json:"device_map" env:"SENTENCE_TRANSFORMERS_DEVICE_MAP"`
	TrustRemote bool   `yaml:"trust_remote" json:"trust_remote" env:"SENTENCE_TRANSFORMERS_TRUST_REMOTE"`
}

type RateLimitConfig struct {
	RequestsPerSecond int           `yaml:"requests_per_second" json:"requests_per_second"`
	BurstSize         int           `yaml:"burst_size" json:"burst_size"`
	Timeout           time.Duration `yaml:"timeout" json:"timeout"`
}

// PerformanceConfig represents performance configuration
type PerformanceConfig struct {
	MaxConcurrency int   `yaml:"max_concurrency" json:"max_concurrency" env:"PERF_MAX_CONCURRENCY"`
	EnableSIMD     bool  `yaml:"enable_simd" json:"enable_simd" env:"PERF_ENABLE_SIMD"`
	MemoryLimit    int64 `yaml:"memory_limit" json:"memory_limit" env:"PERF_MEMORY_LIMIT"`
	GCTarget       int   `yaml:"gc_target" json:"gc_target" env:"PERF_GC_TARGET"`

	// I/O optimization settings
	IO IOConfig `yaml:"io" json:"io"`

	// CPU optimization settings
	CPU CPUConfig `yaml:"cpu" json:"cpu"`
}

// IOConfig represents I/O optimization configuration
type IOConfig struct {
	UseMemoryMap    bool `yaml:"use_memory_map" json:"use_memory_map" env:"IO_USE_MEMORY_MAP"`
	AsyncIO         bool `yaml:"async_io" json:"async_io" env:"IO_ASYNC"`
	VectorizedOps   bool `yaml:"vectorized_ops" json:"vectorized_ops" env:"IO_VECTORIZED_OPS"`
	ReadAheadSize   int  `yaml:"read_ahead_size" json:"read_ahead_size" env:"IO_READ_AHEAD_SIZE"`
	WriteBufferSize int  `yaml:"write_buffer_size" json:"write_buffer_size" env:"IO_WRITE_BUFFER_SIZE"`
}

// CPUConfig represents CPU optimization configuration
type CPUConfig struct {
	EnableSIMD      bool `yaml:"enable_simd" json:"enable_simd" env:"CPU_ENABLE_SIMD"`
	VectorizedMath  bool `yaml:"vectorized_math" json:"vectorized_math" env:"CPU_VECTORIZED_MATH"`
	ParallelCompute bool `yaml:"parallel_compute" json:"parallel_compute" env:"CPU_PARALLEL_COMPUTE"`
	NumThreads      int  `yaml:"num_threads" json:"num_threads" env:"CPU_NUM_THREADS"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string        `yaml:"level" json:"level" env:"LOG_LEVEL"`
	Format     string        `yaml:"format" json:"format" env:"LOG_FORMAT"` // "json", "text"
	Output     string        `yaml:"output" json:"output" env:"LOG_OUTPUT"` // "stdout", "stderr", "file"
	File       string        `yaml:"file" json:"file" env:"LOG_FILE"`
	MaxSize    int           `yaml:"max_size" json:"max_size" env:"LOG_MAX_SIZE"` // MB
	MaxBackups int           `yaml:"max_backups" json:"max_backups" env:"LOG_MAX_BACKUPS"`
	MaxAge     time.Duration `yaml:"max_age" json:"max_age" env:"LOG_MAX_AGE"`
	Compress   bool          `yaml:"compress" json:"compress" env:"LOG_COMPRESS"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *VittoriaConfig {
	return &VittoriaConfig{
		Server: ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			MaxBodySize:  32 << 20, // 32MB
			CORS:         true,
			TLS: TLSConfig{
				Enabled: false,
			},
		},
		Storage: StorageConfig{
			Engine:      "file",
			PageSize:    4096,
			CacheSize:   1000,
			SyncWrites:  true,
			Compression: false,
			WAL: WALConfig{
				Enabled:       true,
				SyncInterval:  1 * time.Second,
				MaxSize:       100 << 20, // 100MB
				CheckpointAge: 5 * time.Minute,
			},
			Backup: BackupConfig{
				Enabled:   false,
				Interval:  24 * time.Hour,
				Retention: 7,
				Directory: "backups",
			},
		},
		Search: SearchConfig{
			Parallel: ParallelSearchConfig{
				Enabled:               true,
				MaxWorkers:            runtime.NumCPU(),
				BatchSize:             100,
				UseCache:              true,
				PreloadVectors:        false,
				MinVectorsForParallel: runtime.NumCPU() * 100,
			},
			Cache: SearchCacheConfig{
				Enabled:         true,
				MaxEntries:      1000,
				TTL:             5 * time.Minute,
				CleanupInterval: 1 * time.Minute,
			},
			Index: IndexConfig{
				DefaultType:   "flat",
				DefaultMetric: "cosine",
				HNSW: HNSWConfig{
					M:              16,
					MaxM:           16,
					MaxM0:          32,
					ML:             1.0 / math.Log(2.0),
					EfConstruction: 200,
					EfSearch:       50,
					Seed:           42,
				},
				Flat: FlatConfig{
					BatchSize: 1000,
				},
				IVF: IVFConfig{
					NClusters: 100,
					NProbe:    10,
				},
			},
			DefaultLimit: 10,
			MaxLimit:     1000,
			MinScore:     0.0,
		},
		Embeddings: EmbeddingsConfig{
			Default: VectorizerConfig{
				Type:       "sentence_transformers",
				Model:      "all-MiniLM-L6-v2",
				Dimensions: 384,
				Options:    make(map[string]interface{}),
			},
			Batch: BatchProcessorConfig{
				Enabled:          true,
				DefaultBatchSize: 32,
				MaxBatchSize:     128,
				MaxRetries:       3,
				RetryDelay:       1 * time.Second,
				Timeout:          30 * time.Second,
				EnableFallback:   true,
				CollectStats:     true,
			},
			Processing: ProcessingConfig{
				ChunkSize:    1024,
				ChunkOverlap: 128,
				MinChunkSize: 100,
				MaxChunkSize: 2048,
				Strategy:     "smart",
				Metadata:     make(map[string]string),
			},
			OpenAI: OpenAIConfig{
				BaseURL:    "https://api.openai.com/v1",
				Model:      "text-embedding-ada-002",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RateLimit: RateLimitConfig{
					RequestsPerSecond: 100,
					BurstSize:         200,
					Timeout:           5 * time.Second,
				},
			},
			HuggingFace: HuggingFaceConfig{
				BaseURL: "https://api-inference.huggingface.co",
				Model:   "sentence-transformers/all-MiniLM-L6-v2",
				Timeout: 30 * time.Second,
				RateLimit: RateLimitConfig{
					RequestsPerSecond: 10,
					BurstSize:         20,
					Timeout:           10 * time.Second,
				},
			},
			Ollama: OllamaConfig{
				BaseURL: "http://localhost:11434",
				Model:   "nomic-embed-text",
				Timeout: 30 * time.Second,
			},
			SentenceTransformers: SentenceTransformersConfig{
				CacheDir:    filepath.Join(os.TempDir(), "sentence_transformers_cache"),
				DeviceMap:   "auto",
				TrustRemote: false,
			},
		},
		Performance: PerformanceConfig{
			MaxConcurrency: runtime.NumCPU() * 2,
			EnableSIMD:     true,
			MemoryLimit:    0, // 0 = unlimited
			GCTarget:       100,
			IO: IOConfig{
				UseMemoryMap:    true,
				AsyncIO:         true,
				VectorizedOps:   true,
				ReadAheadSize:   64 * 1024,  // 64KB
				WriteBufferSize: 256 * 1024, // 256KB
			},
			CPU: CPUConfig{
				EnableSIMD:      true,
				VectorizedMath:  true,
				ParallelCompute: true,
				NumThreads:      runtime.NumCPU(),
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			MaxSize:    100, // 100MB
			MaxBackups: 3,
			MaxAge:     7 * 24 * time.Hour, // 7 days
			Compress:   true,
		},
		DataDir: "data",
		Version: "1.0",
	}
}

// LoadConfig loads configuration from multiple sources with precedence
func LoadConfig(sources ...ConfigSource) (*VittoriaConfig, error) {
	config := DefaultConfig()

	for _, source := range sources {
		if err := source.Load(config); err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", source.Name(), err)
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *VittoriaConfig) Validate() error {
	var errors []string

	// Server validation
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errors = append(errors, "server.port must be between 1 and 65535")
	}
	if c.Server.ReadTimeout <= 0 {
		errors = append(errors, "server.read_timeout must be positive")
	}
	if c.Server.WriteTimeout <= 0 {
		errors = append(errors, "server.write_timeout must be positive")
	}

	// Storage validation
	if c.Storage.PageSize <= 0 || (c.Storage.PageSize&(c.Storage.PageSize-1)) != 0 {
		errors = append(errors, "storage.page_size must be a positive power of 2")
	}
	if c.Storage.CacheSize < 0 {
		errors = append(errors, "storage.cache_size must be non-negative")
	}

	// Search validation
	if c.Search.Parallel.MaxWorkers <= 0 {
		errors = append(errors, "search.parallel.max_workers must be positive")
	}
	if c.Search.Parallel.BatchSize <= 0 {
		errors = append(errors, "search.parallel.batch_size must be positive")
	}
	if c.Search.Cache.MaxEntries < 0 {
		errors = append(errors, "search.cache.max_entries must be non-negative")
	}
	if c.Search.DefaultLimit <= 0 {
		errors = append(errors, "search.default_limit must be positive")
	}
	if c.Search.MaxLimit < c.Search.DefaultLimit {
		errors = append(errors, "search.max_limit must be >= search.default_limit")
	}

	// Embeddings validation
	if c.Embeddings.Default.Dimensions <= 0 {
		errors = append(errors, "embeddings.default.dimensions must be positive")
	}
	if c.Embeddings.Batch.DefaultBatchSize <= 0 {
		errors = append(errors, "embeddings.batch.default_batch_size must be positive")
	}
	if c.Embeddings.Batch.MaxBatchSize < c.Embeddings.Batch.DefaultBatchSize {
		errors = append(errors, "embeddings.batch.max_batch_size must be >= default_batch_size")
	}

	// Performance validation
	if c.Performance.MaxConcurrency <= 0 {
		errors = append(errors, "performance.max_concurrency must be positive")
	}
	if c.Performance.CPU.NumThreads <= 0 {
		errors = append(errors, "performance.cpu.num_threads must be positive")
	}

	// Data directory validation
	if c.DataDir == "" {
		errors = append(errors, "data_dir cannot be empty")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n- %s", strings.Join(errors, "\n- "))
	}

	return nil
}

// ToYAML converts the configuration to YAML format
func (c *VittoriaConfig) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// FromYAML loads configuration from YAML data
func (c *VittoriaConfig) FromYAML(data []byte) error {
	return yaml.Unmarshal(data, c)
}

// Clone creates a deep copy of the configuration
func (c *VittoriaConfig) Clone() *VittoriaConfig {
	data, _ := yaml.Marshal(c)
	clone := &VittoriaConfig{}
	yaml.Unmarshal(data, clone)
	return clone
}
