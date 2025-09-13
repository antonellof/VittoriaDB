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

### 🤖 RAG (Retrieval-Augmented Generation) Complete Example
**File:** `python/rag_complete_example.py`

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
# Start VittoriaDB server
./vittoriadb run

# Install Python library in development mode (one-time setup)
cd sdk/python && ./install-dev.sh

# Run the RAG example (uses centralized Python library)
python examples/python/rag_complete_example.py
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

### 📄 Document Processing Example
**File:** `python/document_processing_example.py`

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

# Run the document processing example (uses centralized Python library)
python examples/python/document_processing_example.py
```

**Features:**
- ✅ Multi-format document processing using centralized library
- ✅ Intelligent text chunking with configurable sizes
- ✅ Metadata extraction and preservation
- ✅ Automatic sample document creation (TXT, MD, HTML)
- ✅ Collection statistics and information display
- ✅ Graceful error handling and connection management

---

### 📊 Performance Benchmark
**File:** `python/performance_benchmark.py`

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

# Run benchmarks (uses centralized Python library)
python examples/python/performance_benchmark.py
```

**Features:**
- ✅ Comprehensive performance metrics using centralized library
- ✅ Individual vector operations with timing
- ✅ Memory usage tracking and reporting
- ✅ Multiple distance metrics comparison (cosine, euclidean)
- ✅ Detailed performance reports with statistics
- ✅ Automatic collection cleanup after tests

---

### 🔍 Basic Usage Example
**File:** `python/basic_usage.py`

Simple introduction to VittoriaDB centralized Python library:
- Connection management with auto-retry
- Collection operations with error handling
- Vector insertion (individual and batch)
- Similarity search with metadata filtering
- Database statistics and cleanup

**Requirements:**
```bash
pip install numpy
```

**Usage:**
```bash
# Start VittoriaDB server
./vittoriadb run

# Run basic usage example (uses centralized Python library)
python examples/python/basic_usage.py
```

**Features:**
- ✅ Complete workflow demonstration using centralized library
- ✅ Automatic collection creation with conflict handling
- ✅ Individual and batch vector operations
- ✅ Metadata filtering and search examples
- ✅ Database statistics and collection management
- ✅ Proper cleanup and connection closing

---

### 🏗️ RAG Application Example
**File:** `python/rag_example.py`

Advanced RAG application using centralized Python library:
- Document chunking strategies
- Embedding model integration (Sentence Transformers)
- Query processing and semantic search
- Response generation with context

**Requirements:**
```bash
pip install sentence-transformers numpy
```

**Usage:**
```bash
# Start VittoriaDB server
./vittoriadb run

# Run RAG application example (uses centralized Python library)
python examples/python/rag_example.py
```

**Features:**
- ✅ Advanced RAG implementation using centralized library
- ✅ Optional Sentence Transformers integration
- ✅ Fallback to random embeddings if transformers unavailable
- ✅ Document chunking and processing
- ✅ Interactive query system
- ✅ Graceful error handling and connection management

## 🔧 Go Examples (`go/`)

### 🚀 Basic Usage Example
**File:** `go/basic_usage.go`

Complete HTTP client implementation in Go:
- VittoriaDB HTTP client with connection management
- Collection creation and management
- Individual and batch vector operations
- Similarity search with metadata filtering
- Performance comparison and benchmarking
- Error handling and cleanup

**Usage:**
```bash
cd examples/go
go run basic_usage.go
```

**Features:**
- ✅ Complete HTTP API client implementation
- ✅ Connection testing and health checks
- ✅ Individual and batch vector operations
- ✅ Filtered search with metadata
- ✅ Performance benchmarking
- ✅ Comprehensive error handling

---

### 🤖 RAG System Example
**File:** `go/rag_example.go`

Complete RAG system implementation in Go:
- Knowledge base creation and management
- Document chunking and processing
- Text embedding generation (simplified for demo)
- Semantic search and retrieval
- Answer generation from context
- Interactive Q&A system

**Usage:**
```bash
cd examples/go
go run rag_example.go
```

**Features:**
- ✅ End-to-end RAG system implementation
- ✅ Document chunking strategies
- ✅ Simplified embedding generation
- ✅ Context-aware answer generation
- ✅ Interactive query system
- ✅ Performance analysis and optimization

---

### 🧪 Simple Index Demo
**File:** `go/simple_demo.go`

Direct usage of VittoriaDB indexing components:
- Flat index operations
- HNSW index operations
- Performance comparisons
- Index factory usage

**Usage:**
```bash
cd examples/go
go run simple_demo.go
```

**Features:**
- ✅ Direct index API usage
- ✅ Performance measurements
- ✅ Index type comparison
- ✅ Memory usage statistics

---

### 🔬 Advanced Features Test
**File:** `go/test_advanced_features.go`

Advanced VittoriaDB functionality testing:
- Complex vector operations
- Advanced indexing features
- Error handling scenarios
- Performance edge cases

**Usage:**
```bash
cd examples/go
go run test_advanced_features.go
```

---

### 🧪 Simple Test
**File:** `go/simple_test.go`

Basic functionality testing:
- Core operations validation
- Simple performance tests
- Basic error scenarios

**Usage:**
```bash
cd examples/go
go run simple_test.go
```

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
   - For RAG applications: `python examples/python/rag_complete_example.py`
   - For document processing: `python examples/python/document_processing_example.py`
   - For performance testing: `python examples/python/performance_benchmark.py`
   - For Go development: `cd examples/go && go run basic_usage.go`
   - For HTTP API testing: `cd examples/curl && ./basic_usage.sh`

3. **Explore the web dashboard:**
   Open http://localhost:8080 in your browser

## 📋 Requirements

### System Requirements
- **VittoriaDB**: Download from [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases)
- **Python**: 3.7+ (for Python examples)
- **Go**: 1.21+ (for Go examples)

### Python Dependencies

> **📦 All Python examples use the centralized VittoriaDB Python library** located in `../sdk/python/vittoriadb/`

```bash
# Core dependencies (required for all examples)
pip install numpy

# For RAG examples with embeddings
pip install sentence-transformers

# For performance benchmarks
pip install psutil

# Optional: for advanced RAG features
pip install openai
```

### Library Installation

> **📦 Install VittoriaDB Python library in development mode** for the best experience:

```bash
# One-time setup: Install in editable/development mode (recommended)
cd sdk/python && ./install-dev.sh

# Or manually:
cd sdk/python && pip install -e .

# Verify installation
python -c "import vittoriadb; print('✅ VittoriaDB Python library ready!')"
```

After installation, all examples use standard imports:
```python
import vittoriadb

# Connect to existing server (no auto-start)
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
```

**Benefits of editable installation:**
- ✅ Clean imports without path manipulation
- ✅ Changes to library code are immediately available
- ✅ Professional development setup
- ✅ IDE autocomplete and type hints work properly
- ✅ Consistent API across all examples
- ✅ Better error handling and debugging

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
