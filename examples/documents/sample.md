---
title: "VittoriaDB Documentation"
author: "VittoriaDB Team"
date: "2025-09-09"
tags: ["vector-database", "ai", "embeddings"]
---

# VittoriaDB: Local Vector Database

## Overview

**VittoriaDB** is a high-performance, embedded vector database designed specifically for local AI development and production deployments.

## Features

### Core Capabilities
- **🎯 Zero Configuration**: Works immediately after installation
- **⚡ High Performance**: HNSW indexing for scalable similarity search
- **📁 Persistent Storage**: ACID-compliant file-based storage with WAL
- **🔌 Dual Interface**: REST API + Native Python client

### Advanced Features
- Multiple Index Types: Flat (exact) and HNSW (approximate) indexing
- Distance Metrics: Cosine, Euclidean, Dot Product, Manhattan
- Metadata Filtering: Rich query capabilities with JSON-based filters
- Batch Operations: Efficient bulk insert and search operations

## Quick Start

```bash
# 1. Start VittoriaDB
./vittoriadb run

# 2. Create a collection
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "docs", "dimensions": 384}'

# 3. Insert a vector
curl -X POST http://localhost:8080/collections/docs/vectors \
  -H "Content-Type: application/json" \
  -d '{"id": "doc1", "vector": [0.1, 0.2, ...], "metadata": {"title": "Test"}}'
```

## Document Processing

VittoriaDB supports processing various document formats:

| Format | Extension | Status |
|--------|-----------|---------|
| Plain Text | `.txt` | ✅ Implemented |
| Markdown | `.md` | ✅ Implemented |
| HTML | `.html` | ✅ Implemented |
| PDF | `.pdf` | 🚧 Placeholder |
| Word | `.docx` | 🚧 Placeholder |

## Architecture

VittoriaDB uses a modular architecture with separate layers for:

1. **HTTP API Server** - RESTful endpoints
2. **Vector Engine** - HNSW and flat indexing
3. **Storage Layer** - File-based persistence with WAL
4. **Document Processor** - Text extraction and chunking

## Performance

- **Insert Speed**: >10k vectors/second (flat index)
- **Search Speed**: <1ms for 1M vectors (HNSW)
- **Memory Usage**: <100MB for 100k vectors
- **Binary Size**: ~8MB compressed

## Use Cases

Perfect for:
- RAG (Retrieval-Augmented Generation) applications
- Semantic search systems
- Recommendation engines
- AI prototyping and development
- Edge computing deployments

---

*Built with ❤️ for the AI community*
