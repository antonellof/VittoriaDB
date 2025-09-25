# API Documentation

VittoriaDB provides a comprehensive REST API for all vector database operations, along with native Go and Python SDKs.

## 🌐 REST API

### Base URL
```
http://localhost:8080
```

### Authentication
Currently, VittoriaDB runs without authentication. Authentication features are planned for future releases.

## 📋 API Endpoints Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/stats` | Database statistics |
| `GET` | `/config` | **NEW!** Current configuration |
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
| `POST` | `/collections/{name}/text` | Insert text (auto-vectorized) |
| `POST` | `/collections/{name}/text/batch` | Batch insert text |
| `GET,POST` | `/collections/{name}/search/text` | Search with text query |
| `POST` | `/collections/{name}/upload` | Upload document |

## 🔧 Server Management

### Health Check
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy",
  "uptime": 3600,
  "collections": 2,
  "total_vectors": 1000,
  "memory_usage": 52428800,
  "disk_usage": 1048576
}
```

### Database Statistics
```bash
curl http://localhost:8080/stats
```

**Response:**
```json
{
  "collections": [
    {
      "name": "documents",
      "vector_count": 500,
      "dimensions": 384,
      "index_type": "hnsw",
      "index_size": 524288,
      "last_modified": "2025-09-13T10:30:00Z"
    }
  ],
  "total_vectors": 1000,
  "total_size": 1048576,
  "queries_total": 42,
  "avg_query_latency": 1.5
}
```

### Configuration Inspection (NEW!)
```bash
curl http://localhost:8080/config
```

**Response:**
```json
{
  "config": {
    "server": {
      "host": "localhost",
      "port": 8080,
      "read_timeout": 30000000000,
      "write_timeout": 30000000000,
      "max_body_size": 33554432,
      "cors": true,
      "tls": {
        "enabled": false,
        "cert_file": "",
        "key_file": ""
      }
    },
    "storage": {
      "engine": "file",
      "page_size": 4096,
      "cache_size": 1000,
      "sync_writes": true
    },
    "search": {
      "parallel": {
        "enabled": true,
        "max_workers": 10,
        "batch_size": 100,
        "preload_vectors": false
      },
      "cache": {
        "enabled": true,
        "max_entries": 1000,
        "ttl": 300000000000,
        "cleanup_interval": 60000000000
      },
      "index": {
        "default_type": "flat",
        "default_metric": "cosine",
        "hnsw": {
          "m": 16,
          "ef_construction": 100,
          "ef_search": 100
        }
      }
    },
    "embeddings": {
      "default": {
        "type": "sentence_transformers",
        "model": "all-MiniLM-L6-v2",
        "dimensions": 384
      },
      "batch": {
        "enabled": true,
        "batch_size": 32,
        "timeout": 30000000000
      }
    },
    "performance": {
      "max_concurrency": 20,
      "enable_simd": true,
      "memory_limit": 2147483648,
      "io": {
        "use_memory_map": true,
        "async_io": true,
        "vectorized_ops": true
      }
    },
    "log": {
      "level": "info",
      "format": "text",
      "output": "stdout"
    },
    "data_dir": "data"
  },
  "features": {
    "parallel_search": true,
    "search_cache": true,
    "memory_mapped_io": true,
    "simd_optimizations": true,
    "async_io": true
  },
  "performance": {
    "max_workers": 10,
    "cache_entries": 1000,
    "cache_ttl": "5m0s",
    "max_concurrency": 20,
    "memory_limit_mb": 2048
  },
  "metadata": {
    "source": "default",
    "loaded_at": "2025-09-25T13:51:07+02:00",
    "version": "v1",
    "description": "VittoriaDB unified configuration"
  }
}
```

**Use Cases:**
- **Debugging**: Inspect current configuration settings
- **Monitoring**: Check feature flags and performance settings
- **Validation**: Verify configuration is loaded correctly
- **Documentation**: Self-documenting API with current settings

**Query Examples:**
```bash
# Get specific configuration section
curl -s http://localhost:8080/config | jq '.config.performance'

