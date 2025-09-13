# Configuration Guide

VittoriaDB provides flexible configuration options for different deployment scenarios, from development to production.

## 🔧 Configuration Methods

### 1. Command Line Arguments
```bash
vittoriadb run \
  --host 0.0.0.0 \
  --port 8080 \
  --data-dir ./data \
  --cors
```

### 2. Environment Variables
```bash
export VITTORIADB_HOST=0.0.0.0
export VITTORIADB_PORT=8080
export VITTORIADB_DATA_DIR=/var/lib/vittoriadb
vittoriadb run
```

### 3. Configuration File
```bash
vittoriadb run --config ./config/vittoriadb.yaml
```

## 📁 Data Directory Configuration

### Default Location
- **Default**: `./data` (relative to where you run the command)
- **Recommended Production**: `/var/lib/vittoriadb`

### Configuration Options
```bash
# Command line flag
vittoriadb run --data-dir /path/to/your/data

# Environment variable
export VITTORIADB_DATA_DIR=/path/to/your/data
vittoriadb run

# Custom location examples
vittoriadb run --data-dir ~/vittoriadb-data
vittoriadb run --data-dir /var/lib/vittoriadb
vittoriadb run --data-dir ./my-vectors
```

### File Structure
```
data/                           # Main data directory
├── collection1/               # Each collection has its own directory
│   ├── metadata.json         # Collection metadata and schema
│   ├── vectors.json          # Vector data (current implementation)
│   ├── vectors.db            # Main database file (planned)
│   ├── vectors.db.wal        # Write-Ahead Log for durability
│   └── index.hnsw            # HNSW index file
├── collection2/
│   ├── metadata.json
│   └── vectors.json
└── .vittoriadb/              # Global database metadata (planned)
    ├── config.json
    └── locks/
```

## 🌐 Server Configuration

### Basic Server Settings
```bash
vittoriadb run \
  --host 0.0.0.0 \              # Bind host (default: localhost)
  --port 8080 \                 # Port to listen on (default: 8080)
  --data-dir ./data \           # Data directory (default: ./data)
  --config config.yaml \        # Configuration file
  --cors                        # Enable CORS (default: true)
```

### Advanced Server Configuration
```yaml
# vittoriadb.yaml
server:
  host: "0.0.0.0"
  port: 8080
  cors: true
  read_timeout: "30s"
  write_timeout: "30s"
  max_body_size: 33554432  # 32MB
```

## 🗄️ Storage Configuration

### Storage Engine Settings
```yaml
storage:
  data_dir: "./data"
  page_size: 4096
  cache_size: 100
  sync_writes: true
  compression: false
```

**Parameters:**
- `data_dir`: Directory for storing database files
- `page_size`: Size of storage pages in bytes (default: 4096)
- `cache_size`: Number of pages to cache in memory (default: 100)
- `sync_writes`: Force synchronous writes for durability (default: true)
- `compression`: Enable data compression (default: false)

### Write-Ahead Logging (WAL)
```yaml
storage:
  wal:
    enabled: true
    sync_interval: "1s"
    checkpoint_interval: "60s"
    max_file_size: 67108864  # 64MB
```

## 📊 Index Configuration

### Default Index Settings
```yaml
index:
  default_type: "hnsw"
  default_metric: "cosine"
```

### HNSW Index Configuration
```yaml
index:
  hnsw:
    m: 16                    # Number of bi-directional links for each node
    max_m: 16               # Maximum number of connections for level 0
    max_m0: 32              # Maximum number of connections for higher levels
    ml: 1.0                 # Level generation factor
    ef_construction: 200    # Size of dynamic candidate list during construction
    ef_search: 50           # Size of dynamic candidate list during search
    seed: 42                # Random seed for reproducible results
```

**HNSW Parameters Explained:**
- `m`: Controls index quality vs memory usage (higher = better quality, more memory)
- `ef_construction`: Higher values improve index quality but slow construction
- `ef_search`: Higher values improve search quality but slow search

### Flat Index Configuration
```yaml
index:
  flat:
    batch_size: 1000        # Batch size for operations
```

## ⚡ Performance Configuration

### General Performance Settings
```yaml
performance:
  max_concurrency: 100      # Maximum concurrent operations
  enable_simd: true         # Enable SIMD optimizations
  memory_limit: 1073741824  # 1GB memory limit
  gc_target: 10             # Go garbage collection target percentage
```

### Memory Management
```bash
# Set memory limit
vittoriadb run --memory-limit 2GB

# Adjust cache size
vittoriadb run --cache-size 200

# Enable SIMD optimizations
vittoriadb run --enable-simd
```

## 🔒 Security Configuration (Planned)

### Authentication
```yaml
security:
  auth:
    enabled: true
    method: "jwt"  # jwt, basic, api_key
    jwt:
      secret: "your-secret-key"
      expiry: "24h"
    basic:
      username: "admin"
      password: "secure-password"
```

### TLS/SSL
```yaml
security:
  tls:
    enabled: true
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
    ca_file: "/path/to/ca.pem"
```

