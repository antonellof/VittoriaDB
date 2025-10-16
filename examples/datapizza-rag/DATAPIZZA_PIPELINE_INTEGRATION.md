# Datapizza AI + VittoriaDB Pipeline Integration

Complete RAG implementation using Datapizza AI's pipeline architecture with VittoriaDB as the vector database.

## 🎯 Overview

This integration provides a production-ready RAG system that combines:

- **Datapizza AI**: Modern Python framework for Gen AI solutions
- **VittoriaDB**: Zero-configuration embedded vector database with HNSW indexing
- **Pipeline Architecture**: Modular, composable RAG workflows

### Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                    Datapizza RAG Pipeline                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐        ┌─────────────────────┐      │
│  │ IngestionPipeline│        │    DagPipeline      │      │
│  └──────────────────┘        └─────────────────────┘      │
│           │                            │                    │
│           │                            │                    │
│  ┌────────▼─────────┐        ┌────────▼────────┐          │
│  │  NodeSplitter    │        │  ToolRewriter   │          │
│  │  ChunkEmbedder   │        │  OpenAIEmbedder │          │
│  └─────────┬────────┘        │  VittoriaDB     │          │
│            │                 │  PromptTemplate │          │
│            │                 │  OpenAIClient   │          │
│            │                 └─────────────────┘          │
│            │                                                │
│  ┌─────────▼──────────────────────────────────────┐       │
│  │       VittoriaDB Vectorstore Adapter           │       │
│  │                                                  │       │
│  │  - create_collection()                          │       │
│  │  - upsert()                                     │       │
│  │  - search()                                     │       │
│  └──────────────────────────────────────────────────┘      │
│                          │                                  │
└──────────────────────────┼──────────────────────────────────┘
                           │
                           ▼
              ┌─────────────────────┐
              │     VittoriaDB      │
              │   Vector Database   │
              │                     │
              │  - HNSW Indexing    │
              │  - ACID Storage     │
              │  - REST API         │
              └─────────────────────┘
```

## 🚀 Quick Start

### 1. Install Dependencies

The required Datapizza packages are included in `requirements.txt`:

```python
datapizza-ai
datapizza-ai-core
datapizza-ai-clients-openai
datapizza-ai-embedders-openai
datapizza-ai-modules-parsers
datapizza-ai-modules-splitters
datapizza-ai-modules-prompt
datapizza-ai-modules-rewriters
datapizza-ai-pipeline
```

### 2. Initialize the Pipeline

```python
from datapizza_rag_pipeline import create_datapizza_rag_pipeline

# Create pipeline with environment variables
rag = create_datapizza_rag_pipeline()

# Or with explicit configuration
rag = create_datapizza_rag_pipeline(
    openai_api_key="your-api-key",
    vittoriadb_url="http://vittoriadb:8080"
)
```

### 3. Create a Collection

```python
# Create a new collection
rag.create_collection("my_documents")

# Replace existing collection
rag.create_collection("my_documents", replace_existing=True)
```

### 4. Ingest Documents

```python
# Ingest plain text
rag.ingest_text(
    text="Your document content here...",
    collection_name="my_documents",
    metadata={"source": "manual", "title": "My Document"}
)

# Ingest a file
rag.ingest_file(
    file_path="path/to/document.txt",
    collection_name="my_documents",
    metadata={"source": "file_upload"}
)
```

### 5. Query the RAG System

#### Non-Streaming Query

```python
result = rag.query(
    question="What is Datapizza AI?",
    collection_name="my_documents",
    k=5,  # Number of chunks to retrieve
    rewrite_query=True  # Use query rewriting
)

print(f"Answer: {result['answer']}")
print(f"Chunks: {len(result['chunks'])}")
print(f"Rewritten Query: {result['rewritten_query']}")
```

#### Streaming Query

```python
async for chunk in rag.query_stream(
    question="How does VittoriaDB work?",
    collection_name="my_documents",
    k=5,
    rewrite_query=True
):
    if chunk['type'] == 'content':
        print(chunk['content'], end="", flush=True)
    elif chunk['type'] == 'chunks_retrieved':
        print(f"\nRetrieved {chunk['count']} chunks")
    elif chunk['type'] == 'done':
        print("\n✅ Done")
```

## 📦 Key Components

### VittoriaDB Vectorstore Adapter

`vittoriadb_vectorstore.py` - Implements the Datapizza vectorstore interface for VittoriaDB.

**Features:**
- Compatible with Datapizza `IngestionPipeline` and `DagPipeline`
- HNSW indexing for fast similarity search
- Metadata storage and filtering
- Configurable distance metrics (cosine, euclidean, dot product)

**Methods:**
- `create_collection(name, vector_config, metric, index_type)`
- `upsert(collection_name, chunks)`
- `search(collection_name, query_vector, k, filters)`
- `delete_collection(name)`

### Datapizza RAG Pipeline

`datapizza_rag_pipeline.py` - Complete RAG system using Datapizza's pipeline architecture.

**Features:**
- **IngestionPipeline**: Automatic document chunking and embedding
- **DagPipeline**: Complex retrieval workflows with query rewriting
- **Modular Design**: Easy to customize and extend
- **Streaming Support**: Real-time response generation

**Components:**
1. **NodeSplitter**: Splits documents into chunks with overlap
2. **ChunkEmbedder**: Generates embeddings for chunks
3. **ToolRewriter**: Rewrites queries for better retrieval
4. **ChatPromptTemplate**: Formats prompts with context
5. **OpenAIClient**: Generates responses

## 🔧 Configuration

Environment variables for customization:

```bash
# OpenAI Configuration
OPENAI_API_KEY=your-api-key
OPENAI_EMBED_MODEL=text-embedding-ada-002
OPENAI_EMBED_DIMENSIONS=1536
LLM_MODEL=gpt-4o-mini

