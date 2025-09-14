# VittoriaDB - Local Vector Database for AI Development

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![Python Version](https://img.shields.io/badge/Python-3.7+-blue.svg)](https://python.org)
[![PyPI version](https://badge.fury.io/py/vittoriadb.svg)](https://pypi.org/project/vittoriadb/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)]()

**VittoriaDB** is a high-performance, embedded vector database designed for local AI development and production deployments. Built with simplicity and performance in mind, it provides a zero-configuration solution for vector similarity search, perfect for RAG applications, semantic search, and AI prototyping.

## 🎯 Why VittoriaDB?

**The Problem:** Existing vector databases are either too complex for local development (requiring Docker, Kubernetes, or cloud deployment) or too limited for production use (in-memory only, no persistence, poor performance).

**The Solution:** VittoriaDB provides a single binary that works out of the box, with no configuration required, while delivering production-grade performance and features.

## ✨ Key Features

- **🎯 Zero Configuration**: Works immediately after installation
- **🤖 Professional Embedding Services**: Industry-standard vectorization options
  - **Ollama**: Local ML models (high quality, no API costs)
  - **OpenAI**: Cloud API (highest quality, paid)
  - **HuggingFace**: Cloud API (good quality, free tier)
  - **Sentence Transformers**: Local Python models (full control)
  - **Pure Vector DB**: Bring your own embeddings
- **⚡ High Performance**: HNSW indexing with sub-millisecond search times
- **📁 Persistent Storage**: ACID-compliant file-based storage with WAL
- **🔌 Dual Interface**: REST API + Native Python client
- **🧠 AI-Ready**: Built for RAG, semantic search, and embedding workflows
- **📦 Single Binary**: No dependencies, cross-platform support
- **🔒 Local First**: Keep your data private and secure

## 📚 Documentation

- **[📦 Installation Guide](docs/installation.md)** - Complete installation instructions for all platforms
- **[🚀 Quick Start](#-quick-start)** - Get started in 30 seconds
- **[🐍 Python SDK](https://pypi.org/project/vittoriadb/)** - Official Python package on PyPI (`pip install vittoriadb`)
- **[🤖 Embedding Services](docs/embeddings.md)** - Complete guide to auto_embeddings() and vectorizers
- **[📖 Usage Examples](#-usage-examples)** - Python, Go, and cURL examples
- **[🛠️ API Reference](docs/api.md)** - Complete REST API documentation
- **[⚙️ Configuration](docs/configuration.md)** - Server and storage configuration
- **[🖥️ CLI Commands](docs/cli.md)** - Command-line interface reference
- **[📊 Performance](docs/performance.md)** - Benchmarks and optimization guide
- **[🧪 Development](docs/development.md)** - Building and contributing guide

## 📦 Installation

### Quick Install (Recommended)
```bash
# One-line installer for latest version
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash
```

### Manual Installation
```bash
# Download for your platform from GitHub Releases
wget https://github.com/antonellof/VittoriaDB/releases/download/v0.2.0/vittoriadb-v0.2.0-linux-amd64.tar.gz
tar -xzf vittoriadb-v0.2.0-linux-amd64.tar.gz
chmod +x vittoriadb-v0.2.0-linux-amd64
./vittoriadb-v0.2.0-linux-amd64 run
```

### Python SDK
```bash
# Install from PyPI (recommended)
pip install vittoriadb

# Or install from source for development
git clone https://github.com/antonellof/VittoriaDB.git
cd VittoriaDB/sdk/python && ./install-dev.sh
```

> 📖 **See [Installation Guide](docs/installation.md) for complete instructions, platform-specific details, and troubleshooting.**

## 🚀 Quick Start

### 30-Second Demo
```bash
# 1. Start VittoriaDB
vittoriadb run

# 2. Create a collection and insert vectors
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "docs", "dimensions": 4}'

curl -X POST http://localhost:8080/collections/docs/vectors \
  -H "Content-Type: application/json" \
  -d '{"id": "doc1", "vector": [0.1, 0.2, 0.3, 0.4], "metadata": {"title": "Test Document"}}'

# 3. Search for similar vectors
curl "http://localhost:8080/collections/docs/search?vector=0.1,0.2,0.3,0.4&limit=5"
```

### Python Quick Start

VittoriaDB offers **four professional approaches** for handling embeddings:

#### 🔧 **Approach 1: Ollama (Recommended)**
```python
import vittoriadb
from vittoriadb.configure import Configure

# Connect to running server
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)

# Create collection with Ollama local ML models (requires: ollama pull nomic-embed-text)
collection = db.create_collection(
    name="documents", 
    dimensions=768,  # nomic-embed-text dimensions
    vectorizer_config=Configure.Vectors.auto_embeddings()  # 🎯 Local ML!
)

# Insert text directly - server generates embeddings using local ML model
collection.insert_text("doc1", "Your document content here", {"title": "My Document"})

# Search with text - server generates query embedding using local ML model
results = collection.search_text("find similar documents", limit=10)
print(f"Found {len(results)} results")
```

#### 🤖 **Approach 2: OpenAI API (Highest Quality)**
```python
# OpenAI embeddings (highest quality, requires API key + credits)
collection = db.create_collection(
    name="openai_docs",
    dimensions=1536,
    vectorizer_config=Configure.Vectors.openai_embeddings(api_key="your_openai_key")
)
```

#### 🤗 **Approach 3: HuggingFace API (Free Tier)**
```python
# HuggingFace embeddings (good quality, free tier available)
collection = db.create_collection(
    name="hf_docs", 
    dimensions=384,
    vectorizer_config=Configure.Vectors.huggingface_embeddings(api_key="your_hf_token")
)
```

#### 🐍 **Approach 4: Sentence Transformers (Local Python)**
```python
# Local Python models (full control, heavy dependencies)
collection = db.create_collection(
    name="local_docs",
    dimensions=384,
    vectorizer_config=Configure.Vectors.sentence_transformers()
)
```

#### 💎 **Approach 5: Pure Vector Database (Manual Embeddings)**
```python
import vittoriadb
from sentence_transformers import SentenceTransformer

db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
model = SentenceTransformer('all-MiniLM-L6-v2')  # Client-side model

# Create collection without vectorizer
collection = db.create_collection(name="documents", dimensions=384)

# Generate embeddings on client side
text = "Your document content here"
embedding = model.encode(text).tolist()
collection.insert("doc1", embedding, {"title": "My Document", "content": text})

# Generate query embedding on client side
query_embedding = model.encode("find similar documents").tolist()
results = collection.search(query_embedding, limit=10)
print(f"Found {len(results)} results")
```

## 🤖 auto_embeddings(): The Smart Default

The `Configure.Vectors.auto_embeddings()` function is VittoriaDB's **intelligent embedding solution** that provides the best balance of quality, performance, and ease of use.

### What Makes auto_embeddings() Special?

```python
# One line for professional ML embeddings
vectorizer_config = Configure.Vectors.auto_embeddings()
```

**Behind the scenes, auto_embeddings():**
1. **Uses Ollama local ML models** - Real neural networks, not statistical approximations
2. **Requires minimal setup** - Just `ollama pull nomic-embed-text`
3. **Works completely offline** - No API keys, no internet required
4. **Provides high quality** - 85-95% accuracy comparable to cloud APIs
5. **Costs nothing to run** - No per-request charges or rate limits

### Why Choose auto_embeddings()?

| Traditional Approach | auto_embeddings() Advantage |
|---------------------|------------------------------|
| ❌ Complex model management | ✅ One-line configuration |
| ❌ API costs and rate limits | ✅ Completely free to use |
| ❌ Internet dependency | ✅ Works offline |
| ❌ Statistical approximations | ✅ Real ML neural networks |
| ❌ Vendor lock-in | ✅ Open-source local models |

### Quick Setup

```bash
# 1. Install Ollama (one-time setup)
curl -fsSL https://ollama.ai/install.sh | sh

# 2. Start Ollama service
ollama serve

# 3. Pull embedding model (one-time download)
ollama pull nomic-embed-text

# 4. Use with VittoriaDB
python -c "
import vittoriadb
from vittoriadb.configure import Configure

db = vittoriadb.connect()
collection = db.create_collection(
    name='test',
    dimensions=768,
    vectorizer_config=Configure.Vectors.auto_embeddings()
)
print('✅ Ready for high-quality local ML embeddings!')
"
```

> 📖 **See [Embedding Services Guide](docs/embeddings.md) for complete documentation, advanced configuration, and comparison of all vectorizer options.**

## 🏗️ Architecture & Embedding Approaches

VittoriaDB is a single-process binary that combines an HTTP server, vector engine, and storage layer. It offers **professional external embedding services** following industry best practices:

### 🔧 **External Service Architecture**

**Clean delegation to specialized embedding services**

```
┌─────────────────────────────────────────────────────────────┐
│ Python Client: Configure.Vectors.auto_embeddings()         │
└─────────────────────┬───────────────────────────────────────┘
                      │ HTTP Request (text)
┌─────────────────────▼───────────────────────────────────────┐
│ VittoriaDB Server: External Service Delegation             │
│ ├─ Text preprocessing and validation                       │
│ ├─ Route to appropriate external service                   │
│ └─ Handle API calls and error management                   │
└─────────────────────┬───────────────────────────────────────┘
                      │ Delegate to external services
┌─────────────────────▼───────────────────────────────────────┐
│ External Embedding Services (Real ML Models)               │
│ ├─ 🔧 Ollama: Local ML models (localhost:11434)           │
│ ├─ 🤖 OpenAI: Cloud API (api.openai.com)                  │
│ ├─ 🤗 HuggingFace: Cloud API (api-inference.huggingface.co)│
│ └─ 🐍 Sentence Transformers: Python subprocess            │
└─────────────────────┬───────────────────────────────────────┘
                      │ Return high-quality embeddings
┌─────────────────────▼───────────────────────────────────────┐
│ Vector Storage & Search Engine                              │
└─────────────────────────────────────────────────────────────┘
```

**Benefits:**
- ✅ **Industry standard** - follows patterns used by Weaviate, Pinecone, Qdrant
- ✅ **High-quality embeddings** - real ML models, not statistical approximations
- ✅ **Flexible deployment** - local ML, cloud APIs, or Python processes
- ✅ **Maintainable codebase** - no complex local ML implementations
- ✅ **Future-proof** - easy to add new services as they emerge

### 🎯 **Service Comparison**

| Service | Quality | Speed | Setup | Cost | Best For |
|---------|---------|-------|-------|------|----------|
| **🔧 Ollama** | High (85-95%) | Fast (~500ms) | `ollama pull nomic-embed-text` | Free | **Recommended** |
| **🤖 OpenAI** | Highest (95%+) | Medium (~300ms) | API key required | $0.0001/1K tokens | **Highest Quality** |
| **🤗 HuggingFace** | High (80-90%) | Medium (~500ms) | API token | Free tier | **Cost Effective** |
| **🐍 Sentence Transformers** | High (85-95%) | Slow (~5s) | `pip install sentence-transformers` | Free | **Full Control** |

> 📖 **See [Performance Guide](docs/performance.md) for detailed architecture diagrams and performance characteristics.**

## 📖 Usage Examples

The [`examples/`](examples/) directory contains comprehensive examples organized by language:

- **🐍 Python**: RAG systems, document processing, performance benchmarks
- **🔧 Go**: Native SDK usage, high-performance testing, advanced features  
- **🌐 cURL**: HTTP API workflows, volume testing, bash scripting

```bash
# Start VittoriaDB
./vittoriadb run

# Run examples
python examples/python/rag_complete_example.py
cd examples/go && go run basic_usage.go
cd examples/curl && ./basic_usage.sh
```

> 📖 **See [examples/README.md](examples/README.md) for complete documentation and requirements.**

### Go Library Example
```go
import "github.com/antonellof/VittoriaDB/pkg/core"

// Create database and collection
db := core.NewDatabase()
db.Open(ctx, &core.Config{DataDir: "./my-vectors"})

// Insert and search vectors
collection.Insert(ctx, &core.Vector{
    ID: "doc1", 
    Vector: []float32{0.1, 0.2, 0.3, 0.4},
    Metadata: map[string]interface{}{"title": "My Document"},
})
```

### Python SDK Example  
```python
# Install: pip install vittoriadb
import vittoriadb
from vittoriadb.configure import Configure

# Connect to server
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)

# Create collection with automatic embeddings
collection = db.create_collection(
    name="docs", 
    dimensions=768,
    vectorizer_config=Configure.Vectors.auto_embeddings()  # Uses Ollama
)

# Insert text directly - server generates embeddings
collection.insert_text("doc1", "Your document content", {"title": "My Document"})

# Search with text - server generates query embedding
results = collection.search_text("find similar content", limit=10)
```

### RAG Application Example
```python
from sentence_transformers import SentenceTransformer

model = SentenceTransformer('all-MiniLM-L6-v2')
collection = db.create_collection("knowledge", dimensions=384)

# Add documents with embeddings
for doc in documents:
    embedding = model.encode(doc).tolist()
    collection.insert(f"doc_{i}", embedding, {"text": doc})

# Search knowledge base
def search_knowledge(query):
    embedding = model.encode(query).tolist()
    return collection.search(embedding, limit=3)
```

## 🛠️ REST API

VittoriaDB provides a comprehensive REST API for all vector database operations:

```bash
# Create collection
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "docs", "dimensions": 384}'

# Insert vector
curl -X POST http://localhost:8080/collections/docs/vectors \
  -H "Content-Type: application/json" \
  -d '{"id": "doc1", "vector": [0.1, 0.2, 0.3, 0.4], "metadata": {"title": "My Doc"}}'

# Search vectors
curl -G http://localhost:8080/collections/docs/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10'
```

> 📖 **See [API Reference](docs/api.md) for complete endpoint documentation, examples, and response formats.**

## 🎯 Performance

### Benchmarks (v0.2.0)
- **Insert Speed**: >2.6M vectors/second (HNSW)
- **Search Speed**: <1ms latency (sub-millisecond for optimized queries)
- **Memory Usage**: Linear scaling - 1MB for 1K vectors, 167MB for 50K vectors
- **Startup Time**: <100ms cold start
- **Binary Size**: ~8MB compressed

### Comprehensive Performance Results
📊 **[View Complete Benchmark Results](https://gist.github.com/antonellof/19069bb56573fcf72ce592b3c2f2fc74)** - Detailed performance testing with Native Go SDK integration

**Key Highlights:**
- **Peak Insert Rate**: 2,645,209 vectors/sec
- **Peak Search Rate**: 1,266.72 searches/sec  
- **Lowest Latency**: 789.44µs
- **Large-Scale Performance**: 1,685,330 vectors/sec for 87.89 MB dataset

> 📖 **See [Performance Guide](docs/performance.md) for detailed benchmarks, optimization tips, and scaling characteristics.**

## 🔧 Configuration

### Basic Configuration
```bash
# Start with custom settings
vittoriadb run \
  --host 0.0.0.0 \
  --port 8080 \
  --data-dir ./data \
  --cors

# Use configuration file
vittoriadb run --config vittoriadb.yaml
```

### Data Directory
VittoriaDB stores all data in a configurable directory (default: `./data`):
```bash
vittoriadb run --data-dir /path/to/your/data
export VITTORIADB_DATA_DIR=/path/to/your/data
```

> 📖 **See [Configuration Guide](docs/configuration.md) for complete options, YAML configuration, and data directory management.**

## 🖥️ CLI Commands

### Core Commands
```bash
# Start the server
vittoriadb run

# Show version and build info
vittoriadb version

# Inspect database
vittoriadb info [--data-dir <path>]
vittoriadb stats [--data-dir <path>]
```

### Server Options
```bash
vittoriadb run \
  --host 0.0.0.0 \
  --port 8080 \
  --data-dir ./data \
  --config config.yaml
```

> 📖 **See [CLI Reference](docs/cli.md) for complete command documentation, options, and environment variables.**

## 📋 System Requirements

- **Operating System**: Linux, macOS, or Windows
- **Memory**: 512MB RAM minimum (2GB+ recommended)
- **Disk Space**: 100MB for binary + storage for your data
- **Network**: Port 8080 (configurable)

### Development Requirements
- **Go**: Version 1.21+ (for building from source)
- **Python**: Version 3.7+ (for Python client)

## 🚀 Releases & Distribution

VittoriaDB provides cross-platform binaries for all major platforms:

| Platform | Architecture | Status |
|----------|-------------|---------|
| **Linux** | AMD64/ARM64 | ✅ Available |
| **macOS** | Intel/Apple Silicon | ✅ Available |
| **Windows** | AMD64 | ✅ Available |

All releases are automatically built and published to [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases) with checksums and automated builds via GitHub Actions.

## 🧪 Development

### Building from Source
```bash
# Clone and build
git clone https://github.com/antonellof/VittoriaDB.git
cd VittoriaDB
go build -o vittoriadb ./cmd/vittoriadb

# Install Python SDK (optional)
cd sdk/python && ./install-dev.sh
```

### Testing
```bash
# Run Go tests
go test ./... -v

# Run Python tests
cd sdk/python && python -m pytest tests/ -v

# Test functionality
./vittoriadb run &
curl http://localhost:8080/health
```

> 📖 **See [Development Guide](docs/development.md) for complete build instructions, testing, debugging, and contribution guidelines.**

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Start for Contributors
1. Fork and clone the repository
2. Install Go 1.21+ and Python 3.7+
3. Create a feature branch
4. Make your changes and add tests
5. Submit a pull request

> 📖 **See [Development Guide](docs/development.md) for detailed setup, testing, and contribution workflows.**

## 📞 Support

- **📖 Documentation**: Complete guides in [`docs/`](docs/) directory
- **🐛 Issues**: [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues)
- **💬 Discussions**: [GitHub Discussions](https://github.com/antonellof/VittoriaDB/discussions)
- **📦 Releases**: [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases)

### Getting Help
1. Check the documentation in [`docs/`](docs/)
2. Search existing issues
3. Create an issue for bugs or feature requests
4. Start a discussion for questions

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

<div align="center">

**🚀 VittoriaDB - Making Vector Databases Local and Simple**

*Built with ❤️ for the AI community*

[![GitHub Stars](https://img.shields.io/github/stars/antonellof/VittoriaDB?style=social)](https://github.com/antonellof/VittoriaDB)
[![GitHub Forks](https://img.shields.io/github/forks/antonellof/VittoriaDB?style=social)](https://github.com/antonellof/VittoriaDB)

</div>
