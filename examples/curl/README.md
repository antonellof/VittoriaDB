# VittoriaDB cURL/Bash Examples

This directory contains comprehensive bash scripts that demonstrate VittoriaDB functionality using cURL and HTTP API calls. These examples are perfect for:

- Understanding the VittoriaDB HTTP API
- Testing and debugging VittoriaDB operations
- Integration with shell scripts and automation
- Learning vector database concepts through hands-on examples

## 📋 Prerequisites

### System Requirements
- **bash** (4.0+)
- **curl** (for HTTP requests)
- **jq** (recommended for JSON processing)
- **bc** (for mathematical calculations)

### Installation
```bash
# macOS
brew install jq bc

# Ubuntu/Debian
sudo apt-get install jq bc

# CentOS/RHEL
sudo yum install jq bc
```

### VittoriaDB Server
Make sure VittoriaDB is running:
```bash
./vittoriadb run
```

## 🚀 Available Examples

### 1. Basic Usage (`basic_usage.sh`)
**Purpose:** Introduction to VittoriaDB HTTP API operations

**Features:**
- ✅ Connection testing and health checks
- ✅ Collection creation and management
- ✅ Individual vector insertion
- ✅ Batch vector operations
- ✅ Similarity search (basic and filtered)
- ✅ Database statistics and monitoring
- ✅ Performance comparison (individual vs batch)
- ✅ Proper cleanup and error handling

**Usage:**
```bash
cd examples/curl
chmod +x basic_usage.sh
./basic_usage.sh
```

**What you'll learn:**
- HTTP API endpoints and request formats
- JSON payload structures for vectors and metadata
- Search parameters and filtering options
- Performance optimization techniques
- Error handling and debugging

---

### 2. Volume Testing (`volume_test.sh`)
**Purpose:** Performance testing with different data volumes (KB, MB, GB)

**Features:**
- ✅ **KB-scale testing:** Small vectors, low dimensions (32D, 100 vectors)
- ✅ **MB-scale testing:** Medium vectors, moderate dimensions (256D, 1K vectors)
- ✅ **GB-scale testing:** Large vectors, high dimensions (512D, 5K vectors)
- ✅ Index type comparison (Flat vs HNSW)
- ✅ Memory usage monitoring
- ✅ Performance benchmarking and analysis
- ✅ Search parameter optimization (HNSW ef values)
- ✅ Stress testing and resource monitoring

**Usage:**
```bash
cd examples/curl
chmod +x volume_test.sh
./volume_test.sh
```

**Test Scenarios:**
| Scale | Dimensions | Vectors | Index Type | Estimated Size |
|-------|------------|---------|------------|----------------|
| KB    | 32         | 100     | Flat       | ~13 KB         |
| MB    | 256        | 1,000   | Flat       | ~1 MB          |
| GB    | 512        | 5,000   | HNSW       | ~10 MB         |

**Performance Metrics:**
- Insert throughput (vectors/second)
- Search latency (milliseconds)
- Memory usage (MB)
- Index build time
- Query accuracy (for HNSW)

---

### 3. RAG System (`rag_example.sh`)
**Purpose:** Complete Retrieval-Augmented Generation system implementation

**Features:**
- ✅ **Knowledge Base Creation:** Document chunking and ingestion
- ✅ **Text Embedding:** Simplified embedding generation (demo purposes)
- ✅ **Semantic Search:** Context-aware information retrieval
- ✅ **Answer Generation:** Simple response generation from context
- ✅ **Filtered Queries:** Category-based search filtering
- ✅ **Interactive Demo:** Real-time Q&A system
- ✅ **Performance Analysis:** Query timing and optimization
- ✅ **Multi-category Support:** Technology, RAG, embeddings, algorithms

**Usage:**
```bash
cd examples/curl
chmod +x rag_example.sh
./rag_example.sh
```

**Knowledge Base Topics:**
- Vector databases and VittoriaDB
- RAG architecture and implementation
- Embeddings and semantic search
- HNSW algorithms and indexing
- Machine learning in production

**Sample Queries:**
- "What is a vector database?"
- "How does VittoriaDB work?"
- "What is RAG?"
- "How do embeddings work?"
- "What is HNSW algorithm?"

**Interactive Features:**
- Real-time question answering
- Automatic category detection
- Confidence scoring
- Source attribution
- Context-aware responses

## 🛠️ Script Features

### Common Functionality
All scripts include:
- **Colored output** for better readability
- **Error handling** with meaningful messages
- **Progress indicators** for long operations
- **Performance timing** and metrics
- **Automatic cleanup** after execution
- **Connection testing** before operations
- **JSON validation** and pretty printing

