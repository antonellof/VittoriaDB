# Performance Guide

This guide covers VittoriaDB's performance characteristics, optimization strategies, and benchmarking results.

## 📊 Performance Overview

### Benchmarks (v0.4.0)
- **Insert Speed**: >2.6M vectors/second (HNSW, small datasets), >1.7M vectors/second (large datasets)
- **Search Speed**: <1ms for small datasets (HNSW), sub-millisecond latency for optimized queries
- **Memory Usage**: Linear scaling - 1MB for 1K vectors, 167MB for 50K vectors (768 dimensions)
- **Startup Time**: <100ms (cold start), <50ms (warm start)
- **Binary Size**: ~8MB (compressed), ~25MB (uncompressed)
- **Index Build**: <2 seconds for 100k vectors (HNSW)
- **Document Processing**: >1000 documents/minute (PDF/DOCX)
- **Python Client**: Zero-overhead connection management

### Comprehensive Performance Results
📊 **[View Complete Benchmark Results](https://gist.github.com/antonellof/19069bb56573fcf72ce592b3c2f2fc74)** - Detailed performance testing with Native Go SDK integration

**Key Performance Highlights:**
- **Peak Insert Rate**: 2,645,209 vectors/sec (HNSW, small dataset)
- **Peak Search Rate**: 1,266.72 searches/sec (HNSW, small dataset)  
- **Lowest Latency**: 789.44µs (HNSW, small dataset)
- **Large-Scale Performance**: 1,685,330 vectors/sec for 87.89 MB dataset
- **Memory Efficiency**: Linear scaling with excellent performance characteristics

## 🎯 Performance Characteristics

### Scaling Characteristics
- **Vectors**: Tested up to 1M vectors (10M planned)
- **Dimensions**: Up to 2,048 dimensions (tested), 10,000+ supported
- **Collections**: Unlimited (limited by disk space)
- **File Size**: Individual collection files up to 2GB
- **Concurrent Users**: 100+ simultaneous connections
- **Throughput**: >1000 queries/second (HNSW), >100 queries/second (flat)

### Platform Performance
| Platform | Architecture | Relative Performance | Notes |
|----------|-------------|---------------------|-------|
| **Linux** | AMD64 | 100% (baseline) | Optimal performance |
| **Linux** | ARM64 | 95% | Excellent on modern ARM |
| **macOS** | Intel | 98% | Near-native performance |
| **macOS** | Apple Silicon | 105% | Superior ARM performance |
| **Windows** | AMD64 | 92% | Good cross-platform performance |

## ⚡ Performance Optimizations

### Index Optimization

#### HNSW Index
- **Hierarchical Navigable Small World** for sub-linear search
- **Best for**: Large datasets (>10k vectors)
- **Trade-offs**: Higher memory usage, faster search

**HNSW Parameters:**
```yaml
index:
  hnsw:
    m: 16                    # Higher = better quality, more memory
    ef_construction: 200     # Higher = better quality, slower build
    ef_search: 50           # Higher = better search, slower queries
```

#### Flat Index
- **Exact search** with linear scan
- **Best for**: Small datasets (<10k vectors), exact results required
- **Trade-offs**: Lower memory usage, slower search for large datasets

### Memory Optimization

#### Cache Configuration
```bash
# Increase cache size for better performance
vittoriadb run --cache-size 500

# Monitor memory usage
vittoriadb stats --memory
```

#### Memory Limits
```bash
# Set memory limit to prevent OOM
vittoriadb run --memory-limit 4GB

# Enable garbage collection tuning
vittoriadb run --gc-target 5
```

### SIMD Operations
```bash
# Enable SIMD optimizations (when available)
vittoriadb run --enable-simd
```

### Batch Operations
```python
# Use batch operations for better throughput
vectors = [{"id": f"doc_{i}", "vector": [...], "metadata": {...}} for i in range(1000)]
collection.insert_batch(vectors)
```

### WAL Optimization
```yaml
storage:
  wal:
    sync_interval: "1s"      # Batch writes for better performance
    checkpoint_interval: "60s"
```

## 📈 Performance Tuning Guide

### For High Insert Throughput

1. **Use Batch Operations**
   ```python
   # Instead of individual inserts
   for vector in vectors:
       collection.insert(vector)
   
   # Use batch insert
   collection.insert_batch(vectors)
   ```

2. **Optimize Index Settings**
   ```yaml
   index:
     hnsw:
       ef_construction: 100  # Lower for faster builds
   ```

3. **Disable Sync Writes (Development)**
   ```yaml
   storage:
     sync_writes: false
   ```

### For High Search Performance

1. **Use HNSW Index**
   ```bash
   # Create collection with HNSW
   curl -X POST http://localhost:8080/collections \
     -d '{"name": "fast_search", "dimensions": 384, "index_type": 1}'
   ```

2. **Optimize Search Parameters**
   ```yaml
   index:
     hnsw:
       ef_search: 100  # Higher for better accuracy
   ```

3. **Use Appropriate Batch Sizes**
   ```python
   # Search in batches for multiple queries
   results = collection.search_batch(query_vectors, limit=10)
   ```

### For Memory Efficiency

1. **Choose Appropriate Index Type**
   ```python
   # For small datasets, use flat index
   collection = db.create_collection("small", dimensions=384, index_type="flat")
   
   # For large datasets, use HNSW
   collection = db.create_collection("large", dimensions=384, index_type="hnsw")
   ```

2. **Optimize Vector Dimensions**
   ```python
   # Use appropriate dimensions for your use case
   # Higher dimensions = more memory usage
   collection = db.create_collection("docs", dimensions=384)  # Good balance
   ```

3. **Enable Compression**
   ```yaml
   storage:
     compression: true
   ```

## 🔍 Performance Monitoring

### Built-in Metrics

#### Database Statistics
```bash
curl http://localhost:8080/stats
```

**Response:**
```json
{
  "total_vectors": 100000,
  "total_size": 104857600,
  "queries_total": 1000,
  "queries_per_sec": 150.5,
  "avg_query_latency": 6.6
}
```

#### Collection Statistics
```bash
curl http://localhost:8080/collections/documents/stats
```

#### Memory Usage
```bash
# Check memory usage
ps aux | grep vittoriadb

# Or use built-in stats
vittoriadb stats --memory
```

### Performance Profiling

#### Go Profiling
```bash
# Enable profiling
vittoriadb run --profile --profile-port 6060

# Access profiling endpoints
curl http://localhost:6060/debug/pprof/
```

#### Python Client Profiling
```python
import time
import vittoriadb

# Measure operation times
start = time.time()
collection.insert_batch(vectors)
insert_time = time.time() - start

print(f"Batch insert took {insert_time:.2f}s ({len(vectors)/insert_time:.0f} vectors/sec)")
```

## 🧪 Benchmarking

### Built-in Benchmarks

#### Go Benchmarks
```bash
# Run Go benchmarks
go test ./pkg/core -bench=. -benchmem

# Run specific benchmarks
go test ./pkg/index -bench=BenchmarkHNSW -benchmem
```

#### Python Benchmarks
```bash
# Run Python performance tests
cd examples/python
python performance_benchmark.py
```

#### cURL Volume Tests
```bash
# Run cURL volume tests
cd examples/curl
./volume_test.sh
```

### Custom Benchmarks

#### Insert Performance Test
```python
import time
import numpy as np
import vittoriadb

db = vittoriadb.connect()
collection = db.create_collection("benchmark", dimensions=384)

# Generate test data
vectors = [
    {
        "id": f"vec_{i}",
        "vector": np.random.random(384).tolist(),
        "metadata": {"index": i}
    }
    for i in range(10000)
]

# Measure insert performance
start = time.time()
collection.insert_batch(vectors)
duration = time.time() - start

print(f"Inserted {len(vectors)} vectors in {duration:.2f}s")
print(f"Insert rate: {len(vectors)/duration:.0f} vectors/sec")
```

#### Search Performance Test
```python
# Measure search performance
query_vector = np.random.random(384).tolist()
num_searches = 100

start = time.time()
for _ in range(num_searches):
    results = collection.search(query_vector, limit=10)
duration = time.time() - start

print(f"Performed {num_searches} searches in {duration:.2f}s")
print(f"Search rate: {num_searches/duration:.1f} searches/sec")
print(f"Average latency: {duration/num_searches*1000:.1f}ms")
```

## 📊 Performance Comparison

### Index Type Comparison

| Metric | Flat Index | HNSW Index |
|--------|------------|------------|
| **Build Time** | Instant | Seconds to minutes |
| **Memory Usage** | Low | Higher |
| **Search Accuracy** | 100% (exact) | 95-99% (approximate) |
| **Search Speed (1K vectors)** | ~1ms | ~0.1ms |
| **Search Speed (100K vectors)** | ~100ms | ~1ms |
| **Insert Speed** | Very fast | Fast |
| **Best Use Case** | Small datasets, exact search | Large datasets, fast search |

### Distance Metric Performance

| Metric | Relative Performance | Use Case |
|--------|---------------------|----------|
| **Cosine** | 100% (baseline) | Text embeddings, normalized vectors |
| **Euclidean** | 98% | General purpose, spatial data |
| **Dot Product** | 105% | Similarity scoring, recommendation |
| **Manhattan** | 95% | High-dimensional sparse data |

## 🎯 Performance Best Practices

### Data Modeling
1. **Choose appropriate dimensions** (384-768 for most text embeddings)
2. **Normalize vectors** for cosine similarity
3. **Use meaningful metadata** for filtering
4. **Batch operations** when possible

### Index Selection
1. **Use Flat index** for <10K vectors or when exact results are required
2. **Use HNSW index** for >10K vectors and approximate search is acceptable
3. **Tune HNSW parameters** based on your accuracy/speed requirements

### System Configuration
1. **Allocate sufficient memory** (2-4x your data size)
2. **Use SSD storage** for better I/O performance
3. **Enable SIMD** if available on your platform
4. **Monitor memory usage** and adjust cache size accordingly

### Application Design
1. **Use connection pooling** for high-concurrency applications
2. **Implement proper error handling** and retries
3. **Cache frequently accessed vectors** in your application
4. **Use appropriate batch sizes** (100-1000 vectors per batch)

## 🚨 Performance Troubleshooting

### Common Performance Issues

#### Slow Inserts
```bash
# Check if using batch operations
# Enable async writes for development
vittoriadb run --async-writes

# Increase batch size
collection.insert_batch(vectors, batch_size=1000)
```

#### Slow Searches
```bash
# Check index type
curl http://localhost:8080/collections/mydata/stats

# Consider HNSW for large datasets
# Tune ef_search parameter
```

#### High Memory Usage
```bash
# Check memory stats
vittoriadb stats --memory

# Reduce cache size
vittoriadb run --cache-size 100

# Enable compression
vittoriadb run --compression
```

#### High CPU Usage
```bash
# Check concurrent operations
# Reduce max_concurrency
vittoriadb run --max-concurrency 50

# Monitor with profiling
vittoriadb run --profile
```

### Performance Debugging

#### Enable Debug Logging
```bash
vittoriadb run --log-level debug
```

#### Monitor System Resources
```bash
# Monitor CPU and memory
top -p $(pgrep vittoriadb)

# Monitor I/O
iotop -p $(pgrep vittoriadb)

# Monitor network
netstat -i
```

#### Analyze Query Patterns
```python
# Log query times
import time

def timed_search(collection, vector, limit=10):
    start = time.time()
    results = collection.search(vector, limit=limit)
    duration = time.time() - start
    print(f"Search took {duration*1000:.1f}ms, found {len(results)} results")
    return results
```