# Check if parallel search is enabled
curl -s http://localhost:8080/config | jq '.features.parallel_search'

# Get performance metrics
curl -s http://localhost:8080/config | jq '.performance'

# Check configuration source and load time
curl -s http://localhost:8080/config | jq '.metadata'
```

## 📚 Collection Management

### List Collections
```bash
curl http://localhost:8080/collections
```

**Response:**
```json
{
  "collections": [
    {
      "name": "documents",
      "dimensions": 384,
      "metric": "cosine",
      "index_type": "hnsw",
      "vector_count": 500,
      "created": "2025-09-13T10:00:00Z",
      "modified": "2025-09-13T10:30:00Z"
    }
  ],
  "count": 1
}
```

### Create Collection
```bash
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "documents",
    "dimensions": 384,
    "metric": 0,
    "index_type": 1
  }'
```

**Parameters:**
- `name`: Collection name (string)
- `dimensions`: Vector dimensions (integer)
- `metric`: Distance metric (integer: 0=cosine, 1=euclidean, 2=dot_product, 3=manhattan)
- `index_type`: Index type (integer: 0=flat, 1=hnsw, 2=ivf)
- `config`: Optional configuration object

**Advanced Collection Creation:**
```bash
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "large_docs",
    "dimensions": 1536,
    "metric": 0,
    "index_type": 1,
    "config": {
      "m": 32,
      "ef_construction": 400,
      "ef_search": 50
    }
  }'
```

**Collection with Content Storage (RAG-Optimized):**
```bash
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "rag_documents",
    "dimensions": 384,
    "metric": 0,
    "index_type": 1,
    "vectorizer_config": {
      "type": "sentence_transformers",
      "model": "all-MiniLM-L6-v2"
    },
    "content_storage": {
      "enabled": true,
      "field_name": "_content",
      "max_size": 1048576,
      "compressed": false
    }
  }'
```

**Content Storage Configuration:**
- `enabled` (bool): Store original text content alongside vectors (default: true)
- `field_name` (string): Metadata field name for content (default: "_content")  
- `max_size` (int64): Maximum content size in bytes, 0 = unlimited (default: 1MB)
- `compressed` (bool): Compress content to save space (default: false)

### Get Collection Information
```bash
curl http://localhost:8080/collections/documents
```

### Get Collection Statistics
```bash
curl http://localhost:8080/collections/documents/stats
```

### Delete Collection
```bash
curl -X DELETE http://localhost:8080/collections/documents
```

## 🎯 Vector Operations

### Insert Single Vector
```bash
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
```

### Batch Insert Vectors
```bash
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
```

### Get Vector
```bash
curl http://localhost:8080/collections/documents/vectors/doc_001
```

### Delete Vector
```bash
curl -X DELETE http://localhost:8080/collections/documents/vectors/doc_001
```

## 🔍 Vector Search

### Basic Similarity Search
```bash
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10'
```

### Search with Metadata
```bash
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=5' \
  --data-urlencode 'include_metadata=true'
```

### Search with Filters
```bash
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10' \
  --data-urlencode 'filter={"category": "technology"}'
```

### Advanced Search with Multiple Filters
```bash
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10' \
  --data-urlencode 'filter={"category": "technology", "author": "John Doe"}'
```

### Search with Pagination
```bash
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=10' \
  --data-urlencode 'offset=20'
```

### Search with Original Content (RAG-Optimized)
```bash
curl -G http://localhost:8080/collections/documents/search \
  --data-urlencode 'vector=[0.1,0.2,0.3,0.4]' \
  --data-urlencode 'limit=5' \
  --data-urlencode 'include_content=true' \
  --data-urlencode 'include_metadata=true'
