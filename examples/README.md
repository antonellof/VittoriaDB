# VittoriaDB Examples

This directory contains comprehensive examples demonstrating VittoriaDB's capabilities across different use cases and programming languages, organized by language and interface type.

> **📦 All Python examples use the centralized VittoriaDB Python library** located in `../sdk/python/` for consistent API usage and better maintainability.

## 📁 Directory Structure

```
examples/
├── python/          # Python client examples
├── go/             # Go native examples  
├── curl/           # HTTP API examples with bash/curl
├── documents/      # Sample documents for testing
└── README.md       # This file
```

### Operations (backup, metrics, filters, IVF)

These capabilities complement the tutorial scripts:

- **Backup / restore**: Use the `vittoriadb` CLI (not HTTP). Stop the server before restore. Example: `./vittoriadb backup --data-dir ./data --output ./backup.tar.gz` and `./vittoriadb restore --data-dir ./restored --input ./backup.tar.gz`.
- **Prometheus**: `GET /metrics` exposes search counters, mean latency, and approximate QPS since process start.
- **Metadata filters**: Search accepts a **flat map** (`{"category": "tech"}`) or a **structured** filter tree; see `examples/python/15_observability_metrics_ivf.py` for both and for `Collection.search(..., use_post=True)`.
- **IVF index**: Collection field `index_type` / `IndexType.IVF` with `config` keys `nlist` and `nprobe` (prototype). Helper in Python: `Configure.Index.ivf_flat(...)`.

## 🐍 Python Examples (`python/`)

The Python examples are organized in a logical progression from basic manual vector operations to advanced external service embedding features. Each file demonstrates different vectorization approaches following industry best practices.

### 📚 Learning Path (Recommended Order)

#### 15. Observability and IVF (metrics, filters, IVF index)
**File:** `15_observability_metrics_ivf.py`

Prometheus `/metrics`, `/stats` query fields, IVF collection creation, and metadata filters (flat map vs structured body via POST search).

**Usage:**
```bash
# With server: ./vittoriadb run --data-dir ./data
python examples/python/15_observability_metrics_ivf.py
```

**Features:**
- IVF index with `Configure.Index.ivf_flat`
- Flat and structured metadata filters
- `db.prometheus_metrics()` and search-related stats

---

#### 00. Basic Manual Vector Operations
**File:** `00_basic_usage_manual_vectors.py`

Introduction to VittoriaDB with manual vector handling:
- Connection management and health checks
- Collection creation and management
- Manual vector insertion (individual and batch)
- Similarity search with metadata filtering
- Database statistics and cleanup

**Usage:**
```bash
python examples/python/00_basic_usage_manual_vectors.py
```

**Features:**
- ✅ Complete workflow demonstration
- ✅ Manual vector operations
- ✅ Metadata filtering examples
- ✅ Proper cleanup and connection management

#### 01. Client-Side Automatic Embeddings (Basic)
**File:** `01_client_side_embeddings_basic.py`

Introduction to automatic embedding generation on the client side:
- Client-side sentence-transformers integration
- Automatic text-to-vector conversion
- Basic semantic search capabilities
- Performance comparison with manual approaches

**Usage:**
```bash
pip install sentence-transformers
python examples/python/01_client_side_embeddings_basic.py
```

**Features:**
- ✅ Client-side automatic text vectorization
- ✅ Sentence transformers integration
- ✅ Basic semantic search demonstration
- ✅ Performance analysis

#### 02. Server-Side Automatic Embeddings (Basic)
**File:** `02_server_side_embeddings_basic.py`

**🚀 NEW FEATURE:** Server-side automatic text vectorization using `Configure.Vectors.auto_embeddings()`:
- Zero client-side dependencies (no sentence-transformers required)
- Automatic text-to-vector conversion on the server
- Semantic search with server-side query vectorization
- Production-ready embedding generation

**Usage:**
```bash
python examples/python/02_server_side_embeddings_basic.py
```

**Features:**
- ✅ No client-side model loading required
- ✅ Consistent embeddings across all clients
- ✅ Centralized model management
- ✅ Zero configuration automatic embeddings

#### 03. Server-Side Automatic Embeddings (Advanced)
**File:** `03_server_side_embeddings_advanced.py`

Advanced testing of server-side embedding functionality:
- Comprehensive API testing (single, batch, search)
- Performance benchmarking and analysis
- Error handling and validation testing
- Quality assurance for semantic search

