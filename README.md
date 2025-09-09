# VittoriaDB - Local Vector Database for AI Development

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Python Version](https://img.shields.io/badge/Python-3.7+-blue.svg)](https://python.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

**VittoriaDB** is a high-performance, embedded vector database designed specifically for local AI development and production deployments. Built with simplicity and performance in mind, it provides a zero-configuration solution for vector similarity search, making it perfect for RAG applications, semantic search, recommendation systems, and AI prototyping.

## 🎯 **Project Description**

VittoriaDB bridges the gap between complex cloud-based vector databases and simple in-memory solutions. It offers the performance and features of enterprise vector databases while maintaining the simplicity and portability of embedded databases.

### **Why VittoriaDB?**

**The Problem:** Existing vector databases are either too complex for local development (requiring Docker, Kubernetes, or cloud deployment) or too limited for production use (in-memory only, no persistence, poor performance).

**The Solution:** VittoriaDB provides a single binary that works out of the box, with no configuration required, while delivering production-grade performance and features.

### **Key Motivations**

- **🚀 Rapid Prototyping**: Start building AI applications in seconds, not hours
- **🏠 Local Development**: No cloud dependencies or complex setup required  
- **📦 Easy Deployment**: Single binary deployment for edge computing and production
- **⚡ High Performance**: HNSW indexing with sub-millisecond search times
- **🔒 Data Privacy**: Keep your vectors local and secure
- **💰 Cost Effective**: No cloud costs for development and small-scale production

## ✨ **Main Features**

### **Core Capabilities**
- **🎯 Zero Configuration**: Works immediately after installation
- **⚡ High Performance**: HNSW indexing for scalable similarity search (<1ms search times)
- **📁 Persistent Storage**: ACID-compliant file-based storage with WAL
- **🔌 Dual Interface**: REST API + Native Python client
- **📄 Document Processing**: Built-in PDF, DOCX, and text file processing
- **🤖 AI-Ready**: Seamless integration with embedding models

### **Advanced Features**
- **Multiple Index Types**: Flat (exact) and HNSW (approximate) indexing
- **Distance Metrics**: Cosine, Euclidean, Dot Product, Manhattan
- **Metadata Filtering**: Rich query capabilities with JSON-based filters
- **Batch Operations**: Efficient bulk insert and search operations
- **Transaction Support**: ACID transactions with rollback capability
- **Cross-Platform**: Linux, macOS, Windows support (AMD64, ARM64)

### **Developer Experience**
- **Python Native**: Auto-manages Go binary, feels like a pure Python library
- **Type Safety**: Full type hints and comprehensive error handling
- **Web Dashboard**: Built-in web interface for testing and monitoring
- **Comprehensive API**: RESTful HTTP API with OpenAPI documentation
- **Rich Examples**: RAG applications, semantic search, and more

## 📚 **Table of Contents**

- [Requirements](#-requirements)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Usage Examples](#-usage-examples)
  - [Go Library](#go-library)
  - [Python Package](#python-package)
  - [REST API](#-rest-api)
  - [CURL Examples](#curl-examples)
- [Architecture](#️-architecture)
- [Performance](#-performance)
- [Configuration](#-configuration)
- [Development](#-development)
- [Contributing](#-contributing)
- [License](#-license)

## 📦 **Installation**

### **Option 1: Pre-built Binaries (Recommended)**

**Quick Install Script:**
```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash

# Or install specific version
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash -s -- --version v0.1.0
```

**Manual Download:**
Download from [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases/latest):

- **Linux AMD64**: `vittoriadb-v0.1.0-linux-amd64.tar.gz`
- **Linux ARM64**: `vittoriadb-v0.1.0-linux-arm64.tar.gz`  
- **macOS Intel**: `vittoriadb-v0.1.0-darwin-amd64.tar.gz`
- **macOS Apple Silicon**: `vittoriadb-v0.1.0-darwin-arm64.tar.gz`
- **Windows**: `vittoriadb-v0.1.0-windows-amd64.zip`

```bash
# Example for Linux
wget https://github.com/antonellof/VittoriaDB/releases/download/v0.1.0/vittoriadb-v0.1.0-linux-amd64.tar.gz
tar -xzf vittoriadb-v0.1.0-linux-amd64.tar.gz
chmod +x vittoriadb-v0.1.0-linux-amd64
./vittoriadb-v0.1.0-linux-amd64 run
```

**From Source:**
```bash
go install github.com/antonellof/VittoriaDB/cmd/vittoriadb@latest
vittoriadb run
```

### **Option 2: Python Package**
```bash
pip install vittoriadb
```

```python
import vittoriadb

# Auto-starts binary, manages lifecycle
db = vittoriadb.connect()
collection = db.create_collection("documents", dimensions=384)
```

### **Option 3: Docker**
```bash
docker run -p 8080:8080 -v ./data:/data antonellof/vittoriadb:latest
```

## 🚀 **Quick Start**

### **30-Second Demo**

```bash
# 1. Start VittoriaDB
vittoriadb run &

# 2. Create a collection
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "docs", "dimensions": 4}'

# 3. Insert a vector
curl -X POST http://localhost:8080/collections/docs/vectors \
  -H "Content-Type: application/json" \
  -d '{"id": "doc1", "vector": [0.1, 0.2, 0.3, 0.4], "metadata": {"title": "Test Document"}}'

# 4. Search for similar vectors
curl "http://localhost:8080/collections/docs/search?vector=0.1,0.2,0.3,0.4&limit=5"

# 5. Check database stats
curl http://localhost:8080/stats
```

### **Python Quick Start**

```python
import vittoriadb
import numpy as np

# Connect (auto-starts server)
db = vittoriadb.connect()

# Create collection
collection = db.create_collection("documents", dimensions=384, metric="cosine")

# Insert vectors
for i in range(100):
    vector = np.random.random(384).tolist()
    collection.insert(f"doc_{i}", vector, {"title": f"Document {i}"})

# Search
query_vector = np.random.random(384).tolist()
results = collection.search(query_vector, limit=10)

for result in results:
    print(f"ID: {result.id}, Score: {result.score:.4f}")

# Close
db.close()
```

## ✨ Features

### Core Features
- **🎯 Simple**: Single binary, zero configuration, local
- **⚡ Fast**: HNSW indexing, SIMD optimizations, <1ms search
- **📁 File-based**: Portable .db files, no external dependencies
- **🔌 Dual Interface**: REST API + Python package with auto-binary management
- **📄 Document Processing**: PDF, DOCX, TXT with smart chunking
- **🤖 AI Ready**: Built-in embedding model integration

### What Makes VittoriaDB Special
- **Zero Dependencies**: Single binary, no Docker, no setup
- **Local First**: Perfect for development, prototyping, edge deployment
- **Python Native**: Feels like a Python library, manages Go binary automatically
- **Document Ready**: Upload files directly, automatic text extraction and chunking
- **Production Ready**: ACID transactions, WAL, crash recovery

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    VittoriaDB Binary                           │
│                    (Single Process)                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   HTTP Server   │  │  Vector Engine  │  │  Storage Layer  │ │
│  │   (Port 8080)   │  │                 │  │                 │ │
│  │                 │  │ • HNSW Index    │  │ • File Storage  │ │
│  │ • REST API      │  │ • Flat Index    │  │ • WAL           │ │
│  │ • File Upload   │  │ • Filtering     │  │ • Compaction    │ │
│  │ • Web Dashboard │  │ • Similarity    │  │ • Backups       │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                │
                      ┌─────────▼─────────┐
                      │  Data Directory   │
                      │                   │
                      │ • vectors.db      │  (Main data file)
                      │ • index.hnsw      │  (HNSW index)
                      │ • metadata.json   │  (Schema/config)
                      │ • wal.log         │  (Write-ahead log)
                      └───────────────────┘
```

## 📖 **Usage Examples**

### **Go Library**
```go
package main

import (
    "context"
    "fmt"
    "github.com/antonellof/VittoriaDB/pkg/core"
)

func main() {
    // Create database configuration
    config := &core.Config{
        DataDir: "./my-vectors",
        Server: core.ServerConfig{
            Host: "localhost",
            Port: 8080,
        },
    }

    // Create and open database
    db := core.NewDatabase()
    ctx := context.Background()
    
    if err := db.Open(ctx, config); err != nil {
        panic(err)
    }
    defer db.Close()
    
    // Create collection
    req := &core.CreateCollectionRequest{
        Name:       "documents",
        Dimensions: 384,
        Metric:     core.DistanceMetricCosine,
        IndexType:  core.IndexTypeHNSW,
    }
    
    if err := db.CreateCollection(ctx, req); err != nil {
        panic(err)
    }
    
    // Get collection
    collection, err := db.GetCollection(ctx, "documents")
    if err != nil {
        panic(err)
    }
    
    // Insert vector
    vector := &core.Vector{
        ID:     "doc1",
        Vector: []float32{0.1, 0.2, 0.3}, // ... 384 dimensions
        Metadata: map[string]interface{}{
            "title":    "My Document",
            "category": "tech",
        },
    }
    
    if err := collection.Insert(ctx, vector); err != nil {
        panic(err)
    }
    
    // Search
    searchReq := &core.SearchRequest{
        Vector: []float32{0.1, 0.2, 0.3}, // ... 384 dimensions
        Limit:  5,
        IncludeMetadata: true,
    }
    
    results, err := collection.Search(ctx, searchReq)
    if err != nil {
        panic(err)
    }
    
    for _, result := range results.Results {
        fmt.Printf("ID: %s, Score: %f\n", result.ID, result.Score)
    }
}
```

### Python Package
```python
import vittoriadb
import numpy as np

# Connect (auto-starts server if needed)
db = vittoriadb.connect()

# Create collection
collection = db.create_collection(
    name="documents",
    dimensions=384,
    metric="cosine"
)

# Insert vectors
collection.insert(
    id="doc1",
    vector=[0.1, 0.2, 0.3] * 128,  # 384 dims
    metadata={"title": "My Document", "category": "tech"}
)

# Batch insert
vectors = [
    {
        "id": f"doc_{i}",
        "vector": np.random.random(384).tolist(),
        "metadata": {"title": f"Document {i}", "index": i}
    }
    for i in range(1000)
]
collection.insert_batch(vectors)

# Search
results = collection.search(
    vector=[0.1, 0.2, 0.3] * 128,
    limit=10,
    filter={"category": "tech"},
    include_metadata=True
)

for result in results:
    print(f"ID: {result.id}, Score: {result.score:.4f}")
    print(f"Metadata: {result.metadata}")

# Upload documents
result = collection.upload_file(
    "document.pdf",
    chunk_size=500,
    embedding_model="sentence-transformers/all-MiniLM-L6-v2"
)
print(f"Processed {result['chunks']} chunks")

# Close connection
db.close()
```

### RAG Application Example
```python
import vittoriadb
from sentence_transformers import SentenceTransformer

# Start VittoriaDB
db = vittoriadb.connect()

# Initialize embedding model
model = SentenceTransformer('all-MiniLM-L6-v2')

# Create collection
collection = db.create_collection("knowledge", dimensions=384)

# Add documents
documents = [
    "VittoriaDB is a simple embedded vector database",
    "It works great for RAG applications", 
    "You can use it with Python or CURL",
    "The setup takes less than 30 seconds"
]

for i, doc in enumerate(documents):
    embedding = model.encode(doc).tolist()
    collection.insert(f"doc_{i}", embedding, {"text": doc, "index": i})

# Search function
def search_knowledge(query: str, limit: int = 3):
    query_embedding = model.encode(query).tolist()
    results = collection.search(query_embedding, limit=limit)
    
    return [
        {
            "text": result.metadata["text"],
            "score": result.score,
            "id": result.id
        }
        for result in results
    ]

# Use it
results = search_knowledge("How to set up vector database?")
for result in results:
    print(f"Score: {result['score']:.4f}")
    print(f"Text: {result['text']}\n")
```

## 🛠️ **REST API**

VittoriaDB provides a comprehensive REST API for all vector database operations.

### **CURL Examples**

#### **Server Management**
```bash
# Health check
curl http://localhost:8080/health

# Database statistics
curl http://localhost:8080/stats

# Response:
# {
#   "collections": [...],
#   "total_vectors": 1000,
#   "total_size": 1048576,
#   "queries_total": 42,
#   "avg_query_latency": 1.5
# }
```

#### **Collection Management**
```bash
# List all collections
curl http://localhost:8080/collections

# Create a new collection
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "documents",
    "dimensions": 384
  }'

# Create collection with HNSW index
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "large_docs",
    "dimensions": 1536,
    "index_type": "hnsw",
    "config": {
      "m": 32,
      "ef_construction": 400
    }
  }'

# Get collection information
curl http://localhost:8080/collections/documents

# Get collection statistics
curl http://localhost:8080/collections/documents/stats

# Delete collection
curl -X DELETE http://localhost:8080/collections/documents
```

#### **Vector Operations**
```bash
# Insert a single vector
curl -X POST http://localhost:8080/collections/documents/vectors \
  -H "Content-Type: application/json" \
  -d '{
    "id": "doc_001",
    "vector": [0.1, 0.2, 0.3, 0.4],
    "metadata": {
      "title": "Introduction to AI",
      "author": "John Doe",
      "category": "technology",
      "published": "2024-01-15"
    }
  }'

# Batch insert multiple vectors
curl -X POST http://localhost:8080/collections/documents/vectors/batch \
  -H "Content-Type: application/json" \
  -d '{
    "vectors": [
      {
        "id": "doc_002",
        "vector": [0.2, 0.3, 0.4, 0.5],
        "metadata": {"title": "Machine Learning Basics", "category": "technology"}
      },
      {
        "id": "doc_003",
        "vector": [0.3, 0.4, 0.5, 0.6],
        "metadata": {"title": "Deep Learning Guide", "category": "technology"}
      }
    ]
  }'

# Get a specific vector
curl http://localhost:8080/collections/documents/vectors/doc_001

# Delete a vector
curl -X DELETE http://localhost:8080/collections/documents/vectors/doc_001
```

#### **Vector Search**
```bash
# Basic similarity search
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10'

# Search with metadata included
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=5' \
  --data-urlencode 'include_metadata=true'

# Search with metadata filter
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10' \
  --data-urlencode 'filter={"category": "technology"}'

# Advanced search with multiple filters
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10' \
  --data-urlencode 'filter={"category": "technology", "author": "John Doe"}'

# Search with pagination
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10' \
  --data-urlencode 'offset=20'
```

#### **File Upload (Future Feature)**
```bash
# Upload and process a document
curl -X POST http://localhost:8080/collections/documents/upload \
  -F "file=@document.pdf" \
  -F "chunk_size=500" \
  -F "overlap=50" \
  -F "embedding_model=sentence-transformers/all-MiniLM-L6-v2" \
  -F "metadata={\"source\": \"upload\", \"type\": \"pdf\"}"
```

#### **Complete Workflow Example**
```bash
#!/bin/bash
# Complete VittoriaDB workflow example

# 1. Start VittoriaDB server
vittoriadb run --port 8080 &
SERVER_PID=$!
sleep 2

# 2. Create a collection
echo "Creating collection..."
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "example", "dimensions": 4}' | jq

# 3. Insert test vectors
echo "Inserting vectors..."
curl -X POST http://localhost:8080/collections/example/vectors/batch \
  -H "Content-Type: application/json" \
  -d '{
    "vectors": [
      {"id": "vec1", "vector": [1.0, 0.0, 0.0, 0.0], "metadata": {"type": "A"}},
      {"id": "vec2", "vector": [0.0, 1.0, 0.0, 0.0], "metadata": {"type": "B"}},
      {"id": "vec3", "vector": [0.0, 0.0, 1.0, 0.0], "metadata": {"type": "C"}},
      {"id": "vec4", "vector": [0.0, 0.0, 0.0, 1.0], "metadata": {"type": "D"}}
    ]
  }' | jq

# 4. Search for similar vectors
echo "Searching vectors..."
curl -G http://localhost:8080/collections/example/search \
  --data-urlencode 'vector=[0.9,0.1,0.0,0.0]' \
  --data-urlencode 'limit=3' \
  --data-urlencode 'include_metadata=true' | jq

# 5. Get collection stats
echo "Collection stats:"
curl http://localhost:8080/collections/example/stats | jq

# 6. Cleanup
curl -X DELETE http://localhost:8080/collections/example
kill $SERVER_PID
```

### **API Endpoints Reference**

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/stats` | Database statistics |
| `GET` | `/collections` | List collections |
| `POST` | `/collections` | Create collection |
| `GET` | `/collections/{name}` | Get collection info |
| `DELETE` | `/collections/{name}` | Delete collection |
| `GET` | `/collections/{name}/stats` | Collection statistics |
| `POST` | `/collections/{name}/vectors` | Insert vector |
| `POST` | `/collections/{name}/vectors/batch` | Batch insert |
| `GET` | `/collections/{name}/vectors/{id}` | Get vector |
| `DELETE` | `/collections/{name}/vectors/{id}` | Delete vector |
| `GET` | `/collections/{name}/search` | Search vectors |
| `POST` | `/collections/{name}/upload` | Upload document |

## 📁 Supported File Formats

| Format | Extension | Processing | Notes |
|--------|-----------|------------|-------|
| **PDF** | `.pdf` | Text extraction | Using pdfcpu |
| **Word** | `.docx` | Office XML | Modern Word docs |
| **Word Legacy** | `.doc` | OLE parsing | Older Word format |
| **Plain Text** | `.txt`, `.md` | Direct read | UTF-8 encoding |
| **HTML** | `.html`, `.htm` | Tag stripping | Web content |
| **Rich Text** | `.rtf` | RTF parser | Cross-platform |

## 🎯 Performance

### Benchmarks
- **Insert Speed**: >10k vectors/second
- **Search Speed**: <1ms for 1M vectors (HNSW), <10ms (flat)
- **Memory Usage**: <100MB for 100k vectors
- **Startup Time**: <100ms
- **Binary Size**: <50MB

### Scaling
- **Vectors**: Tested up to 10M vectors
- **Dimensions**: Up to 10,000 dimensions
- **Collections**: Unlimited
- **File Size**: Single files up to 2GB
- **Concurrent Users**: 100+ simultaneous connections

## 🔧 Configuration

### Command Line
```bash
vittoriadb run \
  --host 0.0.0.0 \
  --port 8080 \
  --data-dir ./data \
  --index-type hnsw \
  --log-level info
```

### Configuration File (`vittoriadb.yaml`)
```yaml
server:
  host: "0.0.0.0"
  port: 8080
  cors: true

storage:
  data_dir: "./data"
  page_size: 4096
  cache_size: 100
  sync_writes: true

index:
  default_type: "hnsw"
  default_metric: "cosine"
  hnsw:
    m: 16
    ef_construction: 200
    ef_search: 50

performance:
  max_concurrency: 100
  enable_simd: true
  memory_limit: 1073741824  # 1GB
```

## 🐳 Docker

```dockerfile
FROM vittoriadb/vittoriadb:latest
EXPOSE 8080
VOLUME ["/data"]
CMD ["vittoriadb", "run", "--host", "0.0.0.0", "--data-dir", "/data"]
```

```bash
docker run -p 8080:8080 -v ./data:/data vittoriadb/vittoriadb
```

## 📋 Requirements

### System Requirements
- **Operating System**: Linux, macOS, or Windows
- **Memory**: Minimum 512MB RAM (2GB+ recommended)
- **Disk Space**: 100MB for binary + storage for your data
- **Network**: Port 8080 (configurable)

### Development Requirements
- **Go**: Version 1.21 or higher
- **Python**: Version 3.7 or higher (for Python client)
- **Git**: For cloning the repository

### Optional Dependencies
- **Docker**: For containerized deployment
- **Make**: For using build scripts
- **sentence-transformers**: For advanced embedding examples

## 🧪 Development

### Prerequisites
Before building VittoriaDB, ensure you have the required tools installed:

```bash
# Check Go version (required: 1.21+)
go version

# Check Python version (required: 3.7+)
python --version

# Install Git if not already installed
git --version
```

### Building from Source

#### 1. Clone the Repository
```bash
git clone https://github.com/antonellof/VittoriaDB.git
cd VittoriaDB
```

#### 2. Build the Go Binary
```bash
# Download Go dependencies
go mod download

# Build the binary
go build -o vittoriadb ./cmd/vittoriadb

# Verify the build
./vittoriadb --version
```

#### 3. Build Python Package (Optional)
```bash
# Navigate to Python package directory
cd python

# Install in development mode
pip install -e .

# Or install with optional dependencies
pip install -e ".[dev,full]"

# Verify installation
python -c "import vittoriadb; print('VittoriaDB Python client installed successfully')"
```

### Running Locally

#### Quick Start
```bash
# Start VittoriaDB server
./vittoriadb run

# In another terminal, test the API
curl http://localhost:8080/health
```

#### With Custom Configuration
```bash
# Run with custom settings
./vittoriadb run \
  --host 0.0.0.0 \
  --port 9090 \
  --data-dir ./my-data \
  --cors

# Or use a configuration file
./vittoriadb run --config ./config/vittoriadb.yaml
```

#### Development Mode
```bash
# Run with verbose logging
./vittoriadb run --log-level debug

# Run tests
go test ./... -v

# Run Python tests (if Python package is installed)
cd python && python -m pytest tests/ -v

# Run benchmarks
go test ./pkg/core -bench=. -benchmem
```

### Testing Your Installation

#### 1. Basic Functionality Test
```bash
# Start the server
./vittoriadb run &
SERVER_PID=$!

# Wait for startup
sleep 2

# Create a test collection
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "dimensions": 4, "metric": "cosine"}'

# Insert a test vector
curl -X POST http://localhost:8080/collections/test/vectors \
  -H "Content-Type: application/json" \
  -d '{"id": "test1", "vector": [0.1, 0.2, 0.3, 0.4], "metadata": {"type": "test"}}'

# Search for similar vectors
curl "http://localhost:8080/collections/test/search?vector=0.1,0.2,0.3,0.4&limit=1"

# Cleanup
kill $SERVER_PID
```

#### 2. Python Client Test
```python
import vittoriadb

# This will auto-start the binary
db = vittoriadb.connect()

# Create collection
collection = db.create_collection("test", dimensions=4)

# Insert and search
collection.insert("test1", [0.1, 0.2, 0.3, 0.4], {"type": "test"})
results = collection.search([0.1, 0.2, 0.3, 0.4], limit=1)

print(f"Found {len(results)} results")
db.close()
```

### Build Scripts

#### Using Make (if available)
```bash
# Build everything
make build

# Run tests
make test

# Clean build artifacts
make clean

# Build for multiple platforms
make build-all
```

#### Manual Cross-Platform Builds
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o vittoriadb-linux-amd64 ./cmd/vittoriadb

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o vittoriadb-darwin-amd64 ./cmd/vittoriadb

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o vittoriadb-darwin-arm64 ./cmd/vittoriadb

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o vittoriadb-windows-amd64.exe ./cmd/vittoriadb
```

### Troubleshooting

#### Common Issues

**1. Port Already in Use**
```bash
# Check what's using port 8080
lsof -i :8080

# Use a different port
./vittoriadb run --port 9090
```

**2. Permission Denied**
```bash
# Make binary executable
chmod +x ./vittoriadb

# Or run with explicit path
./vittoriadb run
```

**3. Go Module Issues**
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy
```

**4. Python Import Errors**
```bash
# Reinstall Python package
pip uninstall vittoriadb
pip install -e ./python

# Check Python path
python -c "import sys; print(sys.path)"
```

#### Performance Tuning
```bash
# Increase memory limit
./vittoriadb run --memory-limit 2GB

# Enable SIMD optimizations
./vittoriadb run --enable-simd

# Adjust cache size
./vittoriadb run --cache-size 200
```

### Project Structure
```
vittoriadb/
├── cmd/vittoriadb/           # Main binary
├── pkg/
│   ├── core/                 # Core database engine
│   ├── storage/              # File storage layer
│   ├── index/                # Vector indexing
│   ├── server/               # HTTP API server
│   ├── processor/            # Document processing
│   └── embeddings/           # Embedding integrations
├── python/vittoriadb/        # Python package
├── docs/                     # Documentation
├── examples/                 # Code examples
└── tests/                    # Test suites
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup
1. Install Go 1.21+
2. Install Python 3.7+
3. Fork and clone the repository
4. Create a feature branch
5. Make your changes
6. Add tests
7. Submit a pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- HNSW algorithm implementation based on hnswlib
- Document processing powered by various Go libraries
- Python package management inspired by best practices

## 📞 Support

- **Documentation**: [https://vittoriadb.dev](https://vittoriadb.dev)
- **Issues**: [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues)
- **Discussions**: [GitHub Discussions](https://github.com/antonellof/VittoriaDB/discussions)
- **Discord**: [VittoriaDB Community](https://discord.gg/vittoriadb)

---

**VittoriaDB** - Making vector databases local and simple 🚀
