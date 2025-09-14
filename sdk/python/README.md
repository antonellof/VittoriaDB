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

### Basic Usage

```python
import vittoriadb

# Auto-starts VittoriaDB server and connects
db = vittoriadb.connect()

# Create a collection
collection = db.create_collection(
    name="documents",
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

### Automatic Text Embeddings (🚀 NEW!)

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

## 📊 Performance and Scalability

- **Insert Speed**: >10,000 vectors/second with flat indexing, >5,000 with HNSW
- **Search Speed**: Sub-millisecond search times for 1M vectors using HNSW
- **Memory Usage**: <100MB for 100,000 vectors (384 dimensions)
- **Scalability**: Tested up to 1 million vectors, supports up to 2,048 dimensions


## 📋 API Reference

### VittoriaDB Class
- `connect(url=None, auto_start=True, **kwargs)` - Connect to VittoriaDB
- `create_collection(name, dimensions, metric="cosine", vectorizer_config=None)` - Create collection
- `get_collection(name)` - Get existing collection
- `list_collections()` - List all collections
- `delete_collection(name)` - Delete collection
- `health()` - Get server health status
- `close()` - Close connection

### Collection Class
- `insert(id, vector, metadata=None)` - Insert single vector
- `insert_batch(vectors)` - Insert multiple vectors
- `insert_text(id, text, metadata=None)` - Insert text (auto-vectorized)
- `insert_text_batch(texts)` - Insert multiple texts (auto-vectorized)
- `search(vector, limit=10, filter=None)` - Vector similarity search
- `search_text(query, limit=10, filter=None)` - Text search (auto-vectorized)
- `upload_file(file_path, chunk_size=500, **kwargs)` - Upload and process document
- `get(id)` - Get vector by ID
- `delete(id)` - Delete vector by ID
- `count()` - Get total vector count

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

---

**Happy building with VittoriaDB! 🚀**
