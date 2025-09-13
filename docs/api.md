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
        "category": "technology"
      }
    }
  ],
  "total": 100,
  "took_ms": 5,
  "request_id": "req_123456"
}
```

## 📄 Document Upload (Future Feature)

```bash
curl -X POST http://localhost:8080/collections/documents/upload \
  -F "file=@document.pdf" \
  -F "chunk_size=500" \
  -F "overlap=50" \
  -F "embedding_model=sentence-transformers/all-MiniLM-L6-v2" \
  -F "metadata={\"source\": \"upload\", \"type\": \"pdf\"}"
```

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
