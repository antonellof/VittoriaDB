package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// CLIManager provides command-line interface for configuration management
type CLIManager struct {
	config *VittoriaConfig
}

// NewCLIManager creates a new CLI configuration manager
func NewCLIManager() *CLIManager {
	return &CLIManager{}
}

// GenerateConfig generates a sample configuration file
func (cli *CLIManager) GenerateConfig(outputPath string, includeComments bool) error {
	config := DefaultConfig()

	var data []byte
	var err error

	if includeComments {
		data, err = cli.marshalWithComments(config)
	} else {
		data, err = yaml.Marshal(config)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration file generated: %s\n", outputPath)
	return nil
}

// ValidateConfig validates a configuration file
func (cli *CLIManager) ValidateConfig(configPath string) error {
	config, err := LoadConfigFromFile(configPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create manager with validators
	manager := CreateDefaultManager()

	// Run all validators
	for _, validator := range manager.validators {
		if err := validator.Validate(config); err != nil {
			fmt.Printf("⚠️  Warning from %s validator:\n%s\n\n", validator.Name(), err.Error())
		}
	}

	fmt.Printf("✅ Configuration is valid: %s\n", configPath)
	return nil
}

// ShowConfig displays the current configuration
func (cli *CLIManager) ShowConfig(configPath string, format string) error {
	var config *VittoriaConfig
	var err error

	if configPath != "" {
		config, err = LoadConfigFromFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		config = DefaultConfig()
	}

	switch strings.ToLower(format) {
	case "yaml", "yml":
		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		fmt.Print(string(data))

	case "table":
		cli.printConfigTable(config)

	default:
		return fmt.Errorf("unsupported format: %s (supported: yaml, table)", format)
	}

	return nil
}

// CompareConfigs compares two configuration files
func (cli *CLIManager) CompareConfigs(config1Path, config2Path string) error {
	cfg1, err := LoadConfigFromFile(config1Path)
	if err != nil {
		return fmt.Errorf("failed to load config1: %w", err)
	}

	cfg2, err := LoadConfigFromFile(config2Path)
	if err != nil {
		return fmt.Errorf("failed to load config2: %w", err)
	}

	// Convert to YAML for comparison
	data1, _ := yaml.Marshal(cfg1)
	data2, _ := yaml.Marshal(cfg2)

	if string(data1) == string(data2) {
		fmt.Println("✅ Configurations are identical")
		return nil
	}

	fmt.Printf("📊 Configuration differences between %s and %s:\n\n", config1Path, config2Path)

	// Simple line-by-line comparison
	lines1 := strings.Split(string(data1), "\n")
	lines2 := strings.Split(string(data2), "\n")

	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}

	for i := 0; i < maxLines; i++ {
		line1 := ""
		line2 := ""

		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			fmt.Printf("Line %d:\n", i+1)
			fmt.Printf("  %s: %s\n", filepath.Base(config1Path), line1)
			fmt.Printf("  %s: %s\n", filepath.Base(config2Path), line2)
			fmt.Println()
		}
	}

	return nil
}

