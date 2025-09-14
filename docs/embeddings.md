# VittoriaDB Embedding Services

VittoriaDB provides professional embedding services through external integrations, following industry best practices used by major vector databases.

## 🎯 Overview

Instead of implementing custom embedding algorithms, VittoriaDB delegates text vectorization to specialized external services. This approach ensures:

- **High-quality embeddings** from proven ML models
- **Industry-standard compatibility** with existing workflows
- **Maintainable codebase** without complex ML implementations
- **Flexible deployment options** for different environments

## 🤖 Supported Embedding Services

### 🔧 Ollama (Recommended)
**Local ML models without API dependencies**

```python
# Install: pip install vittoriadb
import vittoriadb
from vittoriadb.configure import Configure

db = vittoriadb.connect()

# Automatic embeddings using local Ollama models
collection = db.create_collection(
    name="documents",
    dimensions=768,
    vectorizer_config=Configure.Vectors.auto_embeddings()
)
```

**Features:**
- ✅ **High-quality ML embeddings** (comparable to cloud APIs)
- ✅ **No API costs** or rate limits
- ✅ **Works offline** completely
- ✅ **Fast inference** (~500ms per text)
- ✅ **Privacy-first** (data never leaves your machine)

**Setup:**
```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama service
ollama serve

# Pull embedding model
ollama pull nomic-embed-text

# Ready to use with VittoriaDB!
```

**Models Available:**
- `nomic-embed-text` (768 dims) - Default, high-quality general purpose
- `all-minilm` (384 dims) - Smaller, faster model
- `mxbai-embed-large` (1024 dims) - Larger, higher quality model

### 🤖 OpenAI API (Highest Quality)
**Cloud-based embeddings with state-of-the-art quality**

```python
# OpenAI embeddings (highest quality available)
collection = db.create_collection(
    name="documents",
    dimensions=1536,
    vectorizer_config=Configure.Vectors.openai_embeddings(
        model="text-embedding-ada-002",
        api_key="sk-your-openai-key"
    )
)
```

**Features:**
- ✅ **Highest quality embeddings** available
- ✅ **Proven at scale** (used by millions of applications)
- ✅ **Fast API responses** (~300ms)
- ✅ **Multiple model options** (ada-002, text-embedding-3-small, etc.)

**Setup:**
```bash
# Get API key from OpenAI
# https://platform.openai.com/api-keys

# Set environment variable (recommended)
export OPENAI_API_KEY='sk-your-actual-key'

# Or pass directly in code
Configure.Vectors.openai_embeddings(api_key="sk-your-key")
```

**Models Available:**
- `text-embedding-ada-002` (1536 dims) - Default, balanced quality/cost
- `text-embedding-3-small` (1536 dims) - Latest, improved quality
- `text-embedding-3-large` (3072 dims) - Highest quality, higher cost

### 🤗 HuggingFace API (Free Tier)
**Cloud-based embeddings with generous free tier**

```python
# HuggingFace embeddings (good quality, free tier)
collection = db.create_collection(
    name="documents",
    dimensions=384,
    vectorizer_config=Configure.Vectors.huggingface_embeddings(
        model="sentence-transformers/all-MiniLM-L6-v2",
        api_key="hf_your-token"
    )
)
```

**Features:**
- ✅ **Good quality embeddings** from proven models
- ✅ **Generous free tier** (30,000 requests/month)
- ✅ **Large model selection** (thousands of models available)
- ✅ **Open-source models** (transparent and reproducible)

**Setup:**
```bash
# Get API token from HuggingFace
# https://huggingface.co/settings/tokens

# Set environment variable
export HUGGINGFACE_API_KEY='hf_your-token'

# Or pass directly in code
Configure.Vectors.huggingface_embeddings(api_key="hf_your-token")
```

### 🐍 Sentence Transformers (Local Python)
**Local Python models with full control**

```python
# Local Python models (full control, heavy dependencies)
collection = db.create_collection(
    name="documents",
    dimensions=384,
    vectorizer_config=Configure.Vectors.sentence_transformers(
        model="all-MiniLM-L6-v2"
    )
)
```

**Features:**
- ✅ **Full local control** (no external dependencies)
- ✅ **Thousands of models** available via HuggingFace Hub
- ✅ **Customizable** (fine-tune models for your domain)
- ✅ **Works offline** completely

**Setup:**
```bash
# Install Python dependencies
pip install sentence-transformers

# Models download automatically on first use
```

## 🎯 auto_embeddings() Function

The `Configure.Vectors.auto_embeddings()` function is VittoriaDB's **flagship embedding configuration**, designed to provide the best balance of quality, performance, and ease of use.

### What is auto_embeddings()?

`auto_embeddings()` is an **intelligent embedding configuration** that:

1. **Uses Ollama by default** - Provides high-quality local ML embeddings
2. **Requires minimal setup** - Just `ollama pull nomic-embed-text`
3. **Works completely offline** - No API keys or internet required
4. **Provides real ML quality** - Not statistical approximations
5. **Costs nothing to run** - No per-request charges

### Why auto_embeddings()?

Traditional vector databases force you to choose between:
- **High quality** (expensive cloud APIs)
- **Local deployment** (complex model management)
- **Simple setup** (poor quality statistical methods)

`auto_embeddings()` gives you **all three**:

```python
# One line for high-quality local ML embeddings
vectorizer_config=Configure.Vectors.auto_embeddings()
```

### How it Works

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Client calls Configure.Vectors.auto_embeddings()        │
└─────────────────────┬───────────────────────────────────────┘
                      │ 
┌─────────────────────▼───────────────────────────────────────┐
│ 2. VittoriaDB configures Ollama vectorizer                 │
│    - Model: nomic-embed-text (768 dimensions)              │
│    - URL: http://localhost:11434                           │
│    - Type: Local ML (no API keys needed)                   │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│ 3. Text processing delegated to Ollama                     │
│    - Real neural network embeddings                        │
│    - Trained on massive text corpora                       │
│    - High semantic understanding                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│ 4. High-quality embeddings returned to VittoriaDB         │
│    - 768-dimensional dense vectors                         │
│    - Optimized for semantic similarity                     │
│    - Ready for storage and search                          │
└─────────────────────────────────────────────────────────────┘
```

### Comparison with Other Approaches

| Approach | Quality | Setup | Cost | Speed | Dependencies |
|----------|---------|-------|------|-------|--------------|
| **auto_embeddings()** | 🟢 High | 🟢 Simple | 🟢 Free | 🟢 Fast | Ollama only |
| **openai_embeddings()** | 🟢 Highest | 🟡 API Key | 🔴 Paid | 🟢 Fast | Internet + API |
| **sentence_transformers()** | 🟢 High | 🟡 Python | 🟢 Free | 🔴 Slow | Python + models |
| **Manual embeddings** | 🟡 Variable | 🔴 Complex | 🟡 Variable | 🟡 Variable | Client models |

## 📚 Usage Examples

### Basic Usage
```python
import vittoriadb
from vittoriadb.configure import Configure

# Connect to VittoriaDB
client = vittoriadb.connect(url="http://localhost:8080")

# Create collection with automatic embeddings
collection = client.create_collection(
    name="my_documents",
    dimensions=768,  # nomic-embed-text dimensions
    vectorizer_config=Configure.Vectors.auto_embeddings()
)

# Insert text - embeddings generated automatically
collection.insert_text("doc1", "Artificial intelligence transforms information processing")
collection.insert_text("doc2", "Machine learning enables pattern recognition in data")

# Search with text - query embedding generated automatically
results = collection.search_text("AI and pattern recognition", limit=5)
for result in results:
    print(f"Score: {result.score:.4f} | ID: {result.id}")
```

### Custom Model
```python
# Use different Ollama model
collection = client.create_collection(
    name="custom_docs",
    dimensions=384,
    vectorizer_config=Configure.Vectors.auto_embeddings(
        model="all-minilm",  # Smaller, faster model
        dimensions=384
    )
)
```

### Advanced Configuration
```python
# Custom Ollama configuration
collection = client.create_collection(
    name="advanced_docs",
    dimensions=768,
    vectorizer_config=Configure.Vectors.ollama_embeddings(
        model="nomic-embed-text",
        base_url="http://custom-ollama-server:11434",  # Custom Ollama server
        dimensions=768
    )
)
```

## 🔧 Troubleshooting

### Common Issues

**Error: "failed to connect to Ollama (is it running?)"**
```bash
# Solution: Start Ollama service
ollama serve

# Verify Ollama is running
curl http://localhost:11434/api/version
```

**Error: "model not found"**
```bash
# Solution: Pull the required model
ollama pull nomic-embed-text

# List available models
ollama list
```

**Error: "API request failed with status 401"**
```bash
# Solution: Check your API key
export OPENAI_API_KEY='sk-your-actual-key'
# OR
export HUGGINGFACE_API_KEY='hf_your-token'
```

### Performance Optimization

**For high-throughput applications:**
- Use **batch operations** when possible
- Consider **multiple Ollama instances** for parallel processing
- Use **connection pooling** for API-based vectorizers
- Monitor **rate limits** for cloud APIs

**For low-latency applications:**
- Use **Ollama local models** (fastest after warm-up)
- Configure **appropriate timeouts** for network-based services
- Consider **caching** for frequently used texts

## 🏆 Best Practices

### Model Selection
- **General purpose**: `nomic-embed-text` (768 dims)
- **Smaller/faster**: `all-minilm` (384 dims)
- **Highest quality**: OpenAI `text-embedding-ada-002` (1536 dims)
- **Domain-specific**: Choose specialized models from HuggingFace

### Production Deployment
- **Local deployment**: Use Ollama for cost-effective, high-quality embeddings
- **Cloud deployment**: Use OpenAI or HuggingFace APIs for managed infrastructure
- **Hybrid**: Use Ollama for development, cloud APIs for production
- **High-volume**: Consider dedicated Ollama servers or API rate limit management

### Security Considerations
- **API keys**: Store in environment variables, never in code
- **Local models**: Keep Ollama updated for security patches
- **Network**: Use HTTPS for production API calls
- **Data privacy**: Use local models (Ollama/Sentence Transformers) for sensitive data

## 📖 Further Reading

- **[API Reference](api.md)** - Complete REST API documentation
- **[Configuration Guide](configuration.md)** - Server and vectorizer configuration
- **[Performance Guide](performance.md)** - Benchmarks and optimization
- **[Examples](../examples/README.md)** - Comprehensive usage examples