```

**Search Response:**
```json
{
  "results": [
    {
      "id": "doc_001",
      "score": 0.95,
      "vector": [0.1, 0.2, 0.3, 0.4],
      "metadata": {
        "title": "Introduction to AI",
        "category": "technology",
        "_content": "Artificial intelligence is transforming..."
      },
      "content": "Artificial intelligence is transforming how we process data through machine learning algorithms..."
    }
  ],
  "total": 100,
  "took_ms": 5,
  "request_id": "req_123456"
}
```

**Search Parameters:**
- `include_content` (bool): Include original text content in results (requires content storage enabled)

## 🤖 RAG (Retrieval-Augmented Generation) Support

VittoriaDB now includes built-in support for RAG systems by automatically storing original text content alongside vector embeddings. This eliminates the need for external content storage and provides seamless integration with LLMs.

**Benefits:**
- ✅ **No External Storage Required**: Original content stored directly in VittoriaDB
- ✅ **Atomic Operations**: Vector and content always in sync
- ✅ **Fast Retrieval**: Single query returns both similarity scores and original text
- ✅ **Configurable**: Adjustable content limits and storage options
- ✅ **RAG-Ready**: Perfect for feeding context to language models

**Example RAG Workflow:**
1. Store documents with `InsertText()` → Automatically generates embeddings + stores content
2. Search with `include_content=true` → Returns relevant text passages
3. Feed retrieved content to LLM → Generate contextual responses

## 🔤 Text Operations (Auto-Vectorized)

For collections with vectorizer configuration, you can insert and search text directly without manually generating embeddings.

### Insert Single Text
```bash
curl -X POST http://localhost:8080/collections/documents/text \
  -H "Content-Type: application/json" \
  -d '{
    "id": "text_001",
    "text": "Artificial intelligence is transforming how we process data through machine learning.",
    "metadata": {
      "category": "technology",
      "source": "article",
      "date": "2024-01-15"
    }
  }'
```

**Requirements:** Collection must have `vectorizer_config` enabled.

**Response:**
```json
{
  "status": "inserted",
  "id": "text_001",
  "embedding_generated": true
}
```

### Batch Insert Text
```bash
curl -X POST http://localhost:8080/collections/documents/text/batch \
  -H "Content-Type: application/json" \
  -d '{
    "texts": [
      {
        "id": "text_002",
        "text": "Machine learning algorithms can identify patterns in large datasets.",
        "metadata": {"category": "AI", "type": "definition"}
      },
      {
        "id": "text_003", 
        "text": "Vector databases enable efficient similarity search for AI applications.",
        "metadata": {"category": "database", "type": "explanation"}
      }
    ]
  }'
```

**Response:**
```json
{
  "status": "batch_inserted",
  "inserted_count": 2,
  "failed_count": 0,
  "processing_time": 1250
}
```

### Text Search
Search using natural language queries (automatically vectorized):

```bash
# GET method
curl -G http://localhost:8080/collections/documents/search/text \
  --data-urlencode 'query=machine learning and artificial intelligence' \
  --data-urlencode 'limit=5'

# POST method (recommended for complex queries)
curl -X POST http://localhost:8080/collections/documents/search/text \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning and artificial intelligence",
    "limit": 5,
    "filter": {"category": "technology"}
  }'
```

**Requirements:** Collection must have `vectorizer_config` enabled.

**Response:**
```json
{
  "results": [
    {
      "id": "text_001",
      "score": 0.89,
      "metadata": {
        "category": "technology",
        "source": "article"
      }
    }
  ],
  "query_embedding_generated": true,
  "took_ms": 45
}
```

## 📄 Document Upload

VittoriaDB supports intelligent document processing with automatic vectorization. Upload documents and they are automatically processed, chunked, and vectorized based on your collection's configuration.

### Supported Document Formats
- **PDF** - Text extraction from PDF documents
- **DOCX** - Microsoft Word documents
- **TXT** - Plain text files
- **MD** - Markdown files (with frontmatter parsing)
- **HTML** - HTML documents (with tag stripping)

### Upload Document
```bash
curl -X POST http://localhost:8080/collections/documents/upload \
  -F "file=@document.pdf" \
  -F "chunk_size=500" \
  -F "chunk_overlap=50" \
  -F "language=en" \
  -F "metadata={\"source\": \"upload\", \"type\": \"pdf\", \"author\": \"user\"}"