**Usage:**
```bash
python examples/python/03_server_side_embeddings_advanced.py
```

**Features:**
- ✅ Full server-side embedding API testing
- ✅ Performance benchmarking (5-6s per operation)
- ✅ Semantic search accuracy validation (0.74+ scores)
- ✅ Batch processing efficiency analysis (4x faster)
- ✅ Comprehensive error handling validation

#### 04. Embedding Methods Comparison
**File:** `04_embedding_methods_comparison.py`

Side-by-side comparison of all embedding approaches:
- **Manual embeddings** (traditional approach)
- **Client-side automatic** (using sentence-transformers)
- **Server-side automatic** (new VittoriaDB feature)
- Performance analysis and winner determination

**Usage:**
```bash
python examples/python/04_embedding_methods_comparison.py
```

**Features:**
- ✅ Side-by-side comparison of all approaches
- ✅ Performance analysis and timing comparisons
- ✅ Semantic search quality demonstration
- ✅ Clear winner analysis (server-side automatic!)

#### 05. Production Features Showcase
**File:** `05_production_features_showcase.py`

Comprehensive demonstration of production-ready features:
- Multiple vectorizer types (Sentence Transformers, OpenAI)
- Document processing with automatic embeddings
- Performance analysis and scalability testing
- Complete API coverage and error handling

**Usage:**
```bash
python examples/python/05_production_features_showcase.py
```

**Features:**
- ✅ Production-grade feature demonstration
- ✅ Multiple vectorizer backend support
- ✅ Complete API endpoint coverage
- ✅ Enterprise-ready error handling
- ✅ Performance and scalability analysis

### 🤖 RAG (Retrieval-Augmented Generation) Examples

#### 06. RAG Basic Example
**File:** `06_rag_basic_example.py`

Basic RAG system implementation:
- Simple document ingestion and processing
- Vector embedding generation
- Basic semantic search and retrieval
- Simple context-aware responses

**Usage:**
```bash
pip install sentence-transformers
python examples/python/06_rag_basic_example.py
```

**Features:**
- ✅ Basic RAG pipeline implementation
- ✅ Simple document processing
- ✅ Basic semantic search
- ✅ Context retrieval and response generation

#### 07. RAG Complete Workflow
**File:** `07_rag_complete_workflow.py`

A comprehensive RAG system implementation showing:
- Document ingestion and processing
- Vector embedding generation using Sentence Transformers
- Semantic search and retrieval
- Context-aware response generation
- Interactive demo with sample knowledge base

**Requirements:**
```bash
pip install sentence-transformers numpy
```

**Usage:**
```bash
pip install sentence-transformers
python examples/python/07_rag_complete_workflow.py
```

**Features:**
- ✅ Complete RAG pipeline using centralized Python library
- ✅ Sentence Transformer embeddings (384 dimensions)
- ✅ Interactive Q&A system with graceful error handling
- ✅ Sample knowledge base about VittoriaDB
- ✅ Semantic search with confidence scoring
- ✅ Context-aware response generation
- ✅ Automatic duplicate handling and retry logic
- ✅ Professional error handling with detailed feedback

---

### 📄 Document Processing and Performance Examples

#### 08. Document Processing Workflow
**File:** `08_document_processing_workflow.py`

Demonstrates VittoriaDB's document processing capabilities:
- Processing various formats (TXT, MD, HTML, PDF, DOCX)
- Document upload and chunking
- Metadata extraction and preservation
- Collection management

**Requirements:**
```bash
pip install numpy
```

**Usage:**
```bash
# Start VittoriaDB server
./vittoriadb run

# Run the document processing example
python examples/python/08_document_processing_workflow.py
```

**Features:**
- ✅ Multi-format document processing using centralized library
- ✅ Intelligent text chunking with configurable sizes
- ✅ Metadata extraction and preservation
- ✅ Automatic sample document creation (TXT, MD, HTML)
- ✅ Collection statistics and information display
- ✅ Graceful error handling and connection management

---

#### 09. Performance Benchmarks
**File:** `09_performance_benchmarks.py`

Comprehensive performance testing suite:
- Insert performance (individual vs batch)
- Search performance comparison
- Memory usage monitoring
- Index type comparison (Flat vs HNSW)

**Requirements:**
```bash
pip install numpy psutil
```

