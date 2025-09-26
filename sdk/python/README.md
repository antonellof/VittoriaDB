# VittoriaDB Python SDK

[![PyPI version](https://badge.fury.io/py/vittoriadb.svg)](https://badge.fury.io/py/vittoriadb)
[![Python versions](https://img.shields.io/pypi/pyversions/vittoriadb.svg)](https://pypi.org/project/vittoriadb/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**VittoriaDB Python SDK** is a client library for **VittoriaDB**, a high-performance embedded vector database built in Go. This SDK provides a clean, Pythonic interface to interact with VittoriaDB server instances, with automatic binary management and server lifecycle control.

## 🏗️ Architecture

**VittoriaDB** consists of two components:
- **🚀 VittoriaDB Server** (Go binary): High-performance vector database engine
- **🐍 Python SDK** (this package): Client library with automatic server management

The Python SDK can either:
- **Auto-manage** the Go server binary (downloads, starts, stops automatically)
- **Connect** to an existing VittoriaDB server instance

## 🚀 Key Features

- **🎯 Zero Configuration**: Works immediately after installation with sensible defaults
- **🤖 Automatic Embeddings**: Server-side text vectorization with multiple model support
- **📄 Document Processing**: Built-in support for PDF, DOCX, TXT, MD, and HTML files
- **🔧 Auto Binary Management**: Automatically downloads and manages VittoriaDB binaries
- **⚡ High Performance**: HNSW indexing provides sub-millisecond search times
- **🐍 Pythonic API**: Clean, intuitive Python interface with type hints
- **🔌 Dual Mode**: Works with existing servers or auto-starts local instances
- **🤖 RAG-Ready**: Built-in content storage for Retrieval-Augmented Generation
- **🚀 NEW v0.5.0**: Unified configuration system with YAML, env vars, and CLI support
- **⚡ NEW v0.5.0**: I/O optimization with memory-mapped storage and SIMD operations
- **🔧 NEW v0.5.0**: Enhanced batch processing with intelligent error recovery
- **🧠 NEW v0.5.0**: Smart chunking with sentence-aware text segmentation
- **🔄 NEW v0.5.0**: Parallel search engine with configurable worker pools and caching

## 📦 Installation

```bash
pip install vittoriadb
```

The package automatically downloads the appropriate VittoriaDB binary for your platform during installation.

## 🚀 Server Management

### Automatic Server Management (Recommended)
```python
import vittoriadb

# SDK automatically downloads, starts, and manages the VittoriaDB server
db = vittoriadb.connect()  # auto_start=True by default
# ... use the database ...
db.close()  # Automatically stops the server
```

### Manual Server Management
```bash
# Download VittoriaDB binary manually
# From: https://github.com/antonellof/VittoriaDB/releases

# Start server manually
./vittoriadb run --port 8080 --data-dir ./data

# In Python, connect to existing server
import vittoriadb
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
```

### Connection Options
```python
# Auto-start with custom configuration
db = vittoriadb.connect(
    auto_start=True,
    port=9090,
    host="localhost", 
    data_dir="./my_vectors"
)

# Connect to remote server
db = vittoriadb.connect(
    url="http://remote-server:8080",
    auto_start=False
)
```

## 🚀 Quick Start

VittoriaDB offers two complementary APIs for different use cases:

### Collection-Based API (Vector Operations)
Perfect for direct vector operations and similarity search:

```python
import vittoriadb

# Auto-starts VittoriaDB server and connects
db = vittoriadb.connect()

# Create a collection
collection = db.create_collection(
    name="vectors",
    dimensions=384,
    metric="cosine"
)

# Insert vectors with metadata
collection.insert(
    id="doc1",
    vector=[0.1, 0.2, 0.3] * 128,  # 384 dimensions
    metadata={"title": "My Document", "category": "tech"}
)

# Search for similar vectors
results = collection.search(
    vector=[0.1, 0.2, 0.3] * 128,
    limit=5,
    include_metadata=True
)

for result in results:
    print(f"ID: {result.id}, Score: {result.score:.4f}")
    print(f"Metadata: {result.metadata}")

# Close connection
db.close()
```

### Document-Based API (Schema & Documents)
Perfect for structured documents with full-text and vector search:

```python
from vittoriadb import create_document_db

# Create database with schema
db = create_document_db({
    "title": "string",
    "content": "string", 
    "price": "number",
    "embedding": "vector[384]",
    "metadata": {"author": "string"}
})

# Insert documents
db.insert({
    "id": "doc1",
    "title": "Noise cancelling headphones",
    "content": "Best headphones on the market",
    "price": 99.99,
    "embedding": [0.1, 0.2, 0.3] * 128,  # 384 dimensions
    "metadata": {"author": "John Doe"}
})

# Full-text search
results = db.search(term="headphones", mode="fulltext", limit=5)

# Vector search
results = db.search(
    mode="vector",
    vector=[0.1, 0.2, 0.3] * 128,
    similarity=0.8,
    limit=5
)

# Hybrid search (combines text and vector)
results = db.search(
    term="best headphones",
    mode="hybrid", 
    vector=[0.1, 0.2, 0.3] * 128,
    limit=5
)

# Advanced search with filters and facets
results = db.search(
    term="headphones",
    where={"price": {"lt": 200}},
    facets={"metadata.author": {"type": "string", "limit": 10}},
    sort_by={"price": "asc"},
    limit=10
)

for hit in results["hits"]:
    doc = hit["document"]
    print(f"Title: {doc['title']}, Price: ${doc['price']}")
```

### Automatic Text Embeddings

```python
import vittoriadb
from vittoriadb.configure import Configure

# Connect to VittoriaDB
db = vittoriadb.connect()

# Create collection with automatic embeddings
collection = db.create_collection(
    name="smart_docs",
    dimensions=384,
    vectorizer_config=Configure.Vectors.auto_embeddings()  # 🎯 Server-side embeddings!
)

# Insert text directly - embeddings generated automatically!
collection.insert_text(
    id="article1",
    text="Artificial intelligence is transforming how we process data.",
    metadata={"category": "AI", "source": "blog"}
)

# Batch insert multiple texts
texts = [
    {
        "id": "article2",
        "text": "Machine learning enables computers to learn from data.",
        "metadata": {"category": "ML"}
    },
    {
        "id": "article3", 
        "text": "Vector databases provide efficient similarity search.",
        "metadata": {"category": "database"}
    }
]
collection.insert_text_batch(texts)

# Search with natural language queries
results = collection.search_text(
    query="artificial intelligence and machine learning",
    limit=3
)

for result in results:
    print(f"Score: {result.score:.4f}")
    print(f"Text: {result.metadata['text'][:100]}...")

db.close()
```

### Document Upload and Processing

```python
import vittoriadb
from vittoriadb.configure import Configure

db = vittoriadb.connect()

# Create collection with vectorizer for automatic processing
collection = db.create_collection(
    name="knowledge_base",
    dimensions=384,
    vectorizer_config=Configure.Vectors.auto_embeddings()
)

# Upload and process documents automatically
result = collection.upload_file(
    file_path="research_paper.pdf",
    chunk_size=600,
    chunk_overlap=100,
    metadata={"source": "research", "year": "2024"}
)

print(f"Processed {result['chunks_created']} chunks")
print(f"Inserted {result['chunks_inserted']} vectors")

# Search the uploaded content
results = collection.search_text(
    query="machine learning algorithms",
    limit=5
)

db.close()
```

## 🤖 RAG (Retrieval-Augmented Generation) Support (NEW v0.4.0)

VittoriaDB now includes built-in support for RAG systems by automatically storing original text content alongside vector embeddings.

### Content Storage Features
- ✅ **Automatic Content Preservation**: Original text stored with vectors
- ✅ **No External Storage Required**: Self-contained RAG solution
- ✅ **Configurable Limits**: Control storage size and behavior
- ✅ **Fast Retrieval**: Single query returns both vectors and content

### RAG-Optimized Collection
```python
import vittoriadb
from vittoriadb import ContentStorageConfig
from vittoriadb.configure import Configure

db = vittoriadb.connect()

# Create RAG-optimized collection with content storage
collection = db.create_collection(
    name="rag_documents",
    dimensions=384,
    vectorizer_config=Configure.Vectors.auto_embeddings(),
    content_storage=ContentStorageConfig(
        enabled=True,           # Store original content
        field_name="_content",  # Metadata field name
        max_size=1048576,      # 1MB limit per document
        compressed=False       # Compression (future feature)
    )
)

# Insert documents - content automatically preserved
collection.insert_text(
    id="doc1",
    text="VittoriaDB is a high-performance vector database perfect for RAG applications...",
    metadata={"title": "VittoriaDB Guide", "category": "documentation"}
)

# Search with content retrieval for RAG
results = collection.search_text(
    query="vector database RAG",
    limit=5,
    include_content=True  # Retrieve original content for LLM context
)

# Use results for RAG
for result in results:
    print(f"Score: {result.score:.3f}")
    print(f"Content: {result.content}")  # Original text for LLM context
    print(f"Has content: {result.has_content()}")

db.close()
```

### RAG Workflow Example
```python
# 1. Store knowledge base
documents = [
    "VittoriaDB supports automatic embeddings...",
    "RAG systems combine retrieval and generation...",
    "Vector databases enable semantic search..."
]

for i, doc in enumerate(documents):
    collection.insert_text(f"kb_{i}", doc, {"type": "knowledge"})

# 2. Query with content for LLM
query = "How do vector databases work?"
results = collection.search_text(query, include_content=True)

# 3. Build context for LLM
context = "\n".join([r.content for r in results if r.has_content()])

# 4. Send to your LLM
# response = your_llm.generate(query, context)
```

## 🎛️ Vectorizer Configuration

VittoriaDB supports multiple vectorizer backends for automatic embedding generation:

### Sentence Transformers (Default)
```python
from vittoriadb.configure import Configure

config = Configure.Vectors.sentence_transformers(
    model="all-MiniLM-L6-v2",
    dimensions=384
)
```

### OpenAI Embeddings
```python
config = Configure.Vectors.openai_embeddings(
    api_key="your-openai-api-key",
    model="text-embedding-ada-002",
    dimensions=1536
)
```

### HuggingFace Models
```python
config = Configure.Vectors.huggingface_embeddings(
    api_key="your-hf-token",  # Optional for public models
    model="sentence-transformers/all-MiniLM-L6-v2",
    dimensions=384
)
```

### Local Ollama
```python
config = Configure.Vectors.ollama_embeddings(
    model="nomic-embed-text",
    dimensions=768,
    base_url="http://localhost:11434"
)
```

## 📄 Document Processing

VittoriaDB supports automatic processing of various document formats:

| Format | Extension | Status | Features |
|--------|-----------|---------|----------|
| **Plain Text** | `.txt` | ✅ Fully Supported | Direct text processing |
| **Markdown** | `.md` | ✅ Fully Supported | Frontmatter parsing |
| **HTML** | `.html` | ✅ Fully Supported | Tag stripping, metadata |
| **PDF** | `.pdf` | ✅ Fully Supported | Multi-page text extraction |
| **DOCX** | `.docx` | ✅ Fully Supported | Properties, text extraction |

```python
# Upload multiple document types
for file_path in ["doc.pdf", "guide.docx", "readme.md"]:
    result = collection.upload_file(
        file_path=file_path,
        chunk_size=500,
        metadata={"batch": "docs_2024"}
    )
    print(f"Processed {file_path}: {result['chunks_inserted']} chunks")
```

## 🔧 Advanced Configuration

### Collection Configuration
```python
# High-performance HNSW configuration
collection = db.create_collection(
    name="large_dataset",
    dimensions=1536,
    metric="cosine",
    index_type="hnsw",
    config={
        "m": 32,                # HNSW connections per node
        "ef_construction": 400,  # Construction search width
        "ef_search": 100        # Search width
    },
    vectorizer_config=Configure.Vectors.openai_embeddings(api_key="your-key")
)
```

### Connection Options
```python
# Connect to existing server
db = vittoriadb.connect(
    url="http://localhost:8080",
    auto_start=False
)

# Auto-start with custom configuration
db = vittoriadb.connect(
    auto_start=True,
    port=9090,
    data_dir="./my_vectors"
)
```

### Search with Filtering
```python
# Search with metadata filters
results = collection.search(
    vector=query_vector,
    limit=10,
    filter={"category": "technology", "year": 2024},
    include_metadata=True
)

# Text search with filters
results = collection.search_text(
    query="machine learning",
    limit=5,
    filter={"source": "research"}
)
```

## 🔧 Configuration Management (NEW v0.5.0)

VittoriaDB v0.5.0 introduces a powerful unified configuration system that supports YAML files, environment variables, and CLI flags with intelligent precedence.

### Configuration Inspection
```python
import vittoriadb

# Connect to VittoriaDB
db = vittoriadb.connect()

# Get current configuration
config_info = db.config()

print("Server Configuration:")
print(f"  Host: {config_info['config']['server']['host']}")
print(f"  Port: {config_info['config']['server']['port']}")

print("Performance Features:")
print(f"  Parallel Search: {config_info['feature_flags']['parallel_search']}")
print(f"  Search Cache: {config_info['feature_flags']['search_cache']}")
print(f"  Memory-Mapped I/O: {config_info['feature_flags']['mmap_storage']}")
print(f"  SIMD Optimizations: {config_info['feature_flags']['simd_optimizations']}")

db.close()
```

### Advanced Server Configuration
```python
# Start server with custom configuration
db = vittoriadb.connect(
    auto_start=True,
    port=9090,
    host="localhost",
    data_dir="./my_vectors",
    # Additional server arguments can be passed
    extra_args=[
        "--config", "my-config.yaml",  # Use YAML configuration
        "--log-level", "debug",        # Enable debug logging
        "--parallel-workers", "16"     # Set parallel search workers
    ]
)
```

## 📊 Performance and Scalability

- **Insert Speed**: >10,000 vectors/second with flat indexing, >5,000 with HNSW
- **Search Speed**: Sub-millisecond search times for 1M vectors using HNSW
- **Memory Usage**: <100MB for 100,000 vectors (384 dimensions)
- **Scalability**: Tested up to 1 million vectors, supports up to 2,048 dimensions
- **🚀 NEW v0.5.0**: Up to 276x speedup with combined I/O optimizations
- **🚀 NEW v0.5.0**: 5-32x faster parallel search for large datasets
- **🚀 NEW v0.5.0**: SIMD operations provide up to 7.7x vector processing speedup


## 📋 API Reference

### Collection Class
- `insert(id, vector, metadata=None)` - Insert single vector
- `insert_batch(vectors)` - Insert multiple vectors
- `insert_text(id, text, metadata=None)` - Insert text (auto-vectorized with content storage)
- `insert_text_batch(texts)` - Insert multiple texts (auto-vectorized with content storage)
- `search(vector, limit=10, filter=None, include_content=False)` - Vector similarity search
- `search_text(query, limit=10, filter=None, include_content=False)` - Text search with content retrieval
- `upload_file(file_path, chunk_size=500, **kwargs)` - Upload and process document
- `get(id)` - Get vector by ID
- `delete(id)` - Delete vector by ID
- `count()` - Get total vector count

### VittoriaDB Class (Enhanced v0.5.0)
- `connect(url=None, auto_start=True, **kwargs)` - Connect to VittoriaDB
- `create_collection(name, dimensions, metric="cosine", vectorizer_config=None, content_storage=None)` - Create collection with content storage
- `get_collection(name)` - Get existing collection
- `list_collections()` - List all collections
- `delete_collection(name)` - Delete collection
- `health()` - Get server health status
- `stats()` - Get database statistics
- `config()` - Get current server configuration (NEW v0.5.0)
- `close()` - Close connection

## 🤝 Contributing

We welcome contributions! 

- **Users**: Report issues and request features on [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues)
- **Developers**: See [DEVELOPMENT.md](DEVELOPMENT.md) for setup, building, and deployment instructions
- **General**: Check our [Contributing Guide](../../CONTRIBUTING.md) for project guidelines

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## 🔗 Links

- **GitHub**: [https://github.com/antonellof/VittoriaDB](https://github.com/antonellof/VittoriaDB)
- **PyPI**: [https://pypi.org/project/vittoriadb/](https://pypi.org/project/vittoriadb/)
- **Issues**: [https://github.com/antonellof/VittoriaDB/issues](https://github.com/antonellof/VittoriaDB/issues)

## 🚀 What's Next?

- 🔍 **Hybrid Search**: Combine vector and keyword search
- 🔐 **Authentication**: User management and access control
- 🌐 **Distributed Mode**: Multi-node clustering support
- 📊 **Analytics**: Query performance monitoring and optimization
- 🎯 **More Vectorizers**: Support for additional embedding models

## 📝 Changelog v0.5.0

### 🆕 Major New Features
- **🔧 Unified Configuration System**: Complete YAML, environment variables, and CLI flags support with intelligent precedence
- **⚡ I/O Optimization Suite**: Memory-mapped storage, SIMD operations, and async I/O for up to 276x performance improvements
- **🔄 Parallel Search Engine**: Configurable worker pools with 5-32x speedup for large datasets
- **🧠 Smart Chunking**: Sentence-aware text segmentation with abbreviation handling
- **🔧 Enhanced Batch Processing**: Intelligent error recovery and fallback mechanisms
- **📊 Configuration API**: New `/config` endpoint for runtime configuration inspection

### 🚀 Performance Improvements
- **SIMD Vector Operations**: Up to 7.7x speedup for vector processing
- **Memory-Mapped I/O**: Zero-copy operations with microsecond latency
- **Search Caching**: LRU cache with TTL expiration for frequently accessed results
- **Batch Processing**: Robust error handling with individual fallback processing

### 🔧 Developer Experience
- **Configuration Management**: CLI tools for generating, validating, and inspecting configurations
- **Hot-Reloading**: Dynamic configuration updates without server restart
- **Comprehensive Documentation**: Updated guides for all new features
- **Enhanced Examples**: New demos showcasing v0.5.0 capabilities

### 🔄 Backward Compatibility
- All existing APIs work unchanged
- Zero configuration setup continues to work seamlessly
- Automatic migration from legacy configurations
- Default behavior maintains full compatibility with v0.4.x

### 🐛 Bug Fixes
- Fixed Windows compatibility with fallback I/O implementation
- Improved error handling in batch processing workflows
- Enhanced stability in parallel search operations

---

**Happy building with VittoriaDB! 🚀**