```

**Parameters:**
- `file` (required): The document file to upload (max 32MB)
- `chunk_size` (optional): Size of text chunks in characters (default: 500)
- `chunk_overlap` (optional): Overlap between chunks in characters (default: 50)
- `language` (optional): Document language for processing (default: "en")
- `metadata` (optional): Additional metadata as JSON object

**Response:**
```json
{
  "status": "processed",
  "document_id": "doc_1694678400123",
  "document_title": "Research Paper.pdf",
  "document_type": "pdf",
  "chunks_created": 15,
  "chunks_inserted": 15,
  "processing_time": 2340,
  "collection": "documents"
}
```

### Automatic vs Manual Vectorization

The upload behavior depends on your collection configuration:

#### Collections WITH Vectorizer (Recommended)
For collections created with `vectorizer_config`, documents are automatically vectorized:

```bash
# 1. Create collection with vectorizer
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "smart_documents",
    "dimensions": 384,
    "metric": 0,
    "vectorizer_config": {
      "type": "sentence_transformers",
      "model": "all-MiniLM-L6-v2",
      "dimensions": 384
    }
  }'

# 2. Upload document - automatic embedding generation
curl -X POST http://localhost:8080/collections/smart_documents/upload \
  -F "file=@research.pdf" \
  -F "chunk_size=600"
```

**What happens:**
1. 📄 Document is processed and chunked
2. 🤖 Each chunk is automatically vectorized using the collection's vectorizer
3. 💾 Real embeddings are stored and ready for semantic search

#### Collections WITHOUT Vectorizer (Legacy Mode)
For collections without `vectorizer_config`, placeholder vectors are created:

```bash
# 1. Create basic collection (no vectorizer)
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "basic_documents",
    "dimensions": 384,
    "metric": 0
  }'

# 2. Upload document - placeholder vectors only
curl -X POST http://localhost:8080/collections/basic_documents/upload \
  -F "file=@research.pdf"
```

**What happens:**
1. 📄 Document is processed and chunked
2. 🚧 Zero vectors (placeholders) are created for each chunk
3. 💾 Text content and metadata are preserved, but no semantic search capability

### Available Vectorizer Types

| Type | Model | Dimensions | Requirements |
|------|-------|------------|--------------|
| `sentence_transformers` | `all-MiniLM-L6-v2` | 384 | Server-side Python environment |
| `openai` | `text-embedding-ada-002` | 1536 | OpenAI API key |
| `huggingface` | Various models | 384+ | HuggingFace API token (optional) |
| `ollama` | `nomic-embed-text` | 768 | Local Ollama installation |

### Complete Workflow Example

```bash
#!/bin/bash
# Complete document upload workflow

# 1. Create collection with automatic embeddings
echo "Creating collection with vectorizer..."
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "knowledge_base",
    "dimensions": 384,
    "metric": 0,
    "index_type": 1,
    "vectorizer_config": {
      "type": "sentence_transformers",
      "model": "all-MiniLM-L6-v2",
      "dimensions": 384
    }
  }' | jq

# 2. Upload multiple documents
echo "Uploading PDF document..."
curl -X POST http://localhost:8080/collections/knowledge_base/upload \
  -F "file=@research_paper.pdf" \
  -F "chunk_size=600" \
  -F "chunk_overlap=100" \
  -F "metadata={\"category\": \"research\", \"year\": \"2024\"}" | jq

echo "Uploading Word document..."
curl -X POST http://localhost:8080/collections/knowledge_base/upload \
  -F "file=@manual.docx" \
  -F "chunk_size=400" \
  -F "metadata={\"category\": \"documentation\"}" | jq

