# VittoriaDB Examples

This directory contains comprehensive examples demonstrating VittoriaDB's capabilities across different use cases and programming languages.

> **📦 All Python examples use the centralized VittoriaDB Python library** for consistent API usage and better maintainability.

## 🐍 Python Examples

### 🤖 RAG (Retrieval-Augmented Generation) Complete Example
**File:** `rag_complete_example.py`

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
cd python && ./install-dev.sh

# Run the RAG example (uses centralized Python library)
python examples/rag_complete_example.py
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
**File:** `document_processing_example.py`

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
python examples/document_processing_example.py
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
**File:** `performance_benchmark.py`

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
python examples/performance_benchmark.py
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
**File:** `basic_usage.py`

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
python examples/basic_usage.py
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
**File:** `rag_example.py`

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
python examples/rag_example.py
```

**Features:**
- ✅ Advanced RAG implementation using centralized library
- ✅ Optional Sentence Transformers integration
- ✅ Fallback to random embeddings if transformers unavailable
- ✅ Document chunking and processing
- ✅ Interactive query system
- ✅ Graceful error handling and connection management

## 🔧 Go Examples

### 🧪 Simple Index Demo
**File:** `simple_demo.go`

Direct usage of VittoriaDB indexing components:
- Flat index operations
- HNSW index operations
- Performance comparisons
- Index factory usage

**Usage:**
```bash
cd examples
go run simple_demo.go
```

**Features:**
- ✅ Direct index API usage
- ✅ Performance measurements
- ✅ Index type comparison
- ✅ Memory usage statistics

---

### 🔬 Advanced Features Test
**File:** `test_advanced_features.go`

Advanced VittoriaDB functionality testing:
- Complex vector operations
- Advanced indexing features
- Error handling scenarios
- Performance edge cases

**Usage:**
```bash
cd examples
go run test_advanced_features.go
```

---

### 🧪 Simple Test
**File:** `simple_test.go`

Basic functionality testing:
- Core operations validation
- Simple performance tests
- Basic error scenarios

**Usage:**
```bash
cd examples
go run simple_test.go
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
   - For RAG applications: `python examples/rag_complete_example.py`
   - For document processing: `python examples/document_processing_example.py`
   - For performance testing: `python examples/performance_benchmark.py`
   - For Go development: `cd examples && go run simple_demo.go`

3. **Explore the web dashboard:**
   Open http://localhost:8080 in your browser

## 📋 Requirements

### System Requirements
- **VittoriaDB**: Download from [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases)
- **Python**: 3.7+ (for Python examples)
- **Go**: 1.21+ (for Go examples)

### Python Dependencies

> **📦 All Python examples use the centralized VittoriaDB Python library** located in `../python/vittoriadb/`

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
cd python && ./install-dev.sh

# Or manually:
cd python && pip install -e .

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
**Solution:** Install the Python library in development mode: `cd python && ./install-dev.sh`

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
