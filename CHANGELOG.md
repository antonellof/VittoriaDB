# VittoriaDB Changelog

All notable changes to VittoriaDB will be documented in this file.

## [v0.6.0] - 2024-09-26 - Document API Release

### 🚀 Major New Features

#### Modern Document API
- **Schema-based document storage** with flexible type validation
- **Multiple search modes**: full-text, vector, and hybrid search
- **Advanced filtering**: complex where clauses, facets, sorting, and grouping
- **Modern API design** similar to leading vector databases (Pinecone, Weaviate, Qdrant)
- **Production-ready features** with comprehensive error handling

#### Full-Text Search Engine
- **BM25 scoring** with configurable parameters (k, b, d)
- **Advanced text processing**: tokenization, stemming, stop words
- **Language support** with configurable stemming rules
- **Phrase matching** and boolean operators
- **Score boosting** for different fields

#### Hybrid Search Capabilities
- **Text + Vector search** with configurable weights
- **Intelligent result combination** with score normalization
- **Threshold-based filtering** for quality control
- **Multi-field boosting** for relevance tuning

### 🔧 API Enhancements

#### New HTTP Endpoints
- `POST /create` - Create document database with schema
- `POST /documents` - Insert documents with validation
- `GET /documents/{id}` - Retrieve documents by ID
- `PUT /documents/{id}` - Update documents
- `DELETE /documents/{id}` - Delete documents
- `GET /count` - Count documents with optional filtering
- `POST /search` - Advanced search with multiple modes

#### Enhanced Collection API
- **Dimension compatibility checking** - Automatic collection recreation for schema changes
- **Improved error handling** - Better error messages and validation
- **Performance optimizations** - Faster collection operations

### 🐍 Python SDK Updates

#### New Document Client
```python
from vittoriadb import create_document_db

# Modern document-oriented API
doc_db = create_document_db({
    "title": "string",
    "content": "string",
    "embedding": "vector[384]"
})

doc_db.insert({"title": "My Doc", "content": "...", "embedding": [...]})
results = doc_db.search(term="query", mode="fulltext")
```

#### Enhanced Traditional API
- **Improved error handling** with detailed error messages
- **Better connection management** with automatic retries
- **Performance optimizations** for batch operations

### 📚 Documentation & Examples

#### New Examples
- `examples/python/16_document_api_comprehensive.py` - Full document API demo
- `examples/python/17_quick_test.py` - Quick verification test
- `examples/go/16_document_api_comprehensive.go` - Go document API demo
- `examples/go/17_quick_demo.go` - Go quick test
- `examples/README_DOCUMENT_API.md` - Comprehensive documentation

#### Updated Documentation
- Reorganized README.md for better clarity
- Enhanced API documentation with document examples
- Updated configuration guide with new parameters

### 🔧 Bug Fixes
- **Fixed dimension mismatch errors** - Proper schema-based dimension handling
- **Fixed collection reuse issues** - Automatic dimension compatibility checking
- **Fixed document retrieval** - Proper storage and retrieval of document data
- **Fixed data directory paths** - Consistent path handling across components

### ⚡ Performance Improvements
- **Sub-5ms search performance** for document operations
- **Optimized vector storage** with better memory management
- **Improved indexing** for full-text search operations
- **Enhanced batch processing** for document insertion

---

## [v0.5.0] - 2024-09-15 - Performance & Configuration Release

### 🚀 Major Features

#### Complete RAG Web Application
- **💬 ChatGPT-like Interface**: Modern web UI with real-time streaming responses
- **📁 Multi-Format Document Processing**: PDF, DOCX, TXT, MD, HTML support
- **🌐 Intelligent Web Research**: Real-time search with automatic knowledge storage
- **👨‍💻 GitHub Repository Indexing**: Index and search entire codebases
- **🛑 Operation Control**: Stop button for cancelling long-running operations
- **📚 Built-in Content Storage**: No external storage needed for RAG workflows

#### Unified Configuration System
- **🔧 Multiple Sources**: YAML files, environment variables, CLI flags
- **🔄 Configuration Precedence**: Intelligent merging with clear priority
- **⚡ Hot Reloading**: Dynamic configuration updates without restart
- **📊 Configuration API**: Runtime inspection via HTTP endpoint (`/config`)
- **🔧 CLI Tools**: `vittoriadb config` commands for management

#### Performance Optimizations
- **⚡ I/O Optimization**: Memory-mapped storage, SIMD operations, async I/O (up to 276x speedup)
- **🔄 Parallel Search**: Configurable worker pools with 5-32x performance improvements
- **🧠 Smart Chunking**: Sentence-aware text segmentation with abbreviation handling
- **🔧 Enhanced Batch Processing**: Intelligent error recovery and fallback mechanisms

### 🤖 Embedding Services

#### Professional External Services
- **🔧 Ollama**: Local ML models (high quality, no API costs)
- **🤖 OpenAI**: Cloud API (highest quality, paid)
- **🤗 HuggingFace**: Cloud API (good quality, free tier)
- **🐍 Sentence Transformers**: Local Python models (full control)

#### auto_embeddings() Feature
- **One-line configuration** for professional ML embeddings
- **Zero API costs** with local Ollama models
- **High quality** embeddings (85-95% accuracy)
- **Offline capability** with no internet dependency

### 📚 Content Storage
- **Built-in content preservation** for RAG workflows
- **Configurable storage options** with size limits and compression
- **Metadata integration** with automatic content retrieval
- **Search with content** for complete document reconstruction

---

## [v0.4.0] - 2024-08-20 - Advanced Features Release

