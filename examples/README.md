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

## 🐍 Python Examples (`python/`)

The Python examples are organized in a logical progression from basic manual vector operations to advanced external service embedding features. Each file demonstrates different vectorization approaches following industry best practices.

### 📚 Learning Path (Recommended Order)

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

#### 10. Local Vectorizer Validation Test
**File:** `10_local_vectorizer_validation_test.py`

**🧪 VALIDATION TEST:** Comprehensive testing of the pure Go local vectorizer implementation:
- Local vectorizer functionality validation
- Performance measurement and analysis
- Semantic search quality testing
- Zero-dependency verification
- Comparison with other vectorizer approaches

**Usage:**
```bash
python examples/python/10_local_vectorizer_validation_test.py
```

**Features:**
- ✅ Pure Go local vectorizer testing
- ✅ Zero external dependencies validation
- ✅ Performance benchmarking (microsecond-level timing)
- ✅ Semantic search accuracy verification
- ✅ Approach comparison analysis
- ✅ Deterministic embedding validation
- ✅ Offline capability confirmation

**Test Results:**
```
✅ Inserted 5 texts in 0.005s (0.001s per text)
✅ Search time: 0.001s per query
✅ Semantic similarity scores: 0.90+ for relevant matches
✅ No Python subprocess calls
✅ No external dependencies required
```

---

## 📋 Quick Reference

### File Naming Convention
- **00-05**: Core embedding functionality (manual → client-side → server-side)
- **06-07**: RAG (Retrieval-Augmented Generation) examples
- **08-09**: Document processing and performance testing
- **10**: Validation and testing utilities

### Recommended Learning Path
1. **Start here**: `00_basic_usage_manual_vectors.py` - Learn the basics
2. **Client-side**: `01_client_side_embeddings_basic.py` - Understand automatic embeddings
3. **Server-side**: `02_server_side_embeddings_basic.py` - **🚀 NEW FEATURE!**
4. **Advanced**: `03_server_side_embeddings_advanced.py` - Deep dive testing
5. **Compare**: `04_embedding_methods_comparison.py` - See all approaches
6. **Production**: `05_production_features_showcase.py` - Enterprise features
7. **Validation**: `10_local_vectorizer_validation_test.py` - **🧪 Test local vectorizer**

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

### 📋 Go Examples Quick Reference

#### File Naming Convention
- **01**: HTTP client usage (pure vector database approach)
- **02-04**: Native SDK integration (in-process usage)
- **05**: RAG workflow (complete application example)
- **06-08**: Performance testing (volume, basic, optimized)

#### Recommended Learning Path
1. **Start here**: `01_http_client_basic_usage.go` - Learn HTTP API
2. **Native SDK**: `02_native_sdk_simple_demo.go` - Direct integration
3. **Testing**: `03_native_sdk_basic_test.go` - Basic validation
4. **Advanced**: `04_native_sdk_advanced_features.go` - Complex features
5. **RAG**: `05_rag_complete_workflow.go` - Complete application
6. **Performance**: `06_performance_volume_testing.go` - Scalability testing

## 🌐 HTTP API Examples (`curl/`)

### 🚀 Basic Usage with cURL
**File:** `curl/basic_usage.sh`

Complete HTTP API demonstration using bash and cURL:
- Connection testing and health checks
- Collection creation and management
- Individual and batch vector operations
- Similarity search with filtering
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
| **Plain Text** | `.txt` | ✅ Fully Supported | Direct text processing |
| **Markdown** | `.md` | ✅ Fully Supported | Frontmatter parsing |
| **HTML** | `.html` | ✅ Fully Supported | Tag stripping, metadata |
| **PDF** | `.pdf` | ✅ Fully Supported | Multi-page text extraction |
| **DOCX** | `.docx` | ✅ Fully Supported | Properties, text extraction |
| **DOC** | `.doc` | 🚧 Placeholder | Legacy format support |
| **RTF** | `.rtf` | ❌ Not Implemented | Rich text format |

## 🚀 Quick Start

1. **Start VittoriaDB:**
   ```bash
   ./vittoriadb run
   ```

2. **Choose an example:**
   - **Start learning**: `python examples/python/00_basic_usage_manual_vectors.py`
   - **Server-side embeddings**: `python examples/python/02_server_side_embeddings_basic.py` **🚀 NEW!**
   - **Test local vectorizer**: `python examples/python/10_local_vectorizer_validation_test.py` **🧪 VALIDATE!**
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