**Usage:**
```bash
# Start VittoriaDB server
./vittoriadb run

# Run benchmarks
python examples/python/09_performance_benchmarks.py
```

**Features:**
- ✅ Comprehensive performance metrics using centralized library
- ✅ Individual vector operations with timing
- ✅ Memory usage tracking and reporting
- ✅ Multiple distance metrics comparison (cosine, euclidean)
- ✅ Detailed performance reports with statistics
- ✅ Automatic collection cleanup after tests

#### 11. All Vectorizers Comparison
**File:** `11_all_vectorizers_comparison.py`

**🧪 STATUS MATRIX:** Probe every supported vectorizer (Ollama, auto-embeddings, Sentence
Transformers, OpenAI, HuggingFace) and print a status matrix showing which ones are reachable in
the current environment. The HuggingFace path now hits the real Inference API (added in v0.6.0).

**Usage:**
```bash
python examples/python/11_all_vectorizers_comparison.py
```

#### 12. OpenAI API Testing
**File:** `12_openai_api_testing.py`

Smoke-tests the OpenAI vectorizer integration. Requires `OPENAI_API_KEY`.

#### 13. Vectorization Comparison Test
**File:** `13_vectorization_comparison_test.py`

Side-by-side semantic-quality analysis across manual vectors, sentence-transformers, and OpenAI
embeddings.

#### 14. Simple Vectorization Test
**File:** `14_simple_vectorization_test.py`

Zero-dependency curl-based smoke test that exercises the HTTP API and the deterministic
fallback vectorization path.

---

## 📋 Quick Reference

### File Naming Convention
- **00-05**: Core embedding functionality (manual → client-side → server-side)
- **06-07**: RAG (Retrieval-Augmented Generation) examples
- **08-09**: Document processing and performance testing
- **11-14**: Vectorizer comparison & smoke tests

### Recommended Learning Path
1. **Start here**: `00_basic_usage_manual_vectors.py` - Learn the basics
2. **Client-side**: `01_client_side_embeddings_basic.py` - Understand automatic embeddings
3. **Server-side**: `02_server_side_embeddings_basic.py` - **🚀 Server-side auto-embeddings**
4. **Advanced**: `03_server_side_embeddings_advanced.py` - Deep dive testing
5. **Compare**: `04_embedding_methods_comparison.py` - See all approaches
6. **Production**: `05_production_features_showcase.py` - Enterprise features
7. **Vectorizer Status**: `11_all_vectorizers_comparison.py` - **🧪 Status matrix**
8. **Quality Analysis**: `13_vectorization_comparison_test.py` - **🔬 Advanced analysis**
9. **Smoke Test**: `14_simple_vectorization_test.py` - **🚀 Zero-dependency check**

#### 13. Vectorization Comparison Test (NEW!)
**File:** `13_vectorization_comparison_test.py`

**🔬 ADVANCED TESTING:** Compare different vectorization approaches with comprehensive analysis:
- Manual vs sentence-transformers vs OpenAI embeddings
- Similarity score analysis and discrimination testing
- Large text processing with proper matching
- Backend-style implementation comparison
- Production readiness assessment

**Usage:**
```bash
pip install requests numpy sentence-transformers openai
cd examples/python
python 13_vectorization_comparison_test.py
```

**Features:**
- ✅ **Multiple approaches** - Manual, sentence-transformers, OpenAI embeddings
- ✅ **Quality analysis** - Semantic discrimination and accuracy testing
- ✅ **Large text support** - Real document processing and chunking
- ✅ **Backend comparison** - Compare with web-ui-rag implementation
- ✅ **Production guidance** - Recommendations for different use cases

#### 14. Simple Vectorization Test (NEW!)
**File:** `14_simple_vectorization_test.py`

**🚀 NO DEPENDENCIES:** Test vectorization using only Python standard library:
- Basic hash-based vector generation
- HTTP API testing with curl
- Similarity analysis without external libraries
- Backend comparison and recommendations
- Development/testing suitability assessment

**Usage:**
```bash
# No pip install required - uses only standard library + curl
cd examples/python
python 14_simple_vectorization_test.py
```

**Features:**
- ✅ **Zero dependencies** - Only Python standard library and curl
- ✅ **Quick testing** - Fast setup for development environments
- ✅ **API validation** - Test HTTP API endpoints and formats
- ✅ **Basic analysis** - Simple similarity and discrimination testing
- ✅ **Clear limitations** - Honest assessment of manual vector quality