// ListEnvVars lists all supported environment variables
func (cli *CLIManager) ListEnvVars(prefix string) {
	fmt.Printf("🌍 Environment Variables (prefix: %s)\n\n", prefix)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "VARIABLE\tDESCRIPTION\tDEFAULT")
	fmt.Fprintln(w, "--------\t-----------\t-------")

	// Server configuration
	fmt.Fprintf(w, "%sHOST\tServer host address\tlocalhost\n", prefix)
	fmt.Fprintf(w, "%sPORT\tServer port number\t8080\n", prefix)
	fmt.Fprintf(w, "%sREAD_TIMEOUT\tHTTP read timeout\t30s\n", prefix)
	fmt.Fprintf(w, "%sWRITE_TIMEOUT\tHTTP write timeout\t30s\n", prefix)
	fmt.Fprintf(w, "%sMAX_BODY_SIZE\tMaximum request body size\t33554432\n", prefix)
	fmt.Fprintf(w, "%sCORS\tEnable CORS\ttrue\n", prefix)

	// Storage configuration
	fmt.Fprintf(w, "%sSTORAGE_ENGINE\tStorage engine type\tfile\n", prefix)
	fmt.Fprintf(w, "%sSTORAGE_PAGE_SIZE\tStorage page size\t4096\n", prefix)
	fmt.Fprintf(w, "%sSTORAGE_CACHE_SIZE\tStorage cache size\t1000\n", prefix)
	fmt.Fprintf(w, "%sSTORAGE_SYNC_WRITES\tSync writes to disk\ttrue\n", prefix)

	// Search configuration
	fmt.Fprintf(w, "%sSEARCH_PARALLEL_ENABLED\tEnable parallel search\ttrue\n", prefix)
	fmt.Fprintf(w, "%sSEARCH_PARALLEL_MAX_WORKERS\tMax parallel workers\t%d\n", prefix, DefaultConfig().Search.Parallel.MaxWorkers)
	fmt.Fprintf(w, "%sSEARCH_CACHE_ENABLED\tEnable search cache\ttrue\n", prefix)
	fmt.Fprintf(w, "%sSEARCH_CACHE_MAX_ENTRIES\tMax cache entries\t1000\n", prefix)

	// Embeddings configuration
	fmt.Fprintf(w, "%sEMBEDDINGS_DEFAULT_TYPE\tDefault vectorizer type\tsentence_transformers\n", prefix)
	fmt.Fprintf(w, "%sEMBEDDINGS_DEFAULT_MODEL\tDefault model name\tall-MiniLM-L6-v2\n", prefix)
	fmt.Fprintf(w, "%sEMBEDDINGS_BATCH_ENABLED\tEnable batch processing\ttrue\n", prefix)

	// Performance configuration
	fmt.Fprintf(w, "%sPERF_MAX_CONCURRENCY\tMax concurrency\t%d\n", prefix, DefaultConfig().Performance.MaxConcurrency)
	fmt.Fprintf(w, "%sPERF_ENABLE_SIMD\tEnable SIMD optimizations\ttrue\n", prefix)
	fmt.Fprintf(w, "%sPERF_IO_USE_MEMORY_MAP\tUse memory-mapped I/O\ttrue\n", prefix)

	// Logging configuration
	fmt.Fprintf(w, "%sLOG_LEVEL\tLogging level\tinfo\n", prefix)
	fmt.Fprintf(w, "%sLOG_FORMAT\tLogging format\ttext\n", prefix)
	fmt.Fprintf(w, "%sLOG_OUTPUT\tLogging output\tstdout\n", prefix)

	// Data directory
	fmt.Fprintf(w, "%sDATA_DIR\tData directory path\tdata\n", prefix)

	w.Flush()

	fmt.Printf("\nExample usage:\n")
	fmt.Printf("  export %sPORT=9090\n", prefix)
	fmt.Printf("  export %sLOG_LEVEL=debug\n", prefix)
	fmt.Printf("  export %sDATA_DIR=/var/lib/vittoriadb\n", prefix)
}

