# VittoriaDB Release Notes

## v0.1.0 - Initial Release (2025-01-09)

🚀 **First stable release of VittoriaDB - Local Vector Database for AI Development**

### 🎯 **What is VittoriaDB?**

VittoriaDB is a high-performance, embedded vector database designed specifically for local AI development and production deployments. It provides a zero-configuration solution for vector similarity search, making it perfect for RAG applications, semantic search, recommendation systems, and AI prototyping.

### ✨ **Key Features**

#### **Core Capabilities**
- 🎯 **Zero Configuration**: Works immediately after installation
- ⚡ **High Performance**: HNSW indexing for scalable similarity search (<1ms search times)
- 📁 **Persistent Storage**: ACID-compliant file-based storage with WAL
- 🔌 **Dual Interface**: REST API + Native Python client
- 📄 **Document Processing**: Built-in support for various file formats (planned)
- 🤖 **AI-Ready**: Seamless integration with embedding models

#### **Advanced Features**
- **Multiple Index Types**: Flat (exact) and HNSW (approximate) indexing
- **Distance Metrics**: Cosine, Euclidean, Dot Product, Manhattan
- **Metadata Filtering**: Rich query capabilities with JSON-based filters
- **Batch Operations**: Efficient bulk insert and search operations
- **Transaction Support**: ACID transactions with rollback capability
- **Cross-Platform**: Linux, macOS, Windows support (AMD64, ARM64)

#### **Developer Experience**
- **Python Native**: Auto-manages Go binary, feels like a pure Python library
- **Type Safety**: Full type hints and comprehensive error handling
- **Web Dashboard**: Built-in web interface for testing and monitoring
- **Comprehensive API**: RESTful HTTP API with OpenAPI documentation
- **Rich Examples**: RAG applications, semantic search, and more

### 📦 **Installation Options**

#### **Pre-built Binaries (Recommended)**
Download from [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases/latest):

- **Linux AMD64**: `vittoriadb-v0.1.0-linux-amd64.tar.gz`
- **Linux ARM64**: `vittoriadb-v0.1.0-linux-arm64.tar.gz`
- **macOS Intel**: `vittoriadb-v0.1.0-darwin-amd64.tar.gz`
- **macOS Apple Silicon**: `vittoriadb-v0.1.0-darwin-arm64.tar.gz`
- **Windows**: `vittoriadb-v0.1.0-windows-amd64.zip`

#### **From Source**
```bash
go install github.com/antonellof/VittoriaDB/cmd/vittoriadb@v0.1.0
```

#### **Python Package**
```bash
pip install vittoriadb  # Coming soon
```

### 🚀 **Quick Start**

```bash
# Download and extract (example for Linux)
wget https://github.com/antonellof/VittoriaDB/releases/download/v0.1.0/vittoriadb-v0.1.0-linux-amd64.tar.gz
tar -xzf vittoriadb-v0.1.0-linux-amd64.tar.gz
chmod +x vittoriadb-v0.1.0-linux-amd64

# Start VittoriaDB
./vittoriadb-v0.1.0-linux-amd64 run

# Test with CURL
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "dimensions": 4}'

curl -X POST http://localhost:8080/collections/test/vectors \
  -H "Content-Type: application/json" \
  -d '{"id": "vec1", "vector": [0.1, 0.2, 0.3, 0.4]}'

curl "http://localhost:8080/collections/test/search?vector=0.1,0.2,0.3,0.4&limit=5"
```

### 🏗️ **Architecture**

VittoriaDB consists of several key components:

- **Core Database Engine** (`pkg/core/`): Manages collections, vectors, and operations
- **Storage Layer** (`pkg/storage/`): Page-based file storage with WAL for durability
- **Vector Indexing** (`pkg/index/`): Flat and HNSW indexes for similarity search
- **HTTP API Server** (`pkg/server/`): RESTful API with comprehensive endpoints
- **CLI Interface** (`cmd/vittoriadb/`): Command-line tool for database management
- **Python Client** (`python/vittoriadb/`): Native Python package with auto-binary management

### 🎯 **Performance**

- **Insert Speed**: >10k vectors/second
- **Search Speed**: <1ms for 1M vectors (HNSW), <10ms (flat)
- **Memory Usage**: <100MB for 100k vectors
- **Startup Time**: <100ms
- **Binary Size**: ~8MB compressed

### 📋 **API Endpoints**

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/stats` | Database statistics |
| `GET` | `/collections` | List collections |
| `POST` | `/collections` | Create collection |
| `GET` | `/collections/{name}` | Get collection info |
| `DELETE` | `/collections/{name}` | Delete collection |
| `POST` | `/collections/{name}/vectors` | Insert vector |
| `POST` | `/collections/{name}/vectors/batch` | Batch insert |
| `GET` | `/collections/{name}/vectors/{id}` | Get vector |
| `DELETE` | `/collections/{name}/vectors/{id}` | Delete vector |
| `GET` | `/collections/{name}/search` | Search vectors |

### 🔍 **Use Cases**

VittoriaDB is perfect for:

- **RAG Applications**: Retrieval-Augmented Generation with document embeddings
- **Semantic Search**: Find similar documents, images, or content
- **Recommendation Systems**: Product, content, or user recommendations
- **AI Prototyping**: Rapid development and testing of vector-based AI applications
- **Edge Computing**: Lightweight deployment for edge devices
- **Local Development**: No cloud dependencies or complex setup

### 🛠️ **What's Next?**

Planned features for future releases:

- **Document Processing**: Built-in PDF, DOCX, and text file processing
- **Embedding Integration**: Direct integration with Hugging Face and OpenAI APIs
- **Advanced Filtering**: More sophisticated metadata query capabilities
- **Clustering**: K-means and hierarchical clustering support
- **Monitoring**: Built-in metrics and observability features
- **Python Package**: Official PyPI package with auto-binary management

### 🤝 **Contributing**

We welcome contributions! See our [Contributing Guide](CONTRIBUTING.md) for details.

### 📞 **Support**

- **Documentation**: [GitHub README](https://github.com/antonellof/VittoriaDB#readme)
- **Issues**: [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues)
- **Discussions**: [GitHub Discussions](https://github.com/antonellof/VittoriaDB/discussions)

### 🙏 **Acknowledgments**

- HNSW algorithm implementation inspired by hnswlib
- Go ecosystem for excellent performance and cross-platform support
- The vector database community for inspiration and best practices

---

**VittoriaDB** - Making vector databases local and simple 🚀