**⚠️ Test Results:**
Both new tests confirm the issue identified in Go examples:
- **Manual vectors**: Poor semantic discrimination (33% accuracy, 0.004 discrimination score)
- **Sentence-transformers**: Good semantic understanding and proper matching
- **Recommendation**: Use sentence-transformers or OpenAI embeddings for production

## 🔧 Go Examples (`go/`)

The Go examples demonstrate both **HTTP client usage** and **native SDK integration**. They are organized in a logical progression from basic HTTP operations to advanced native SDK features and performance testing.

### 📚 Learning Path (Recommended Order)

#### 01. HTTP Client Basic Usage
**File:** `01_http_client_basic_usage.go`

Complete HTTP client implementation demonstrating VittoriaDB as a **pure vector database**:
- HTTP client with connection management
- Collection creation and management
- Manual vector insertion (individual and batch)
- Similarity search with metadata filtering
- Performance comparison and benchmarking
- Comprehensive error handling

**Usage:**
```bash
cd examples/go
go run 01_http_client_basic_usage.go
```

**Features:**
- ✅ Complete HTTP API client implementation
- ✅ Connection testing and health checks
- ✅ Manual vector operations (Approach 3: Pure Vector DB)
- ✅ Filtered search with metadata
- ✅ Performance benchmarking
- ✅ Comprehensive error handling

#### 02. Native SDK Simple Demo
**File:** `02_native_sdk_simple_demo.go`

Direct usage of VittoriaDB native Go SDK components:
- Direct index API usage (Flat and HNSW)
- In-process vector operations
- Performance comparisons between index types
- Memory usage statistics

**Usage:**
```bash
cd examples/go
go run 02_native_sdk_simple_demo.go
```

**Features:**
- ✅ Native Go SDK integration
- ✅ Direct index API usage
- ✅ Performance measurements
- ✅ Index type comparison (Flat vs HNSW)
- ✅ Memory usage statistics

#### 03. Native SDK Basic Test
**File:** `03_native_sdk_basic_test.go`

Basic functionality testing with native SDK:
- Core operations validation
- Simple performance tests
- Basic error scenarios
- SDK integration patterns

**Usage:**
```bash
cd examples/go
go run 03_native_sdk_basic_test.go
```

**Features:**
- ✅ Native SDK basic operations
- ✅ Core functionality validation
- ✅ Simple performance tests
- ✅ Error handling patterns

#### 04. Native SDK Advanced Features
**File:** `04_native_sdk_advanced_features.go`

Advanced VittoriaDB functionality testing with native SDK:
- Complex vector operations
- Advanced indexing features
- Error handling scenarios
- Performance edge cases
- Advanced configuration options

**Usage:**
```bash
cd examples/go
go run 04_native_sdk_advanced_features.go
```

**Features:**
- ✅ Advanced native SDK features
- ✅ Complex vector operations
- ✅ Advanced indexing capabilities
- ✅ Edge case handling
- ✅ Performance optimization

#### 05. RAG Complete Workflow
**File:** `05_rag_complete_workflow.go`

Complete RAG system implementation using HTTP client:
- Knowledge base creation and management
- Document chunking and processing
- Manual embedding generation (client-side)
- Semantic search and retrieval
- Answer generation from context
- Interactive Q&A system

**Usage:**
```bash
cd examples/go
go run 05_rag_complete_workflow.go
```

**Features:**
- ✅ End-to-end RAG system implementation
- ✅ Document chunking strategies
- ✅ Client-side embedding generation (Approach 3: Pure Vector DB)
- ✅ Context-aware answer generation
- ✅ Interactive query system
- ✅ Performance analysis and optimization

### 📊 Performance Testing Examples

#### 06. Performance Volume Testing
**File:** `06_performance_volume_testing.go`

Comprehensive performance testing with different data volumes:
- Multi-scale testing (KB, MB, GB scale)
- Index type comparison (Flat vs HNSW)
- Memory usage monitoring
- Throughput analysis across different scales

**Usage:**
```bash
cd examples/go
go run 06_performance_volume_testing.go
```

**Features:**
- ✅ Multi-scale performance testing
- ✅ Index type optimization analysis
- ✅ Memory usage tracking
- ✅ Throughput measurements
- ✅ Scalability analysis

#### 07. Performance Benchmarks (Basic)
**File:** `07_performance_benchmarks_basic.go`