// CheckEnvironment checks the current environment for configuration
func (cli *CLIManager) CheckEnvironment(prefix string) {
	fmt.Printf("🔍 Environment Configuration Check (prefix: %s)\n\n", prefix)

	config, err := LoadConfigFromEnv(prefix)
	if err != nil {
		fmt.Printf("❌ Error loading from environment: %v\n", err)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SETTING\tVALUE\tSOURCE")
	fmt.Fprintln(w, "-------\t-----\t------")

	// Check key settings
	fmt.Fprintf(w, "Server Host\t%s\t%s\n", config.Server.Host, cli.getEnvSource(prefix+"HOST"))
	fmt.Fprintf(w, "Server Port\t%d\t%s\n", config.Server.Port, cli.getEnvSource(prefix+"PORT"))
	fmt.Fprintf(w, "Data Directory\t%s\t%s\n", config.DataDir, cli.getEnvSource(prefix+"DATA_DIR"))
	fmt.Fprintf(w, "Log Level\t%s\t%s\n", config.Logging.Level, cli.getEnvSource(prefix+"LOG_LEVEL"))
	fmt.Fprintf(w, "Cache Size\t%d\t%s\n", config.Storage.CacheSize, cli.getEnvSource(prefix+"STORAGE_CACHE_SIZE"))
	fmt.Fprintf(w, "Parallel Search\t%t\t%s\n", config.Search.Parallel.Enabled, cli.getEnvSource(prefix+"SEARCH_PARALLEL_ENABLED"))

	w.Flush()
}

func (cli *CLIManager) getEnvSource(envVar string) string {
	if value := os.Getenv(envVar); value != "" {
		return "environment"
	}
	return "default"
}

func (cli *CLIManager) printConfigTable(config *VittoriaConfig) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "SECTION\tSETTING\tVALUE")
	fmt.Fprintln(w, "-------\t-------\t-----")

	// Server settings
	fmt.Fprintf(w, "Server\tHost\t%s\n", config.Server.Host)
	fmt.Fprintf(w, "Server\tPort\t%d\n", config.Server.Port)
	fmt.Fprintf(w, "Server\tCORS\t%t\n", config.Server.CORS)
	fmt.Fprintf(w, "Server\tTLS Enabled\t%t\n", config.Server.TLS.Enabled)

	// Storage settings
	fmt.Fprintf(w, "Storage\tEngine\t%s\n", config.Storage.Engine)
	fmt.Fprintf(w, "Storage\tPage Size\t%d\n", config.Storage.PageSize)
	fmt.Fprintf(w, "Storage\tCache Size\t%d\n", config.Storage.CacheSize)
	fmt.Fprintf(w, "Storage\tSync Writes\t%t\n", config.Storage.SyncWrites)

	// Search settings
	fmt.Fprintf(w, "Search\tParallel Enabled\t%t\n", config.Search.Parallel.Enabled)
	fmt.Fprintf(w, "Search\tMax Workers\t%d\n", config.Search.Parallel.MaxWorkers)
	fmt.Fprintf(w, "Search\tCache Enabled\t%t\n", config.Search.Cache.Enabled)
	fmt.Fprintf(w, "Search\tCache Max Entries\t%d\n", config.Search.Cache.MaxEntries)

	// Embeddings settings
	fmt.Fprintf(w, "Embeddings\tDefault Type\t%s\n", config.Embeddings.Default.Type)
	fmt.Fprintf(w, "Embeddings\tDefault Model\t%s\n", config.Embeddings.Default.Model)
	fmt.Fprintf(w, "Embeddings\tBatch Enabled\t%t\n", config.Embeddings.Batch.Enabled)
	fmt.Fprintf(w, "Embeddings\tBatch Size\t%d\n", config.Embeddings.Batch.DefaultBatchSize)

	// Performance settings
	fmt.Fprintf(w, "Performance\tMax Concurrency\t%d\n", config.Performance.MaxConcurrency)
	fmt.Fprintf(w, "Performance\tSIMD Enabled\t%t\n", config.Performance.EnableSIMD)
	fmt.Fprintf(w, "Performance\tMemory Map I/O\t%t\n", config.Performance.IO.UseMemoryMap)
	fmt.Fprintf(w, "Performance\tAsync I/O\t%t\n", config.Performance.IO.AsyncIO)

	// Logging settings
	fmt.Fprintf(w, "Logging\tLevel\t%s\n", config.Logging.Level)
	fmt.Fprintf(w, "Logging\tFormat\t%s\n", config.Logging.Format)
	fmt.Fprintf(w, "Logging\tOutput\t%s\n", config.Logging.Output)

	// General settings
	fmt.Fprintf(w, "General\tData Directory\t%s\n", config.DataDir)
	fmt.Fprintf(w, "General\tVersion\t%s\n", config.Version)

	w.Flush()
}