# VittoriaDB Configuration
VITTORIADB_URL=http://vittoriadb:8080

# Chunking Configuration
CHUNK_SIZE=1000
CHUNK_OVERLAP=200
RETRIEVAL_K=5
```

## 🎓 Advanced Usage

### Custom Ingestion Pipeline

```python
from datapizza.pipeline import IngestionPipeline
from datapizza.modules.splitters import NodeSplitter
from datapizza.embedders import ChunkEmbedder

# Create custom pipeline
pipeline = IngestionPipeline(
    modules=[
        NodeSplitter(max_char=500, overlap=100),  # Smaller chunks
        ChunkEmbedder(client=embedder),
    ],
    vector_store=vectorstore,
    collection_name="my_collection"
)

pipeline.run("document.txt")
```

### Custom Retrieval Pipeline

```python
from datapizza.pipeline import DagPipeline

# Create DAG pipeline
dag = DagPipeline()

# Add modules
dag.add_module("rewriter", query_rewriter)
dag.add_module("embedder", embedder)
dag.add_module("retriever", vectorstore)
dag.add_module("generator", llm_client)

# Connect modules
dag.connect("rewriter", "embedder", target_key="text")
dag.connect("embedder", "retriever", target_key="query_vector")
dag.connect("retriever", "generator", target_key="context")

# Run pipeline
result = dag.run({
    "rewriter": {"user_prompt": "question"},
    "retriever": {"collection_name": "docs", "k": 5},
    "generator": {"input": "question"}
})
```

### Adding Custom Parsers

```python
# For PDF documents (requires datapizza-ai-parsers-docling)
from datapizza.modules.parsers.docling import DoclingParser

pipeline = IngestionPipeline(
    modules=[
        DoclingParser(),  # Parse PDF/DOCX
        NodeSplitter(max_char=1000),
        ChunkEmbedder(client=embedder),
    ],
    vector_store=vectorstore,
    collection_name="documents"
)
```

## 🧪 Example

See `example_datapizza_pipeline_usage.py` for a complete example demonstrating:

1. Pipeline initialization
2. Collection creation
3. Document ingestion
4. Non-streaming queries
5. Streaming queries

Run the example:

```bash
cd backend
python example_datapizza_pipeline_usage.py
```

## 🔗 Integration with Existing Code

The Datapizza pipeline can coexist with the existing RAG system in `rag_system.py`:

- **Legacy endpoints**: Continue using `rag_system.py` for backward compatibility
- **New endpoints**: Use `datapizza_rag_pipeline.py` for new features
- **Gradual migration**: Replace components one at a time

## 📚 Documentation

- **Datapizza AI**: https://datapizza.tech/en/ai-framework/
- **Datapizza RAG Guide**: https://docs.datapizza.ai/0.0.2/Guides/RAG/rag/
- **VittoriaDB**: https://github.com/antonellof/VittoriaDB

## 🎯 Benefits

### Datapizza AI Pipeline Architecture

✅ **Modular**: Easy to add/remove/replace components  
✅ **Composable**: Build complex workflows with simple modules  
✅ **Testable**: Each module can be tested independently  
✅ **Production-Ready**: Built for real-world applications  
✅ **Extensible**: Create custom modules and pipelines  

### VittoriaDB Integration

✅ **Zero Configuration**: No complex setup required  
✅ **HNSW Indexing**: Fast similarity search  
✅ **ACID Storage**: Reliable data persistence  
✅ **Lightweight**: Single binary, no dependencies  
✅ **REST API**: Easy integration  

## 🚧 Roadmap

- [ ] Add support for more document parsers (PDF, DOCX, HTML)
- [ ] Implement advanced retrieval strategies (hybrid search, reranking)
- [ ] Add evaluation metrics for RAG quality
- [ ] Support for multi-modal embeddings (text + images)
- [ ] Distributed deployment support

## 🤝 Contributing

Contributions are welcome! This integration demonstrates how to:

1. Create custom Datapizza vectorstore adapters
2. Build production RAG pipelines
3. Integrate with existing vector databases

## 📄 License

This integration follows the same license as the parent VittoriaDB project.

---

**Built with ❤️ using:**
- 🍕 [Datapizza AI](https://datapizza.tech) - Modern Gen AI Framework
- 🗄️ [VittoriaDB](https://github.com/antonellof/VittoriaDB) - Embedded Vector Database
- 🤖 [OpenAI](https://openai.com) - LLM & Embeddings