Basic performance benchmarking suite:
- Insert performance (individual vs batch)
- Search performance comparison
- Memory usage monitoring
- Basic optimization patterns

**Usage:**
```bash
cd examples/go
go run 07_performance_benchmarks_basic.go
```

**Features:**
- ✅ Basic performance metrics
- ✅ Insert/search benchmarking
- ✅ Memory usage analysis
- ✅ Performance comparison patterns

#### 08. Performance Benchmarks (Optimized)
**File:** `08_performance_benchmarks_optimized.go`

Optimized performance benchmarking with advanced techniques:
- Optimized batch operations
- Advanced HNSW parameter tuning
- Memory optimization strategies
- High-throughput patterns
- Production-grade performance testing

**Usage:**
```bash
cd examples/go
go run 08_performance_benchmarks_optimized.go
```

**Features:**
- ✅ Optimized performance patterns
- ✅ Advanced HNSW tuning
- ✅ Memory optimization
- ✅ High-throughput testing
- ✅ Production-grade benchmarking

#### 09. Parallel Search & Caching Demo
**File:** `09_parallel_search_and_caching_demo.go`

**🚀 NEW FEATURE:** Comprehensive demonstration of VittoriaDB's enhanced parallel search and caching capabilities:
- Enhanced batch processing with error recovery
- Parallel search engine with configurable workers
- Intelligent caching with LRU eviction and TTL
- Performance statistics and monitoring
- Cache management and optimization

**Usage:**
```bash
cd examples/go
go run 09_parallel_search_and_caching_demo.go
```

**Features:**
- ✅ **165x+ faster cached searches** - Dramatic performance improvements
- ✅ **1M+ vectors/sec insertion** - High-throughput batch processing
- ✅ **Parallel search processing** - Configurable worker pools
- ✅ **Intelligent caching** - LRU eviction with TTL expiration
- ✅ **Performance monitoring** - Real-time statistics and metrics
- ✅ **Cache management** - Manual cache control and optimization
- ✅ **Error recovery** - Robust fallback strategies
- ✅ **Production-ready** - Enterprise-grade reliability features

#### 12. Parallel vs Sequential Benchmark (NEW!)
**File:** `12_parallel_vs_sequential_benchmark.go`

**📊 PERFORMANCE ANALYSIS:** Comprehensive benchmark comparing parallel search and caching against sequential search:
- Sequential vs parallel search performance comparison
- Cache effectiveness analysis with repeated queries
- Overhead analysis for small datasets
- Performance scaling with different dataset sizes
- Real-world performance metrics and recommendations

**Usage:**
```bash
cd examples/go
go run 12_parallel_vs_sequential_benchmark.go
```

**Features:**
- ✅ **Proven performance gains** - 5-32x speedup for large datasets (1000+ vectors)
- ✅ **Cache analysis** - 10-276x speedup for repeated queries
- ✅ **Smart thresholding** - Avoids overhead for small datasets (<1000 vectors)
- ✅ **Realistic benchmarks** - Tests with 100-10,000 vector datasets
- ✅ **Overhead analysis** - Measures parallel search overhead (1.4x for small datasets)
- ✅ **Performance recommendations** - Data-driven guidance for optimization

**📈 Benchmark Results:**
- **Small datasets (100-500 vectors)**: Sequential preferred (lower overhead)
- **Medium datasets (1000 vectors)**: 5x parallel speedup + 20x cache speedup
- **Large datasets (5000+ vectors)**: 15-32x parallel speedup + 125-276x cache speedup
- **Cache hit rate**: 99% for repeated queries

#### 13. Unified Configuration Demo (NEW!)
**File:** `13_unified_configuration_demo.go`

**🔧 CONFIGURATION MANAGEMENT:** Comprehensive demonstration of VittoriaDB's unified configuration system:
- Multiple configuration sources (YAML, environment variables, CLI flags)
- Configuration validation with performance and security checks
- Hot-reloading and dynamic configuration updates
- Legacy configuration migration utilities
- Production-ready configuration management

**Usage:**
```bash
cd examples/go
go run 13_unified_configuration_demo.go
```

**Features:**
- ✅ **Multi-source configuration** - YAML files, environment variables, CLI flags
- ✅ **Validation system** - Performance, security, and resource validators
- ✅ **Hot-reloading** - Dynamic configuration updates without restart
- ✅ **Migration utilities** - Seamless legacy configuration migration
- ✅ **CLI tools integration** - Works with `vittoriadb config` commands
- ✅ **Production scenarios** - Development, production, and high-performance configurations

