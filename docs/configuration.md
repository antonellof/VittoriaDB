# VittoriaDB Configuration Guide

VittoriaDB features a **unified configuration system** that provides comprehensive control over all aspects of the database while maintaining full backward compatibility with existing setups.

## 🎯 Quick Start

### Zero Configuration (Recommended for Development)
```bash
# Just works - no configuration needed!
vittoriadb run
```

### Basic Configuration
```bash
# Traditional CLI flags (fully backward compatible)
vittoriadb run --host 0.0.0.0 --port 8080 --data-dir ./data
```

### Advanced Configuration
```bash
# Generate and use YAML configuration
vittoriadb config generate --output vittoriadb.yaml
vittoriadb run --config vittoriadb.yaml
```

> 📖 **See the [main README](../README.md#-configuration) for basic configuration examples and quick setup.**

## 🔧 Configuration Methods

VittoriaDB supports multiple configuration sources with clear precedence:

### 1. **CLI Flags** (Highest Priority)
```bash
vittoriadb run --host 0.0.0.0 --port 9090 --data-dir /var/lib/vittoriadb
```

### 2. **Environment Variables**
```bash
# Legacy variables (still supported)
export VITTORIADB_HOST=0.0.0.0
export VITTORIADB_PORT=8080
export VITTORIADB_DATA_DIR=/var/lib/vittoriadb

# New unified variables (advanced features)
export VITTORIA_PERF_ENABLE_SIMD=true
export VITTORIA_SEARCH_PARALLEL_MAX_WORKERS=16
export VITTORIA_PERF_IO_USE_MEMORY_MAP=true

vittoriadb run
```

### 3. **YAML Configuration File**
```bash
vittoriadb run --config /etc/vittoriadb/config.yaml
```

### 4. **Sensible Defaults** (Lowest Priority)
Built-in defaults that work out of the box for development and testing.

## 📄 YAML Configuration Reference

### Complete Configuration Structure

```yaml
# VittoriaDB Unified Configuration
# All settings are optional - sensible defaults are provided

# General Settings
data_dir: "./data"                    # Data directory path

# Server Configuration
server:
  host: "localhost"                   # Bind address
  port: 8080                         # HTTP port
  read_timeout: "30s"                # Request read timeout
  write_timeout: "30s"               # Response write timeout
  max_body_size: 33554432            # Max request body size (32MB)
  cors: true                         # Enable CORS headers
  tls:
    enabled: false                   # Enable HTTPS
    cert_file: ""                    # TLS certificate file
    key_file: ""                     # TLS private key file

# Storage Configuration
storage:
  engine: "file"                     # Storage engine: "file" or "memory"
  page_size: 4096                    # Page size in bytes (must be multiple of 512)
  cache_size: 1000                   # Number of pages to cache
  sync_writes: true                  # Sync writes to disk immediately

# Search and Indexing Configuration
search:
  # Parallel Search Settings
  parallel:
    enabled: true                    # Enable parallel search processing
    max_workers: 10                  # Number of worker goroutines (default: CPU cores)
    batch_size: 100                  # Vectors processed per batch
    preload_vectors: false           # Preload vectors into memory
  
  # Search Cache Settings
  cache:
    enabled: true                    # Enable search result caching
    max_entries: 1000                # Maximum cached entries
    ttl: "5m"                        # Time-to-live for cached results
    cleanup_interval: "1m"           # Cache cleanup interval
  
  # Index Configuration
  index:
    default_type: "flat"             # Default index: "flat", "hnsw", "ivf"
    default_metric: "cosine"         # Default distance: "cosine", "euclidean", "dot_product", "manhattan"
    
    # HNSW Index Settings
    hnsw:
      m: 16                          # Number of bi-directional links for each node
      max_m: 32                      # Maximum connections for layer 0
      max_m0: 64                     # Maximum connections for higher layers
      ml: 1.442695                   # Level generation factor
      ef_construction: 100           # Size of dynamic candidate list during construction
      ef_search: 100                 # Size of dynamic candidate list during search
      seed: 42                       # Random seed for reproducible results
    
    # Flat Index Settings
    flat:
      batch_size: 1000               # Batch size for flat index operations

# Embeddings Configuration
embeddings:
  # Default Vectorizer Settings
  default:
    type: "sentence_transformers"    # Vectorizer type
    model: "all-MiniLM-L6-v2"       # Model name
    dimensions: 384                  # Vector dimensions
    options: {}                      # Additional options
  
  # Batch Processing Settings
  batch:
    enabled: true                    # Enable batch processing
    batch_size: 32                   # Default batch size
    fallback_size: 1                 # Fallback to individual processing
    max_workers: 10                  # Number of worker goroutines
    timeout: "60s"                   # Batch processing timeout
  
  # Text Processing Settings
  processing:
    strategy: "smart"                # Chunking strategy: "smart", "sentence", "fixed_size"
    chunk_size: 1024                 # Default chunk size in characters
    chunk_overlap: 128               # Overlap between chunks
    min_chunk_size: 100              # Minimum chunk size
    max_chunk_size: 2048             # Maximum chunk size
    language: "en"                   # Language for text processing
    metadata: {}                     # Default metadata

# Performance Configuration
performance:
  max_concurrency: 20                # Maximum concurrent operations
  enable_simd: true                  # Enable SIMD optimizations
  memory_limit: 2147483648           # Memory limit in bytes (2GB)
  gc_target: 100                     # Garbage collection target percentage
  
  # I/O Performance Settings
  io:
    use_memory_map: true             # Enable memory-mapped I/O
    async_io: true                   # Enable asynchronous I/O
    vectorized_ops: true             # Enable vectorized operations

# Logging Configuration
log:
  level: "info"                      # Log level: "debug", "info", "warn", "error"
  format: "text"                     # Log format: "text", "json"
  output: "stdout"                   # Log output: "stdout", "stderr", "file:/path/to/file.log"
```

## 🌍 Environment Variables Reference

### Legacy Variables (Backward Compatible)
```bash
VITTORIADB_HOST=localhost           # Server host
VITTORIADB_PORT=8080               # Server port
VITTORIADB_DATA_DIR=./data         # Data directory
VITTORIADB_CONFIG=/path/config.yaml # Configuration file
```

### Unified Configuration Variables

#### Server Settings
```bash
VITTORIA_HOST=localhost
VITTORIA_PORT=8080
VITTORIA_SERVER_READ_TIMEOUT=30s
VITTORIA_SERVER_WRITE_TIMEOUT=30s
VITTORIA_SERVER_MAX_BODY_SIZE=33554432
VITTORIA_SERVER_CORS=true
VITTORIA_SERVER_TLS_ENABLED=false
```

#### Storage Settings
```bash
VITTORIA_STORAGE_ENGINE=file
VITTORIA_STORAGE_PAGE_SIZE=4096
VITTORIA_STORAGE_CACHE_SIZE=1000
VITTORIA_STORAGE_SYNC_WRITES=true
```

#### Search and Performance Settings
```bash
VITTORIA_SEARCH_PARALLEL_ENABLED=true
VITTORIA_SEARCH_PARALLEL_MAX_WORKERS=10
VITTORIA_SEARCH_CACHE_ENABLED=true
VITTORIA_SEARCH_CACHE_MAX_ENTRIES=1000
VITTORIA_SEARCH_CACHE_TTL=5m0s

VITTORIA_PERF_ENABLE_SIMD=true
VITTORIA_PERF_IO_USE_MEMORY_MAP=true
VITTORIA_PERF_IO_ASYNC_IO=true
VITTORIA_PERF_MAX_CONCURRENCY=20
```

#### Embeddings Settings
```bash
VITTORIA_EMBEDDINGS_DEFAULT_TYPE=sentence_transformers
VITTORIA_EMBEDDINGS_DEFAULT_MODEL=all-MiniLM-L6-v2
VITTORIA_EMBEDDINGS_DEFAULT_DIMENSIONS=384
VITTORIA_EMBEDDINGS_BATCH_ENABLED=true
VITTORIA_EMBEDDINGS_BATCH_BATCH_SIZE=32
```

#### Logging Settings
```bash
VITTORIA_LOG_LEVEL=info
VITTORIA_LOG_FORMAT=text
VITTORIA_LOG_OUTPUT=stdout
```

## 📊 Configuration Parameters Explained

### Server Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `host` | string | `"localhost"` | IP address to bind the HTTP server to. Use `"0.0.0.0"` for all interfaces |
| `port` | int | `8080` | TCP port for the HTTP server |
| `read_timeout` | duration | `"30s"` | Maximum time to read request headers and body |
| `write_timeout` | duration | `"30s"` | Maximum time to write response |
| `max_body_size` | int64 | `33554432` | Maximum request body size in bytes (32MB) |
| `cors` | bool | `true` | Enable Cross-Origin Resource Sharing headers |

### Storage Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `engine` | string | `"file"` | Storage backend: `"file"` for persistent storage, `"memory"` for in-memory |
| `page_size` | int | `4096` | Page size in bytes (must be multiple of 512) |
| `cache_size` | int | `1000` | Number of pages to keep in memory cache |
| `sync_writes` | bool | `true` | Force sync writes to disk for durability |

### Search Configuration

#### Parallel Search
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | bool | `true` | Enable parallel search processing |
| `max_workers` | int | CPU cores | Number of goroutines for parallel processing |
| `batch_size` | int | `100` | Number of vectors processed per batch |
| `preload_vectors` | bool | `false` | Preload vectors into memory for faster access |

#### Search Cache
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | bool | `true` | Enable search result caching |
| `max_entries` | int | `1000` | Maximum number of cached search results |
| `ttl` | duration | `"5m"` | Time-to-live for cached results |
| `cleanup_interval` | duration | `"1m"` | How often to clean expired cache entries |

#### Index Configuration
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `default_type` | string | `"flat"` | Default index type: `"flat"`, `"hnsw"`, `"ivf"` |
| `default_metric` | string | `"cosine"` | Default distance metric: `"cosine"`, `"euclidean"`, `"dot_product"`, `"manhattan"` |

##### HNSW Index Parameters
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `m` | int | `16` | Number of bi-directional links for each node during construction |
| `max_m` | int | `32` | Maximum number of connections for layer 0 |
| `max_m0` | int | `64` | Maximum number of connections for higher layers |
| `ml` | float64 | `1.442695` | Level generation factor (1/ln(2)) |
| `ef_construction` | int | `100` | Size of dynamic candidate list during index construction |
| `ef_search` | int | `100` | Size of dynamic candidate list during search |
| `seed` | int64 | `42` | Random seed for reproducible index construction |

### Performance Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `max_concurrency` | int | CPU cores × 2 | Maximum number of concurrent operations |
| `enable_simd` | bool | `true` | Enable SIMD optimizations for vector operations |
| `memory_limit` | int64 | `2147483648` | Memory limit in bytes (2GB) |
| `gc_target` | int | `100` | Go garbage collection target percentage |

#### I/O Performance
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `use_memory_map` | bool | `true` | Enable memory-mapped I/O for zero-copy operations |
| `async_io` | bool | `true` | Enable asynchronous I/O operations |
| `vectorized_ops` | bool | `true` | Enable vectorized batch operations |

### Embeddings Configuration

#### Default Vectorizer
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `type` | string | `"sentence_transformers"` | Vectorizer type: `"sentence_transformers"`, `"openai"`, `"ollama"`, `"huggingface"` |
| `model` | string | `"all-MiniLM-L6-v2"` | Model name for the vectorizer |
| `dimensions` | int | `384` | Vector dimensions (must match model output) |
| `options` | map | `{}` | Additional options specific to the vectorizer |

#### Batch Processing
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | bool | `true` | Enable batch processing for embeddings |
| `batch_size` | int | `32` | Number of texts to process in each batch |
| `fallback_size` | int | `1` | Fallback to individual processing if batch fails |
| `max_workers` | int | CPU cores | Number of worker goroutines for batch processing |
| `timeout` | duration | `"60s"` | Timeout for batch processing operations |

#### Text Processing
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `strategy` | string | `"smart"` | Text chunking strategy: `"smart"`, `"sentence"`, `"fixed_size"` |
| `chunk_size` | int | `1024` | Target chunk size in characters |
| `chunk_overlap` | int | `128` | Overlap between consecutive chunks |
| `min_chunk_size` | int | `100` | Minimum allowed chunk size |
| `max_chunk_size` | int | `2048` | Maximum allowed chunk size |
| `language` | string | `"en"` | Language for text processing |

### Logging Configuration

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `level` | string | `"info"` | Log level: `"debug"`, `"info"`, `"warn"`, `"error"` |
| `format` | string | `"text"` | Log format: `"text"` for human-readable, `"json"` for structured |
| `output` | string | `"stdout"` | Log output: `"stdout"`, `"stderr"`, `"file:/path/to/file.log"` |

## 🛠️ Configuration Management Commands

### Generate Configuration
```bash
# Generate sample configuration file
vittoriadb config generate --output vittoriadb.yaml

# Generate with comments and documentation
vittoriadb config generate --output vittoriadb.yaml --include-comments
```

### Validate Configuration
```bash
# Validate configuration file
vittoriadb config validate --file vittoriadb.yaml

# Validate current environment configuration
vittoriadb config validate --env
```

### Inspect Configuration
```bash
# Show current configuration in table format
vittoriadb config show --format table

# Show configuration as JSON
vittoriadb config show --format json

# Show configuration as YAML
vittoriadb config show --format yaml
```

### Environment Variables
```bash
# List all available environment variables
vittoriadb config env --list

# Check current environment configuration
vittoriadb config env --check

# Show environment variables with values
vittoriadb config env --show-values
```

## 🔍 Runtime Configuration Inspection

### HTTP API Endpoint
```bash
# Get current configuration via HTTP API
curl http://localhost:8080/config

# Get specific configuration section
curl -s http://localhost:8080/config | jq '.config.performance'

# Get feature flags
curl -s http://localhost:8080/config | jq '.features'
```

### Configuration Response Structure
```json
{
  "config": {
    "server": { /* Server configuration */ },
    "storage": { /* Storage configuration */ },
    "search": { /* Search configuration */ },
    "embeddings": { /* Embeddings configuration */ },
    "performance": { /* Performance configuration */ },
    "log": { /* Logging configuration */ }
  },
  "features": {
    "parallel_search": true,
    "search_cache": true,
    "memory_mapped_io": true,
    "simd_optimizations": true,
    "async_io": true
  },
  "performance": {
    "max_workers": 10,
    "cache_entries": 1000,
    "cache_ttl": "5m0s",
    "max_concurrency": 20,
    "memory_limit_mb": 2048
  },
  "metadata": {
    "source": "default",
    "loaded_at": "2025-09-25T13:51:07+02:00",
    "version": "v1",
    "description": "VittoriaDB unified configuration"
  }
}
```

## 🚀 Production Configuration Examples

### High-Performance Setup
```yaml
# vittoriadb-production.yaml
server:
  host: "0.0.0.0"
  port: 8080
  cors: false
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/vittoriadb.crt"
    key_file: "/etc/ssl/private/vittoriadb.key"

storage:
  page_size: 8192
  cache_size: 10000
  sync_writes: true

search:
  parallel:
    enabled: true
    max_workers: 32
    batch_size: 500
  cache:
    enabled: true
    max_entries: 50000
    ttl: "10m"

performance:
  max_concurrency: 64
  enable_simd: true
  memory_limit: 8589934592  # 8GB
  io:
    use_memory_map: true
    async_io: true
    vectorized_ops: true

log:
  level: "warn"
  format: "json"
  output: "file:/var/log/vittoriadb/vittoriadb.log"
```

### Development Setup
```yaml
# vittoriadb-dev.yaml
server:
  host: "localhost"
  port: 8080
  cors: true

storage:
  cache_size: 100

search:
  parallel:
    enabled: true
    max_workers: 4
  cache:
    enabled: true
    max_entries: 1000
    ttl: "1m"

log:
  level: "debug"
  format: "text"
  output: "stdout"
```

### Memory-Optimized Setup
```yaml
# vittoriadb-memory-optimized.yaml
storage:
  cache_size: 500
  page_size: 2048

search:
  parallel:
    enabled: true
    max_workers: 8
    batch_size: 50
  cache:
    enabled: true
    max_entries: 5000
    ttl: "2m"

performance:
  max_concurrency: 16
  memory_limit: 1073741824  # 1GB
  gc_target: 50
```

## 🔧 Migration from Legacy Configuration

### Automatic Migration
VittoriaDB automatically migrates legacy configuration formats to the unified system. Your existing setups continue to work without changes.

### Manual Migration
```bash
# Convert legacy environment variables to unified format
export VITTORIA_HOST=$VITTORIADB_HOST
export VITTORIA_PORT=$VITTORIADB_PORT
export VITTORIA_DATA_DIR=$VITTORIADB_DATA_DIR

# Generate configuration from current environment
vittoriadb config generate --from-env --output migrated-config.yaml
```

## 🔍 Troubleshooting Configuration

### Common Issues

1. **Configuration not loading**
   ```bash
   # Check configuration file syntax
   vittoriadb config validate --file vittoriadb.yaml
   
   # Verify file permissions
   ls -la vittoriadb.yaml
   ```

2. **Environment variables not working**
   ```bash
   # List current environment configuration
   vittoriadb config env --check
   
   # Show all available variables
   vittoriadb config env --list
   ```

3. **Performance issues**
   ```bash
   # Check current performance settings
   curl -s http://localhost:8080/config | jq '.performance'
   
   # Verify SIMD and parallel search are enabled
   curl -s http://localhost:8080/config | jq '.features'
   ```

### Debug Configuration Loading
```bash
# Start with debug logging to see configuration loading
VITTORIA_LOG_LEVEL=debug vittoriadb run --config vittoriadb.yaml
```

## 📚 Related Documentation

- **[Main README](../README.md#-configuration)** - Basic configuration and quick start
- **[API Reference](api.md)** - HTTP API endpoints including `/config`
- **[CLI Reference](cli.md)** - Command-line interface and configuration commands
- **[Performance Guide](performance.md)** - Performance tuning and optimization
- **[Development Guide](development.md)** - Development setup and configuration

---

**Need help?** Check the [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues) or [Discussions](https://github.com/antonellof/VittoriaDB/discussions) for configuration-related questions.