## 📝 Logging Configuration

### Log Levels
```bash
# Set log level
vittoriadb run --log-level debug

# Available levels: debug, info, warn, error
```

### Log Configuration
```yaml
logging:
  level: "info"
  format: "json"  # json, text
  output: "stdout"  # stdout, stderr, file
  file:
    path: "/var/log/vittoriadb.log"
    max_size: 100  # MB
    max_backups: 3
    max_age: 28  # days
```

## 🌍 Environment Variables

### Complete Environment Variable List
```bash
# Server Configuration
export VITTORIADB_HOST=0.0.0.0
export VITTORIADB_PORT=8080
export VITTORIADB_CORS=true

# Data Configuration
export VITTORIADB_DATA_DIR=/var/lib/vittoriadb

# Performance Configuration
export VITTORIADB_MEMORY_LIMIT=2GB
export VITTORIADB_CACHE_SIZE=200
export VITTORIADB_MAX_CONCURRENCY=100

# Logging Configuration
export VITTORIADB_LOG_LEVEL=info
export VITTORIADB_LOG_FORMAT=json

# Configuration File
export VITTORIADB_CONFIG=/etc/vittoriadb/config.yaml
```

## 📋 Complete Configuration File Example

```yaml
# vittoriadb.yaml - Complete configuration example
server:
  host: "0.0.0.0"
  port: 8080
  cors: true
  read_timeout: "30s"
  write_timeout: "30s"
  max_body_size: 33554432

storage:
  data_dir: "/var/lib/vittoriadb"
  page_size: 4096
  cache_size: 200
  sync_writes: true
  compression: false
  wal:
    enabled: true
    sync_interval: "1s"
    checkpoint_interval: "60s"
    max_file_size: 67108864

index:
  default_type: "hnsw"
  default_metric: "cosine"
  hnsw:
    m: 16
    ef_construction: 200
    ef_search: 50
    seed: 42
  flat:
    batch_size: 1000

performance:
  max_concurrency: 100
  enable_simd: true
  memory_limit: 2147483648  # 2GB
  gc_target: 10

logging:
  level: "info"
  format: "json"
  output: "file"
  file:
    path: "/var/log/vittoriadb.log"
    max_size: 100
    max_backups: 3
    max_age: 28

# Future features
security:
  auth:
    enabled: false
  tls:
    enabled: false
```

## 🎯 Configuration for Different Environments

### Development Configuration
```yaml
# dev-config.yaml
server:
  host: "localhost"
  port: 8080
  cors: true

storage:
  data_dir: "./dev-data"
  sync_writes: false  # Faster for development

index:
  default_type: "flat"  # Faster startup

performance:
  max_concurrency: 10
  memory_limit: 536870912  # 512MB

logging:
  level: "debug"
  format: "text"
  output: "stdout"
```

### Production Configuration
```yaml
# prod-config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  cors: false
  read_timeout: "60s"
  write_timeout: "60s"

storage:
  data_dir: "/var/lib/vittoriadb"
  page_size: 4096
  cache_size: 500
  sync_writes: true
  compression: true

index:
  default_type: "hnsw"
  hnsw:
    m: 32
    ef_construction: 400
    ef_search: 100

performance:
  max_concurrency: 200
  enable_simd: true
  memory_limit: 4294967296  # 4GB
  gc_target: 5

logging:
  level: "info"
  format: "json"
  output: "file"
  file:
    path: "/var/log/vittoriadb.log"
    max_size: 100
    max_backups: 5
    max_age: 30
```

### Docker Configuration
```yaml
# docker-config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  cors: true

storage:
  data_dir: "/data"
  cache_size: 200

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

## 🔍 Configuration Validation

### Check Current Configuration
```bash
# Show current configuration
vittoriadb config show

# Validate configuration file
vittoriadb config validate --config ./config.yaml

# Show effective configuration (with defaults)
vittoriadb config effective --config ./config.yaml
```

### Database Information
```bash
# Show database information
vittoriadb info

# Show with custom data directory
vittoriadb info --data-dir /path/to/data

# Show database statistics
vittoriadb stats --data-dir /path/to/data
```

## 🚨 Configuration Troubleshooting

### Common Issues

**Configuration File Not Found**
```bash
# Check file path
ls -la /path/to/config.yaml

# Use absolute path
vittoriadb run --config /absolute/path/to/config.yaml
```

**Permission Issues**
```bash
# Check data directory permissions
ls -la /var/lib/vittoriadb

# Fix permissions
sudo chown -R vittoriadb:vittoriadb /var/lib/vittoriadb
sudo chmod -R 755 /var/lib/vittoriadb
```

**Port Already in Use**
```bash
# Check what's using the port
lsof -i :8080

# Use different port
vittoriadb run --port 9090
```

**Memory Issues**
```bash
# Check available memory
free -h

# Reduce cache size
vittoriadb run --cache-size 50

# Set memory limit
vittoriadb run --memory-limit 1GB
```
