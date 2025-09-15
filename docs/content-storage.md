# Content Storage Enhancement

VittoriaDB has been enhanced to automatically store original text content alongside vector embeddings, making it ideal for RAG (Retrieval-Augmented Generation) systems.

## 🎯 **Problem Solved**

Previously, VittoriaDB followed the Pinecone model where original text was discarded after embedding generation. This required external storage systems for RAG applications. Now VittoriaDB follows the Weaviate/ChromaDB/Qdrant model with built-in content storage.

## 🚀 **New Features**

### **1. Content Storage Configuration**
```go
type ContentStorageConfig struct {
    Enabled    bool   // Whether to store original content
    FieldName  string // Metadata field name (default: "_content")
    MaxSize    int64  // Max content size in bytes (default: 1MB)
    Compressed bool   // Compression support (future)
}
```

### **2. Enhanced Collection Creation**
```bash
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{
    "name": "rag_documents",
    "dimensions": 384,
    "metric": 0,
    "index_type": 1,
    "content_storage": {
      "enabled": true,
      "field_name": "_content", 
      "max_size": 1048576,
      "compressed": false
    }
  }'
```

### **3. Automatic Content Preservation**
When using `InsertText()`, original content is now automatically stored:

```go
// Before: Content was lost
textVector := &TextVector{
    ID: "doc_1",
    Text: "Original content here...",
}
collection.InsertText(ctx, textVector) // Text discarded after embedding

// After: Content preserved automatically
textVector := &TextVector{
    ID: "doc_1", 
    Text: "Original content here...",
}
collection.InsertText(ctx, textVector) // Text stored in metadata["_content"]
```

### **4. Enhanced Search Results**
```go
// Search with content retrieval
req := &SearchRequest{
    Vector:         queryEmbedding,
    Limit:          5,
    IncludeContent: true,  // NEW: Retrieve original content
    IncludeMetadata: true,
}

results := collection.Search(ctx, req)
for _, result := range results.Results {
    fmt.Printf("Content: %s\n", result.Content) // Original text available
}
```

## 📊 **Comparison with Other Vector Databases**

| Database | Content Storage | RAG-Ready | External Storage Required |
|----------|----------------|-----------|---------------------------|
| **VittoriaDB (Enhanced)** | ✅ Built-in | ✅ Yes | ❌ No |
| Weaviate | ✅ Properties | ✅ Yes | ❌ No |
| ChromaDB | ✅ Documents | ✅ Yes | ❌ No |
| Qdrant | ✅ Payload | ✅ Yes | ❌ No |
| Pinecone | ⚠️ Limited Metadata | ⚠️ Partial | ✅ Yes (S3, etc.) |

## 🏗️ **Architecture Benefits**

### **Before (Pinecone Model)**
```
Application → VittoriaDB (vectors only)
           → External Storage (S3, DB) (original content)
           → Manual sync required
```

### **After (Weaviate/ChromaDB Model)**  
```
Application → VittoriaDB (vectors + content)
           → Single source of truth
           → Atomic operations
```

## 🔧 **Configuration Options**

### **Default Configuration**
```go
DefaultContentStorageConfig() = &ContentStorageConfig{
    Enabled:    true,           // Store content by default
    FieldName:  "_content",     // Standard field name
    MaxSize:    1048576,        // 1MB limit
    Compressed: false,          // No compression yet
}
```

### **Custom Configuration**
```go
// Disable content storage (classic mode)
config := &ContentStorageConfig{
    Enabled: false,
}

// Large content with custom field
config := &ContentStorageConfig{
    Enabled:   true,
    FieldName: "original_text",
    MaxSize:   10485760, // 10MB
}
```

## 🎯 **RAG System Integration**

### **Perfect for RAG Workflows**
```go
// 1. Store documents
for _, doc := range documents {
    textVector := &TextVector{
        ID:   doc.ID,
        Text: doc.Content,
        Metadata: map[string]interface{}{
            "title": doc.Title,
            "author": doc.Author,
        },
    }
    collection.InsertText(ctx, textVector)
}

// 2. Search with content
req := &SearchRequest{
    Vector:         queryEmbedding,
    Limit:          5,
    IncludeContent: true,  // Get original content for LLM
}
results, _ := collection.Search(ctx, req)

// 3. Build context for LLM
var context strings.Builder
for _, result := range results.Results {
    context.WriteString(fmt.Sprintf("Source: %s\nContent: %s\n\n", 
        result.Metadata["title"], result.Content))
}

// 4. Send to LLM
response := llm.Generate(userQuery, context.String())
```

## 🚀 **Migration Guide**

### **Existing Collections**
- Collections created before this enhancement will use default content storage settings
- New `InsertText()` operations will automatically start storing content
- Existing vectors remain unchanged

### **Backward Compatibility**
- All existing APIs work unchanged
- `include_content=false` (default) maintains original behavior
- Optional feature - can be disabled per collection

## 🎉 **Summary**

VittoriaDB now provides **best-in-class RAG support** with:

✅ **Automatic content storage** - No external systems needed  
✅ **Configurable limits** - Control storage usage  
✅ **Atomic operations** - Vector and content always in sync  
✅ **Fast retrieval** - Single query for similarity + content  
✅ **Standard compliance** - Follows Weaviate/ChromaDB patterns  
✅ **Backward compatible** - Existing code works unchanged  

This enhancement positions VittoriaDB as a **complete RAG solution** competitive with Weaviate, ChromaDB, and Qdrant while maintaining the performance benefits of a local vector database.