### Helper Functions
```bash
print_header()    # Section headers
print_success()   # Success messages
print_error()     # Error messages
print_info()      # Information messages
print_perf()      # Performance metrics
```

### Configuration
Each script uses configurable variables:
```bash
BASE_URL="http://localhost:8080"    # VittoriaDB server URL
COLLECTION_NAME="demo_collection"   # Collection name
DIMENSIONS=128                      # Vector dimensions
```

## 📊 Performance Benchmarks

### Typical Performance (on modern hardware)

**Individual Operations:**
- Vector insertion: 50-100 vectors/sec
- Search queries: 100-500 queries/sec
- Collection operations: Near-instantaneous

**Batch Operations:**
- Batch insertion: 500-2000 vectors/sec
- Bulk search: 200-1000 queries/sec
- Memory usage: ~4 bytes per dimension per vector

**Index Performance:**
- **Flat Index:** Exact search, O(n) complexity
- **HNSW Index:** Approximate search, O(log n) complexity
- **Build time:** HNSW ~2-5x slower than Flat
- **Search time:** HNSW ~10-100x faster for large datasets

## 🔧 Troubleshooting

### Common Issues

**Connection Refused:**
```bash
curl: (7) Failed to connect to localhost port 8080: Connection refused
```
**Solution:** Start VittoriaDB with `./vittoriadb run`

**JSON Parse Error:**
```bash
parse error: Invalid numeric literal
```
**Solution:** Install `jq` for proper JSON processing

**Permission Denied:**
```bash
bash: ./basic_usage.sh: Permission denied
```
**Solution:** Make scripts executable with `chmod +x *.sh`

**Missing bc Calculator:**
```bash
bc: command not found
```
**Solution:** Install bc with your package manager

### Debug Mode
Run scripts with debug output:
```bash
bash -x ./basic_usage.sh
```

### Verbose cURL
Add verbose flag to see HTTP details:
```bash
# Edit script and add -v to curl commands
curl -v -s -X POST "$BASE_URL/collections" ...
```

## 🎯 Use Cases

### Development and Testing
- **API Testing:** Validate HTTP endpoints and payloads
- **Performance Testing:** Benchmark different configurations
- **Integration Testing:** Test VittoriaDB with other systems
- **Debugging:** Troubleshoot issues with detailed output

### Production Automation
- **Data Migration:** Bulk import/export operations
- **Health Monitoring:** Automated health checks
- **Backup Scripts:** Collection backup and restore
- **Deployment Testing:** Validate deployments

### Learning and Education
- **Vector Database Concepts:** Hands-on learning
- **API Understanding:** HTTP API exploration
- **Performance Analysis:** Optimization techniques
- **RAG Implementation:** End-to-end system building

## 📈 Advanced Usage

### Custom Configurations
Modify scripts for your specific needs:

```bash
# High-dimensional vectors
DIMENSIONS=1024

# Large batch sizes
BATCH_SIZE=500

# Custom index parameters
INDEX_CONFIG='{
    "m": 32,
    "ef_construction": 400,
    "ef_search": 100
}'
```

### Integration Examples
Use in CI/CD pipelines:

```bash
# Health check
if ./basic_usage.sh > /dev/null 2>&1; then
    echo "VittoriaDB is healthy"
else
    echo "VittoriaDB health check failed"
    exit 1
fi

# Performance regression testing
./volume_test.sh | grep "vectors/sec" > performance.log
```

### Monitoring Integration
Extract metrics for monitoring systems:

```bash
# Extract performance metrics
./volume_test.sh | grep "📊" | sed 's/📊 //' > metrics.txt

# Get memory usage
curl -s http://localhost:8080/stats | jq '.memory_usage'
```

## 🔗 Related Examples

After trying these cURL examples, explore:

- **Go Examples:** `../go/` - Native Go client usage
- **Python Examples:** `../python/` - Python library integration
- **Document Processing:** Various file format handling
- **Production Deployment:** Real-world configuration examples

## 💡 Tips and Best Practices

### Performance Optimization
1. **Use batch operations** for better throughput
2. **Choose appropriate index types** based on dataset size
3. **Monitor memory usage** during large operations
4. **Tune HNSW parameters** for your specific use case

### Error Handling
1. **Always check HTTP status codes**
2. **Validate JSON responses** before processing
3. **Implement retry logic** for transient failures
4. **Clean up resources** after operations

### Security Considerations
1. **Use HTTPS** in production environments
2. **Implement authentication** for API access
3. **Validate input data** before processing
4. **Monitor API usage** for unusual patterns

---

**🚀 Happy scripting with VittoriaDB!**

For more information, visit the [main documentation](../../README.md) or explore other example directories.