# 3. Search the uploaded content (semantic search)
echo "Searching uploaded documents..."
curl -X POST http://localhost:8080/collections/knowledge_base/search/text \
  -H "Content-Type: application/json" \
  -d '{
    "query": "machine learning algorithms",
    "limit": 5
  }' | jq

# 4. Check collection statistics
echo "Collection statistics:"
curl http://localhost:8080/collections/knowledge_base/stats | jq
```

### Error Handling

| HTTP Code | Error | Description |
|-----------|-------|-------------|
| `400` | Bad Request | Invalid file format, missing file, or malformed parameters |
| `404` | Not Found | Collection does not exist |
| `413` | Payload Too Large | File exceeds 32MB limit |
| `415` | Unsupported Media Type | File format not supported |
| `500` | Internal Server Error | Processing failed or vectorizer error |

**Error Response Example:**
```json
{
  "error": "Unsupported document type",
  "code": "UNSUPPORTED_FORMAT",
  "details": {
    "filename": "document.xyz",
    "supported_formats": ["pdf", "docx", "txt", "md", "html"]
  }
}
```

### Best Practices

1. **Use Collections with Vectorizers**: Always create collections with `vectorizer_config` for automatic embedding generation
2. **Optimize Chunk Size**: 
   - **Small chunks (300-500)**: Better for precise matching
   - **Large chunks (800-1200)**: Better for context preservation
3. **Add Meaningful Metadata**: Include source, category, date, author for better filtering
4. **Monitor Processing Time**: Large documents may take several seconds to process
5. **Batch Upload**: For multiple documents, upload them sequentially to avoid overwhelming the server

## 🐍 Python SDK

### Connection
```python
import vittoriadb

# Connect to running server
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
```

### Collection Operations
```python
# Create collection
collection = db.create_collection(
    name="documents",
    dimensions=384,
    metric="cosine",
    index_type="hnsw"
)

# List collections
collections = db.list_collections()

# Get collection
collection = db.get_collection("documents")

# Delete collection
db.delete_collection("documents")
```

### Vector Operations
```python
# Insert single vector
success, error = collection.insert(
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

# Get vector
vector = collection.get("doc1")

# Delete vector
collection.delete("doc1")
```

## 🔧 Go SDK

### Database Operations
```go
import (
    "context"
    "github.com/antonellof/VittoriaDB/pkg/core"
)

// Create database
db := core.NewDatabase()
ctx := context.Background()

config := &core.Config{
    DataDir: "./my-vectors",
    Server: core.ServerConfig{
        Host: "localhost",
        Port: 8080,
    },
}

// Open database
if err := db.Open(ctx, config); err != nil {
    panic(err)
}
defer db.Close()
```

### Collection Operations
```go
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
```

### Vector Operations
```go
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
```

## 📊 Data Types

### Distance Metrics
- `0` - Cosine similarity
- `1` - Euclidean distance
- `2` - Dot product
- `3` - Manhattan distance

### Index Types
- `0` - Flat (exact search)
- `1` - HNSW (approximate search)
- `2` - IVF (inverted file, planned)

### Vector Format
```json
{
  "id": "unique_identifier",
  "vector": [0.1, 0.2, 0.3, 0.4],
  "metadata": {
    "key": "value",
    "numeric": 123,
    "boolean": true
  }
}
```

## 🚨 Error Handling

### HTTP Status Codes
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `404` - Not Found
- `409` - Conflict (e.g., collection already exists)
- `500` - Internal Server Error

### Error Response Format
```json
{
  "error": "Collection 'documents' already exists",
  "code": "COLLECTION_EXISTS",
  "details": {
    "collection": "documents",
    "suggestion": "Use a different name or delete the existing collection"
  }
}
```

## 🔄 Complete Workflow Example

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
  -d '{"name": "example", "dimensions": 4, "metric": 0, "index_type": 1}' | jq

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