func (cli *CLIManager) marshalWithComments(config *VittoriaConfig) ([]byte, error) {
	// Create YAML with comments
	yamlContent := `# VittoriaDB Configuration File
# This file contains all configuration options for VittoriaDB
# Environment variables can override these settings using the VITTORIA_ prefix

# Server Configuration
server:
  host: "` + config.Server.Host + `"          # Server host address
  port: ` + fmt.Sprintf("%d", config.Server.Port) + `                    # Server port number
  read_timeout: ` + config.Server.ReadTimeout.String() + `      # HTTP read timeout
  write_timeout: ` + config.Server.WriteTimeout.String() + `     # HTTP write timeout
  max_body_size: ` + fmt.Sprintf("%d", config.Server.MaxBodySize) + `        # Maximum request body size (bytes)
  cors: ` + fmt.Sprintf("%t", config.Server.CORS) + `                   # Enable CORS support
  tls:
    enabled: ` + fmt.Sprintf("%t", config.Server.TLS.Enabled) + `           # Enable TLS/HTTPS
    cert_file: ""             # Path to TLS certificate file
    key_file: ""              # Path to TLS private key file

# Storage Configuration
storage:
  engine: "` + config.Storage.Engine + `"            # Storage engine type (file, memory)
  page_size: ` + fmt.Sprintf("%d", config.Storage.PageSize) + `             # Storage page size (bytes)
  cache_size: ` + fmt.Sprintf("%d", config.Storage.CacheSize) + `            # Number of pages to cache
  sync_writes: ` + fmt.Sprintf("%t", config.Storage.SyncWrites) + `          # Sync writes to disk immediately
  compression: ` + fmt.Sprintf("%t", config.Storage.Compression) + `         # Enable storage compression (future)
  wal:
    enabled: ` + fmt.Sprintf("%t", config.Storage.WAL.Enabled) + `           # Enable Write-Ahead Logging
    sync_interval: ` + config.Storage.WAL.SyncInterval.String() + `   # WAL sync interval
    max_size: ` + fmt.Sprintf("%d", config.Storage.WAL.MaxSize) + `        # Maximum WAL file size (bytes)
    checkpoint_age: ` + config.Storage.WAL.CheckpointAge.String() + ` # WAL checkpoint age

# Search Configuration
search:
  parallel:
    enabled: ` + fmt.Sprintf("%t", config.Search.Parallel.Enabled) + `        # Enable parallel search processing
    max_workers: ` + fmt.Sprintf("%d", config.Search.Parallel.MaxWorkers) + `        # Maximum parallel workers
    batch_size: ` + fmt.Sprintf("%d", config.Search.Parallel.BatchSize) + `         # Batch size for parallel processing
    use_cache: ` + fmt.Sprintf("%t", config.Search.Parallel.UseCache) + `          # Use search result caching
    preload_vectors: ` + fmt.Sprintf("%t", config.Search.Parallel.PreloadVectors) + ` # Preload vectors into memory
  cache:
    enabled: ` + fmt.Sprintf("%t", config.Search.Cache.Enabled) + `           # Enable search result caching
    max_entries: ` + fmt.Sprintf("%d", config.Search.Cache.MaxEntries) + `        # Maximum cache entries
    ttl: ` + config.Search.Cache.TTL.String() + `           # Cache entry time-to-live
    cleanup_interval: ` + config.Search.Cache.CleanupInterval.String() + ` # Cache cleanup interval
  index:
    default_type: "` + config.Search.Index.DefaultType + `"   # Default index type (flat, hnsw, ivf)
    default_metric: "` + config.Search.Index.DefaultMetric + `" # Default distance metric (cosine, euclidean)
  default_limit: ` + fmt.Sprintf("%d", config.Search.DefaultLimit) + `          # Default search result limit
  max_limit: ` + fmt.Sprintf("%d", config.Search.MaxLimit) + `             # Maximum search result limit

# Embeddings Configuration
embeddings:
  default:
    type: "` + config.Embeddings.Default.Type + `"  # Default vectorizer type
    model: "` + config.Embeddings.Default.Model + `"    # Default model name
    dimensions: ` + fmt.Sprintf("%d", config.Embeddings.Default.Dimensions) + `           # Vector dimensions
  batch:
    enabled: ` + fmt.Sprintf("%t", config.Embeddings.Batch.Enabled) + `           # Enable batch processing
    default_batch_size: ` + fmt.Sprintf("%d", config.Embeddings.Batch.DefaultBatchSize) + `  # Default batch size
    max_batch_size: ` + fmt.Sprintf("%d", config.Embeddings.Batch.MaxBatchSize) + `      # Maximum batch size
    max_retries: ` + fmt.Sprintf("%d", config.Embeddings.Batch.MaxRetries) + `         # Maximum retry attempts
    timeout: ` + config.Embeddings.Batch.Timeout.String() + `        # Batch processing timeout
  processing:
    chunk_size: ` + fmt.Sprintf("%d", config.Embeddings.Processing.ChunkSize) + `          # Text chunk size
    chunk_overlap: ` + fmt.Sprintf("%d", config.Embeddings.Processing.ChunkOverlap) + `       # Text chunk overlap
    strategy: "` + config.Embeddings.Processing.Strategy + `"        # Chunking strategy (smart, sentence, paragraph)

# Performance Configuration
performance:
  max_concurrency: ` + fmt.Sprintf("%d", config.Performance.MaxConcurrency) + `      # Maximum concurrent operations
  enable_simd: ` + fmt.Sprintf("%t", config.Performance.EnableSIMD) + `          # Enable SIMD optimizations
  memory_limit: ` + fmt.Sprintf("%d", config.Performance.MemoryLimit) + `         # Memory limit (0 = unlimited)
  gc_target: ` + fmt.Sprintf("%d", config.Performance.GCTarget) + `            # Garbage collection target percentage
  io:
    use_memory_map: ` + fmt.Sprintf("%t", config.Performance.IO.UseMemoryMap) + `    # Use memory-mapped I/O
    async_io: ` + fmt.Sprintf("%t", config.Performance.IO.AsyncIO) + `          # Enable async I/O operations
    vectorized_ops: ` + fmt.Sprintf("%t", config.Performance.IO.VectorizedOps) + `    # Enable vectorized operations
  cpu:
    enable_simd: ` + fmt.Sprintf("%t", config.Performance.CPU.EnableSIMD) + `       # Enable CPU SIMD instructions
    vectorized_math: ` + fmt.Sprintf("%t", config.Performance.CPU.VectorizedMath) + `  # Enable vectorized math operations
    num_threads: ` + fmt.Sprintf("%d", config.Performance.CPU.NumThreads) + `        # Number of CPU threads to use

# Logging Configuration
logging:
  level: "` + config.Logging.Level + `"              # Log level (debug, info, warn, error)
  format: "` + config.Logging.Format + `"            # Log format (text, json)
  output: "` + config.Logging.Output + `"           # Log output (stdout, stderr, file)
  file: ""                  # Log file path (when output=file)
  max_size: ` + fmt.Sprintf("%d", config.Logging.MaxSize) + `              # Max log file size (MB)
  max_backups: ` + fmt.Sprintf("%d", config.Logging.MaxBackups) + `           # Max log file backups
  max_age: ` + config.Logging.MaxAge.String() + `        # Max log file age
  compress: ` + fmt.Sprintf("%t", config.Logging.Compress) + `            # Compress old log files

# General Configuration
data_dir: "` + config.DataDir + `"              # Data directory path
version: "` + config.Version + `"               # Configuration version
`

	return []byte(yamlContent), nil
}
