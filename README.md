# VittoriaDB - High-Performance Local Vector Database

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Python Version](https://img.shields.io/badge/Python-3.7+-blue.svg)](https://python.org)
[![PyPI version](https://badge.fury.io/py/vittoriadb.svg)](https://pypi.org/project/vittoriadb/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**VittoriaDB** is a high-performance, embedded vector database designed for local AI development and production deployments. Built with simplicity and performance in mind, it provides a zero-configuration solution for vector similarity search, perfect for RAG applications, semantic search, and AI prototyping.

**🆕 NEW in v0.6.0:** Document-based API with schema validation, full-text search, and hybrid search capabilities!

# Highlighted Features

- [Full-Text Search](docs/api.md#full-text-search) with BM25 scoring
- [Vector Search](docs/api.md#vector-search) with HNSW indexing
- [Hybrid Search](docs/api.md#hybrid-search) combining text and vector
- [Schema-Based Documents](docs/api.md#document-api) with type validation
- [Advanced Filtering](docs/api.md#filtering) with facets and sorting
- [Automatic Embeddings](docs/embeddings.md) with Ollama, OpenAI, HuggingFace
- [High Performance](docs/performance.md) with sub-millisecond search
- [Zero Configuration](docs/installation.md) - works out of the box
- [Single Binary](docs/installation.md) - no dependencies

# Installation

You can install VittoriaDB using our installer script:

```bash
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash
```

Or download manually from [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases):

```bash
wget https://github.com/antonellof/VittoriaDB/releases/latest/download/vittoriadb-linux-amd64.tar.gz
tar -xzf vittoriadb-linux-amd64.tar.gz
chmod +x vittoriadb
```

For Python development, install the SDK:

```bash
pip install vittoriadb
```

Read the complete documentation at [docs/](docs/).

# Usage

VittoriaDB offers two complementary APIs for different use cases:

## Collection-Based API (Vector Operations)

Perfect for direct vector operations and similarity search:

```python
import vittoriadb

# Connect to server
db = vittoriadb.connect("http://localhost:8080")

# Create a collection
collection = db.create_collection("vectors", dimensions=384)

# Insert vectors
collection.insert("vec1", [0.1, 0.2, 0.3, ...], {"category": "example"})

# Search vectors
results = collection.search([0.1, 0.2, 0.3, ...], limit=10)
```

## Document-Based API (Schema & Documents)

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
    "embedding": [0.1, 0.2, 0.3, ...],
    "metadata": {"author": "John Doe"}
})

# Search documents
results = db.search(term="headphones", mode="fulltext")
results = db.search(vector=[0.1, 0.2, ...], mode="vector") 
results = db.search(term="best", vector=[0.1, 0.2, ...], mode="hybrid")
```

VittoriaDB currently supports these data types:

| Type             | Description                                    | Example                           |
| ---------------- | ---------------------------------------------- | --------------------------------- |
| `string`         | A string of characters                         | `'Hello world'`                   |
| `number`         | A numeric value, either float or integer       | `42`                              |
| `boolean`        | A boolean value                                | `true`                            |
| `vector[<size>]` | A vector of numbers for similarity search      | `[0.403, 0.192, 0.830]`          |
| `string[]`       | An array of strings                            | `['red', 'green', 'blue']`        |
| `number[]`       | An array of numbers                            | `[42, 91, 28.5]`                 |
| `boolean[]`      | An array of booleans                           | `[true, false, false]`            |

# Vector and Hybrid Search Support

VittoriaDB supports vector and hybrid search by setting `mode: 'vector'` or `mode: 'hybrid'` when searching.

To perform vector search, provide embeddings at search time:

```python
from vittoriadb import create_document_db

db = create_document_db({
    "title": "string",
    "embedding": "vector[5]"  # 5-dimensional vector
})

# Insert documents with embeddings
db.insert({"title": "The Prestige", "embedding": [0.938293, 0.284951, 0.348264, 0.948276, 0.56472]})
db.insert({"title": "Barbie", "embedding": [0.192839, 0.028471, 0.284738, 0.937463, 0.092827]})
db.insert({"title": "Oppenheimer", "embedding": [0.827391, 0.927381, 0.001982, 0.983821, 0.294841]})

# Vector search
results = db.search(
    mode="vector",
    vector=[0.938292, 0.284961, 0.248264, 0.748276, 0.26472],
    similarity=0.85,  # Minimum similarity threshold
    limit=10
)

# Hybrid search (combines text and vector)
results = db.search(
    term="prestige",
    mode="hybrid", 
    vector=[0.938292, 0.284961, 0.248264, 0.748276, 0.26472],
    limit=10
)
```

# Automatic Embeddings

VittoriaDB supports automatic embedding generation with multiple providers:

```python
from vittoriadb.configure import Configure

# Ollama (local ML models - recommended)
collection = db.create_collection("docs", dimensions=768,
    vectorizer_config=Configure.Vectors.auto_embeddings())

# OpenAI (highest quality)
collection = db.create_collection("docs", dimensions=1536,
    vectorizer_config=Configure.Vectors.openai_embeddings(api_key="your_key"))

# HuggingFace (good quality, free tier)
collection = db.create_collection("docs", dimensions=384,
    vectorizer_config=Configure.Vectors.huggingface_embeddings(api_key="your_key"))

# Insert text directly - embeddings generated automatically
collection.insert_text("doc1", "VittoriaDB is a vector database", {"category": "docs"})

# Search with text - query embedding generated automatically  
results = collection.search_text("What is VittoriaDB?", limit=5)
```

# RAG Applications

Build knowledge bases and RAG systems with VittoriaDB:

```python
import vittoriadb
from vittoriadb.configure import Configure

# Create collection with automatic embeddings
db = vittoriadb.connect("http://localhost:8080")
collection = db.create_collection("knowledge", dimensions=768,
    vectorizer_config=Configure.Vectors.auto_embeddings())

# Add documents to knowledge base
collection.insert_text("doc1", "VittoriaDB is a high-performance vector database", 
    {"category": "documentation"})
collection.insert_text("doc2", "Vector search enables semantic similarity matching",
    {"category": "concepts"})

# Query the knowledge base
results = collection.search_text("What is vector search?", limit=3)

# Use results for RAG
context = "\n".join([hit["metadata"]["content"] for hit in results["hits"]])
# Send context to your LLM for answer generation
```

# Quick Start

1. **Start VittoriaDB:**
   ```bash
   ./vittoriadb run
   ```

2. **Create and use a collection:**
   ```bash
   # Create collection
   curl -X POST http://localhost:8080/collections \
     -H "Content-Type: application/json" \
     -d '{"name": "vectors", "dimensions": 384}'
   
   # Insert vector
   curl -X POST http://localhost:8080/collections/vectors/vectors \
     -H "Content-Type: application/json" \
     -d '{"id": "vec1", "vector": [0.1, 0.2, 0.3, ...], "metadata": {}}'
   
   # Search vectors
   curl "http://localhost:8080/collections/vectors/search?vector=[0.1,0.2,0.3,...]&limit=10"
   ```

3. **Or use the document API:**
   ```bash
   # Create document database
   curl -X POST http://localhost:8080/create \
     -H "Content-Type: application/json" \
     -d '{"schema": {"title": "string", "embedding": "vector[384]"}}'
   
   # Insert document
   curl -X POST http://localhost:8080/documents \
     -H "Content-Type: application/json" \
     -d '{"document": {"title": "My Doc", "embedding": [0.1, 0.2, ...]}}'
   
   # Search documents
   curl -X POST http://localhost:8080/search \
     -H "Content-Type: application/json" \
     -d '{"term": "my doc", "mode": "fulltext"}'
   ```

# Configuration

VittoriaDB works with zero configuration, but supports advanced options:

```bash
# Basic usage
./vittoriadb run

# Custom configuration
./vittoriadb run --host 0.0.0.0 --port 8080 --data-dir ./data

# Performance optimization
export VITTORIA_PERF_ENABLE_SIMD=true
export VITTORIA_SEARCH_PARALLEL_MAX_WORKERS=16
./vittoriadb run
```

# Performance

- **Insert Speed**: >15,000 vectors/second
- **Search Speed**: Sub-100 microsecond cached searches  
- **Memory Usage**: 40% reduction with memory-mapped storage
- **Parallel Search**: 5-32x speedup for large datasets
- **SIMD Operations**: Up to 7.7x speedup for vector processing

# Official Documentation

Read the complete documentation at:

- **[Installation Guide](docs/installation.md)** - Setup for all platforms
- **[API Reference](docs/api.md)** - Complete REST API documentation  
- **[Configuration](docs/configuration.md)** - Server and performance tuning
- **[Embeddings Guide](docs/embeddings.md)** - Automatic embedding services
- **[Performance Guide](docs/performance.md)** - Benchmarks and optimization
- **[CLI Reference](docs/cli.md)** - Command-line interface
- **[Development Guide](docs/development.md)** - Building and contributing

# Examples

Comprehensive examples are available in the [`examples/`](examples/) directory:

- **Python**: RAG systems, document processing, performance benchmarks
- **Go**: Native SDK usage, high-performance testing, advanced features  
- **cURL**: HTTP API workflows, volume testing, bash scripting

```bash
# Start VittoriaDB
./vittoriadb run

# Run examples
python examples/python/16_document_api_comprehensive.py
go run examples/go/16_document_api_comprehensive.go
```

# License

VittoriaDB is licensed under the [MIT](LICENSE) license.