#### 14. I/O Optimization Demo (NEW!)
**File:** `14_io_optimization_demo.go`

**⚡ PERFORMANCE OPTIMIZATION:** Advanced I/O optimization features including chunked vectorized math, memory-mapped storage, and async I/O:
- Chunked, allocation-friendly vector ops in pure Go (true CPU SIMD intrinsics planned)
- Memory-mapped storage with zero-copy operations
- Async I/O engine with worker pools and batching
- Comprehensive performance benchmarks across different scenarios
- Real-world performance analysis and system information

**Usage:**
```bash
cd examples/go
go run 14_io_optimization_demo.go
```

**Features:**
- ✅ **Chunked vector ops** - Up to 7.7x speedup for batched math on large datasets
- ✅ **Memory-mapped I/O** - Zero-copy reads, 10-50x faster than traditional I/O
- ✅ **Async I/O engine** - Non-blocking operations with worker pools
- ✅ **Vectorized operations** - Better CPU cache utilization and performance
- ✅ **Comprehensive benchmarks** - Performance analysis across dimensions and dataset sizes
- ✅ **System integration** - Works seamlessly with unified configuration system

**📊 Performance Results:**
- **Vectorized math**: 1.03-2.02x speedup for individual operations
- **Parallel batched ops**: 2.4-7.7x speedup for large datasets (10K+ vectors)
- **Memory-mapped storage**: Zero-copy reads with microsecond-level latency
- **Async I/O throughput**: Improved concurrent operation handling
- **System scalability**: Optimal performance across different CPU core counts

#### 10. Large Text Processing Demo
**File:** `10_large_text_processing_demo.go`

**📚 REAL-WORLD USAGE:** Process large text files with proper semantic matching and relevance filtering:
- Large text file processing with smart chunking
- Semantic search with proper matching results
- Relevance scoring and filtering
- Native vs manual vectorization comparison
- Production-ready text processing pipeline

**Usage:**
```bash
cd examples/go
go run 10_large_text_processing_demo.go
```

**Features:**
- ✅ **Large file processing** - Handle README files, documentation, books
- ✅ **Smart chunking** - Sentence-boundary aware text splitting
- ✅ **Proper matching** - Only return relevant results above threshold
- ✅ **Native vectorization** - Uses sentence-transformers when available
- ✅ **Fallback system** - Enhanced manual vectors when native unavailable
- ✅ **Relevance filtering** - Score-based result filtering
- ✅ **Real-world testing** - Actual documentation and text files

#### 11. Similarity Debugging Tool
**File:** `11_similarity_debugging_tool.go`

**🔍 DEBUGGING TOOL:** Analyze and debug vector generation and similarity calculations:
- Vector generation quality analysis
- Similarity score validation
- Score distribution analysis
- Real-world text pair testing
- Vector algorithm debugging

**Usage:**
```bash
cd examples/go
go run 11_similarity_debugging_tool.go
```

**Features:**
- ✅ **Vector validation** - Test with known orthogonal/identical vectors
- ✅ **Generation analysis** - Analyze vector generation quality
- ✅ **Score distribution** - Statistical analysis of similarity scores
- ✅ **Real-world testing** - Test with actual text pairs
- ✅ **Problem identification** - Identify vector generation issues
- ✅ **Algorithm debugging** - Debug similarity calculation problems

**⚠️ Current Issue Identified:**
The manual vector generation creates vectors that are too similar (0.99+ scores even for unrelated topics). For production use:
- **Recommended**: Use native vectorization with `pip install sentence-transformers`
- **Alternative**: Improve manual vector generation algorithm for better diversity

### 📋 Go Examples Quick Reference

#### File Naming Convention
- **01**: HTTP client usage (pure vector database approach)
- **02-04**: Native SDK integration (in-process usage)
- **05**: RAG workflow (complete application example)
- **06-08**: Performance testing (volume, basic, optimized)
- **09**: Parallel search and caching features (NEW!)
- **10**: Large text processing with proper matching (NEW!)
- **11**: Similarity debugging and analysis tool (NEW!)