### 🚀 New Features
- **HNSW Indexing**: Hierarchical Navigable Small World graphs for fast similarity search
- **Multiple Distance Metrics**: Cosine, Euclidean, Manhattan, Dot Product
- **Batch Operations**: Efficient bulk insert and search operations
- **Advanced Filtering**: Complex metadata filtering with multiple operators
- **Collection Statistics**: Detailed performance and usage metrics

### 🔧 API Enhancements
- **Batch Endpoints**: `/vectors/batch` for bulk operations
- **Search Filtering**: Advanced filter syntax with operators
- **Collection Management**: Enhanced collection lifecycle management
- **Statistics API**: Real-time performance monitoring

### ⚡ Performance
- **10x faster search** with HNSW indexing
- **5x faster insertion** with batch operations
- **Reduced memory usage** with optimized data structures
- **Improved startup time** with lazy loading

---

## [v0.3.0] - 2024-07-15 - Production Features Release

### 🚀 New Features
- **Persistent Storage**: File-based storage with ACID compliance
- **Write-Ahead Logging**: Data durability and crash recovery
- **Collection Management**: Create, list, and delete collections
- **Metadata Support**: Rich metadata storage and filtering
- **Health Monitoring**: System health and statistics endpoints

### 🔧 API Improvements
- **RESTful Design**: Consistent REST API patterns
- **Error Handling**: Comprehensive error responses
- **Input Validation**: Request validation and sanitization
- **CORS Support**: Cross-origin resource sharing

### 📚 Documentation
- **Complete API Reference**: Detailed endpoint documentation
- **Installation Guide**: Multi-platform installation instructions
- **Usage Examples**: Python, Go, and cURL examples

---

## [v0.2.0] - 2024-06-10 - Python SDK Release

### 🐍 Python SDK
- **Native Python Client**: Official Python package on PyPI
- **Automatic Connection Management**: Connection pooling and retries
- **Type Hints**: Full type annotation support
- **Async Support**: Asynchronous operations with asyncio

### 🔧 Features
- **Vector Operations**: Insert, search, and delete vectors
- **Metadata Filtering**: Search with metadata constraints
- **Batch Processing**: Efficient bulk operations
- **Error Handling**: Comprehensive exception handling

---

## [v0.1.0] - 2024-05-01 - Initial Release

### 🚀 Core Features
- **Vector Storage**: High-performance vector storage and retrieval
- **Similarity Search**: Cosine similarity search with configurable limits
- **HTTP API**: RESTful API for all operations
- **Single Binary**: Zero-dependency deployment
- **Cross-Platform**: Linux, macOS, and Windows support

### 🔧 Basic Operations
- **Insert Vectors**: Store vectors with optional metadata
- **Search Vectors**: Find similar vectors with distance scoring
- **Health Check**: System status and basic statistics
- **Configuration**: Basic server configuration options

---

## Migration Guides

### Migrating to v0.6.0 (Document API)

#### From Traditional Collection API
```python
# Old way (still supported)
collection = db.create_collection("docs", dimensions=384)
collection.insert("doc1", vector, metadata)
results = collection.search(query_vector)

# New way (recommended)
doc_db = create_document_db({
    "title": "string",
    "content": "string", 
    "embedding": "vector[384]"
})
doc_db.insert({"id": "doc1", "title": "...", "embedding": vector})
results = doc_db.search(term="query", mode="fulltext")
```

#### Schema Definition
```python
# Define your document structure
schema = {
    "title": "string",           # Full-text searchable
    "description": "string",     # Full-text searchable
    "price": "number",           # Filterable, sortable
    "category": "string",        # Filterable, facetable
    "embedding": "vector[384]",  # Vector searchable
    "metadata": {                # Nested object
        "author": "string",
        "rating": "number"
    },
    "available": "boolean"       # Boolean filtering
}
```

### Migrating to v0.5.0 (Configuration)

#### Environment Variables
```bash
# Old format (still works)
export VITTORIADB_HOST=0.0.0.0
export VITTORIADB_PORT=8080

# New unified format (recommended)
export VITTORIA_SERVER_HOST=0.0.0.0
export VITTORIA_SERVER_PORT=8080
export VITTORIA_PERF_ENABLE_SIMD=true
```

#### Configuration Files
```yaml
# vittoriadb.yaml
server:
  host: "0.0.0.0"
  port: 8080
  cors: true

storage:
  engine: "file"
  page_size: 4096
  cache_size: 1000

performance:
  enable_simd: true
  max_concurrency: 20
  io:
    use_memory_map: true
    async_io: true

search:
  parallel:
    enabled: true
    max_workers: 16
  cache:
    enabled: true
    max_entries: 1000
```

---

## Breaking Changes

### v0.6.0
- **None** - Fully backward compatible
- New document API is additive to existing collection API

### v0.5.0
- **Configuration format changes** - Old environment variables still work
- **Content storage** - New optional feature, doesn't affect existing usage

### v0.4.0
- **Index type parameter** - New optional parameter for collection creation
- **Filter syntax** - Enhanced filter syntax (backward compatible)

### v0.3.0
- **Storage format** - Data migration required from v0.2.0
- **API endpoints** - Some endpoint paths changed

### v0.2.0
- **Python SDK** - New installation method via PyPI
- **API responses** - Enhanced response format with metadata

---

## Deprecation Notices

### v0.6.0
- No deprecations in this release

### v0.5.0
- **Old environment variable format** - Will be removed in v1.0.0
- **Legacy configuration files** - Will be removed in v1.0.0

### v0.4.0
- **Flat index only** - HNSW is now the default for new collections

---

For complete details on any release, see the [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases) page.
