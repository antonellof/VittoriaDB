# CLI Commands

VittoriaDB provides a comprehensive command-line interface for database management and operations.

## 🖥️ Core Commands

### Version Information
```bash
# Show version information
vittoriadb version
# VittoriaDB v0.2.0
# Build Time: 2025-09-13T11:45:48Z
# Git Commit: dce98c0
# Git Tag: v0.2.0
```

### Server Management
```bash
# Start the server
vittoriadb run [options]

# Start with custom configuration
vittoriadb run \
  --host 0.0.0.0 \
  --port 8080 \
  --data-dir ./data \
  --cors
```

### Database Inspection
```bash
# Show database information
vittoriadb info [--data-dir <path>]

# Show database statistics
vittoriadb stats [--data-dir <path>]

# Example output:
# 🚀 VittoriaDB v0.2.0 - Database Information
# =====================================
# 📁 Data Directory: /Users/you/project/data
# 📍 Relative Path: ./data
# 
# 📂 Collections (2 found):
#    • documents/
#      - metadata.json (245 B)
#      - vectors.json (1.2 KB)
#    • embeddings/
#      - metadata.json (198 B)
#      - vectors.json (856 B)
```

### Collection Management
```bash
# Create a collection
vittoriadb create <name> --dimensions <n> [options]

# Import data (planned)
vittoriadb import <file> --collection <name>

# Backup database (planned)
vittoriadb backup --output <file>

# Restore database (planned)
vittoriadb restore --input <file>
```

## ⚙️ Server Command Options

### Basic Options
```bash
vittoriadb run \
  --host 0.0.0.0 \              # Bind host (default: localhost)
  --port 8080 \                 # Port to listen on (default: 8080)
  --data-dir ./data \           # Data directory (default: ./data)
  --config config.yaml \        # Configuration file
  --cors                        # Enable CORS (default: true)
```

### Advanced Options
```bash
vittoriadb run \
  --log-level debug \           # Logging level (debug, info, warn, error)
  --memory-limit 2GB \          # Memory limit
  --enable-simd \               # Enable SIMD optimizations
  --cache-size 200              # Cache size in pages
```

## 🌍 Environment Variables

VittoriaDB supports configuration via environment variables:

```bash
# Data directory
export VITTORIADB_DATA_DIR=/path/to/data

# Server host
export VITTORIADB_HOST=0.0.0.0

# Server port
export VITTORIADB_PORT=8080

# Configuration file
export VITTORIADB_CONFIG=/path/to/config.yaml

# Log level
export VITTORIADB_LOG_LEVEL=debug
```

## 📁 Data Directory Management

### Default Behavior
- **Default Location**: `./data` (relative to where you run the command)
- **Auto-creation**: Directory is created if it doesn't exist
- **Permissions**: Ensures proper read/write permissions

### Custom Data Directory
```bash
# Command line flag (highest priority)
vittoriadb run --data-dir /path/to/your/data

# Environment variable
export VITTORIADB_DATA_DIR=/path/to/your/data
vittoriadb run

# Configuration file
vittoriadb run --config config.yaml  # data_dir specified in config
```

### Data Directory Examples
```bash
# Home directory
vittoriadb run --data-dir ~/vittoriadb-data

# System directory
vittoriadb run --data-dir /var/lib/vittoriadb

# Project-specific
vittoriadb run --data-dir ./my-vectors

# Temporary directory
vittoriadb run --data-dir /tmp/vittoriadb-test
```

## 🔍 Database Inspection Commands

### Basic Information
```bash
# Show current data directory and collections
vittoriadb info

# Show with custom data directory
vittoriadb info --data-dir /path/to/data
```

### Detailed Statistics
```bash
# Show database statistics
vittoriadb stats --data-dir /path/to/data

# Example output includes:
# - Collection counts
# - Vector counts per collection
# - Index sizes
# - Memory usage
# - Performance metrics
```

### File Structure Inspection
```bash
# The info command shows the complete file structure:
# data/
# ├── collection1/
# │   ├── metadata.json (245 B)
# │   ├── vectors.json (1.2 KB)
# │   └── index.hnsw (856 B)
# ├── collection2/
# │   ├── metadata.json (198 B)
# │   └── vectors.json (2.1 KB)
```

## 🚀 Startup Information

When starting VittoriaDB, you'll see comprehensive startup information:

```bash
vittoriadb run
# 🚀 VittoriaDB v0.2.0 starting...
# 📁 Data directory: /Users/you/project/data
# 🌐 HTTP server: http://localhost:8080
# 📊 Web dashboard: http://localhost:8080/
# ⚙️  Configuration:
#    • Index type: flat
#    • Distance metric: cosine
#    • Page size: 4096 bytes
#    • Cache size: 100 pages
#    • CORS enabled: true
```

## 🛠️ Configuration File

VittoriaDB supports YAML configuration files:

```yaml
# vittoriadb.yaml
server:
  host: "0.0.0.0"
  port: 8080
  cors: true

storage:
  data_dir: "./data"
  page_size: 4096
  cache_size: 100
  sync_writes: true

index:
  default_type: "hnsw"
  default_metric: "cosine"
  hnsw:
    m: 16
    ef_construction: 200
    ef_search: 50

performance:
  max_concurrency: 100
  enable_simd: true
  memory_limit: 1073741824  # 1GB

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### Using Configuration Files
```bash
# Specify config file
vittoriadb run --config ./config/vittoriadb.yaml

# Environment variable
export VITTORIADB_CONFIG=/etc/vittoriadb/config.yaml
vittoriadb run
```

## 🔧 Troubleshooting Commands

### Port Issues
```bash
# Check what's using port 8080
lsof -i :8080

# Use a different port
vittoriadb run --port 9090
```

### Permission Issues
```bash
# Make binary executable
chmod +x ./vittoriadb

# Check data directory permissions
ls -la ./data
```

### Debug Mode
```bash
# Run with verbose logging
vittoriadb run --log-level debug

# Check configuration
vittoriadb info --data-dir ./data
```

## 📋 Command Reference

| Command | Description | Options |
|---------|-------------|---------|
| `vittoriadb version` | Show version information | None |
| `vittoriadb run` | Start the server | `--host`, `--port`, `--data-dir`, `--config`, `--cors` |
| `vittoriadb info` | Show database information | `--data-dir` |
| `vittoriadb stats` | Show database statistics | `--data-dir` |
| `vittoriadb create` | Create collection | `--dimensions`, `--metric`, `--index-type` |

## 🔄 Process Management

### Background Execution
```bash
# Run in background
vittoriadb run &
SERVER_PID=$!

# Stop server
kill $SERVER_PID

# Or use process management
nohup vittoriadb run > vittoriadb.log 2>&1 &
```

### Service Integration
```bash
# systemd service example
sudo systemctl start vittoriadb
sudo systemctl enable vittoriadb
sudo systemctl status vittoriadb
```

## 📊 Performance Monitoring

### Built-in Metrics
```bash
# Server provides metrics at startup and via API
curl http://localhost:8080/stats

# CLI stats command shows:
# - Memory usage
# - Query performance
# - Index statistics
# - Collection metrics
```

### Log Analysis
```bash
# Enable debug logging for performance analysis
vittoriadb run --log-level debug > performance.log

# Monitor in real-time
tail -f performance.log | grep -E "(query|insert|search)"
```