#### Recommended Learning Path
1. **Start here**: `01_http_client_basic_usage.go` - Learn HTTP API
2. **Native SDK**: `02_native_sdk_simple_demo.go` - Direct integration
3. **Testing**: `03_native_sdk_basic_test.go` - Basic validation
4. **Advanced**: `04_native_sdk_advanced_features.go` - Complex features
5. **RAG**: `05_rag_complete_workflow.go` - Complete application
6. **Performance**: `06_performance_volume_testing.go` - Scalability testing
7. **NEW Features**: `09_parallel_search_and_caching_demo.go` - Enhanced performance 🚀
8. **Large Text**: `10_large_text_processing_demo.go` - Real-world text processing 📚
9. **Debug Tool**: `11_similarity_debugging_tool.go` - Vector analysis and debugging 🔍

## 🌐 HTTP API Examples (`curl/`)

### 🚀 Basic Usage with cURL
**File:** `curl/basic_usage.sh`

Complete HTTP API demonstration using bash and cURL:
- Connection testing, health checks, **`GET /metrics`** (Prometheus text)
- Collection creation and management
- Individual and batch vector operations
- Similarity search with filtering (flat `filter` objects in JSON bodies)
- Performance comparison and analysis
- Comprehensive error handling

**Usage:**
```bash
cd examples/curl
chmod +x basic_usage.sh
./basic_usage.sh
```

**Features:**
- ✅ Complete HTTP API workflow
- ✅ Colored output and progress indicators
- ✅ JSON validation and pretty printing
- ✅ Performance timing and metrics
- ✅ Automatic cleanup and error handling
- ✅ Cross-platform bash compatibility

---

### 📊 Volume Testing
**File:** `curl/volume_test.sh`

Comprehensive performance testing with different data volumes:
- **KB-scale testing:** Small vectors (32D, 100 vectors)
- **MB-scale testing:** Medium vectors (256D, 1K vectors)  
- **GB-scale testing:** Large vectors (512D, 5K vectors)
- Index type comparison (Flat vs HNSW)
- Memory usage monitoring and analysis
- Performance benchmarking across scales

**Usage:**
```bash
cd examples/curl
chmod +x volume_test.sh
./volume_test.sh
```

**Features:**
- ✅ Multi-scale performance testing
- ✅ Index type optimization analysis
- ✅ Memory usage tracking
- ✅ HNSW parameter tuning
- ✅ Stress testing and resource monitoring
- ✅ Detailed performance reports

---

### 🤖 RAG System with cURL
**File:** `curl/rag_example.sh`

Complete RAG system implementation using HTTP API:
- Knowledge base creation and document ingestion
- Text embedding generation (simplified for demo)
- Semantic search and information retrieval
- Context-aware answer generation
- Interactive Q&A system
- Performance analysis and optimization

**Usage:**
```bash
cd examples/curl
chmod +x rag_example.sh
./rag_example.sh
```

**Features:**
- ✅ End-to-end RAG system via HTTP API
- ✅ Document chunking and processing
- ✅ Category-based filtering and search
- ✅ Interactive demo mode
- ✅ Performance metrics and analysis
- ✅ Multi-topic knowledge base

**Requirements:**
- `bash` (4.0+)
- `curl` (for HTTP requests)
- `jq` (recommended for JSON processing)
- `bc` (for mathematical calculations)

**Installation:**
```bash
# macOS
brew install jq bc

# Ubuntu/Debian
sudo apt-get install jq bc
```

## 📁 Document Samples

The `documents/` directory contains sample documents for testing:

### Sample Files
- `sample.txt` - Plain text document about VittoriaDB
- `sample.md` - Markdown document with frontmatter
- `sample.html` - HTML document with metadata
- `test/simple.docx` - DOCX document with properties
- `test/simple.pdf` - PDF document (basic format)

### Document Types Supported
| Format | Extension | Status | Features |
|--------|-----------|---------|----------|
| **Plain Text** | `.txt`, `.text` | ✅ Fully Supported | Direct text processing |
| **Markdown** | `.md`, `.markdown` | ✅ Fully Supported | Frontmatter parsing |
| **HTML** | `.html`, `.htm` | ✅ Fully Supported | Tag stripping, metadata |
| **PDF** | `.pdf` | ✅ Fully Supported | Multi-page text extraction (`github.com/ledongthuc/pdf`) |
| **DOCX** | `.docx` | ✅ Fully Supported | Office Open XML parsing (stdlib zip + xml) |

## 🚀 Quick Start

1. **Start VittoriaDB:**
   ```bash
   ./vittoriadb run
   ```

2. **Choose an example:**
   - **Start learning**: `python examples/python/00_basic_usage_manual_vectors.py`
   - **Server-side embeddings**: `python examples/python/02_server_side_embeddings_basic.py`
   - **Vectorizer status matrix**: `python examples/python/11_all_vectorizers_comparison.py`
   - **RAG applications**: `python examples/python/07_rag_complete_workflow.py`
   - **Document processing**: `python examples/python/08_document_processing_workflow.py`
   - **Performance testing**: `python examples/python/09_performance_benchmarks.py`
   - **Go development**: `cd examples/go && go run 01_http_client_basic_usage.go`
   - **HTTP API testing**: `cd examples/curl && ./basic_usage.sh`

3. **Explore the web dashboard:**
   Open http://localhost:8080 in your browser

## 📋 Requirements

### System Requirements
- **VittoriaDB**: Download from [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases)
- **Python**: 3.7+ (for Python examples)
- **Go**: 1.21+ (for Go examples)

### Python Dependencies

> **📦 All Python examples use the VittoriaDB Python library from PyPI**

```bash
# Core library (required for all examples)
pip install vittoriadb

# Additional dependencies for specific examples
pip install numpy                    # For basic usage and performance examples
pip install sentence-transformers    # For RAG and embedding examples
pip install psutil                   # For performance benchmarks
pip install openai                   # Optional: for OpenAI embedding examples
```

### Library Installation

> **📦 Install VittoriaDB Python library from PyPI** for the best experience:

```bash
# Install from PyPI (recommended)
pip install vittoriadb

# Verify installation
python -c "import vittoriadb; print('✅ VittoriaDB Python library ready!')"
```

All examples use standard imports:
```python
import vittoriadb

# Connect to existing server (no auto-start)
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
```

**Benefits of PyPI installation:**
- ✅ Clean, simple installation process
- ✅ Automatic dependency management
- ✅ Professional production-ready setup
- ✅ IDE autocomplete and type hints work properly
- ✅ Consistent API across all examples
- ✅ Regular updates through pip

### Go Dependencies
All Go dependencies are managed via `go.mod` and will be downloaded automatically.

## 💡 Usage Tips

### Performance Optimization
- Use **HNSW indexing** for large datasets (>10k vectors)
- Use **batch operations** for better throughput
- Configure **chunk sizes** based on your content type
- Monitor **memory usage** during large operations

### RAG Best Practices
- Choose appropriate **embedding models** for your domain
- Optimize **chunk sizes** for your use case (300-800 characters)
- Use **metadata filtering** to improve relevance
- Implement **context ranking** for better responses

### Document Processing
- **PDF**: Works best with text-based PDFs
- **DOCX**: Extracts text and document properties
- **HTML**: Strips tags and preserves structure
- **Markdown**: Parses frontmatter metadata

## 🔧 Troubleshooting

### Common Issues

**Connection Error:**
```
❌ Failed to connect to VittoriaDB
```
**Solution:** Start VittoriaDB with `./vittoriadb run` and ensure port 8080 is available

**Import Error:**
```
ModuleNotFoundError: No module named 'vittoriadb'
```
**Solution:** Install the Python library in development mode: `cd sdk/python && ./install-dev.sh`

**Dependency Error:**
```
ModuleNotFoundError: No module named 'sentence_transformers'
```
**Solution:** Install dependencies with `pip install sentence-transformers numpy`

**Collection Exists Error:**
```
❌ Collection already exists
```
**Solution:** Examples now handle this gracefully with automatic conflict resolution

**Memory Issues:**
```
Out of memory during large operations
```
**Solution:** Reduce batch sizes or use streaming operations

**Enum Value Error:**
```
❌ 0 is not a valid DistanceMetric
```
**Solution:** Updated - enums now properly handle integer values from Go server

### Getting Help

1. Check the [main README](../README.md) for setup instructions
2. Visit the web dashboard at http://localhost:8080
3. Review the API documentation in the dashboard
4. Check [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues) for known problems

## 🎯 Next Steps

After running the examples:

1. **Integrate with your data**: Replace sample documents with your own
2. **Choose embedding models**: Select models appropriate for your domain
3. **Optimize performance**: Tune chunk sizes and index parameters
4. **Build applications**: Use VittoriaDB in your AI/ML projects
5. **Deploy**: Use the single binary for easy deployment

---

**🚀 Happy building with VittoriaDB!**
