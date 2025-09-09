#!/usr/bin/env python3
"""
Document Processing Example with VittoriaDB

This example demonstrates how to:
1. Process various document formats (TXT, MD, HTML, PDF, DOCX)
2. Upload documents to VittoriaDB collections
3. Search through processed documents
4. Handle document metadata and chunking

Requirements:
    pip install requests sentence-transformers

Usage:
    python examples/document_processing_example.py
"""

import os
import json
from typing import List, Dict, Any, Optional
from pathlib import Path
import vittoriadb

class VittoriaDBDocumentProcessor:
    """Document processing client for VittoriaDB."""
    
    def __init__(self):
        self.db = None
        self.collections = {}
    
    def connect(self) -> bool:
        """Connect to VittoriaDB."""
        try:
            # Connect to existing VittoriaDB server (don't auto-start)
            self.db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
            return True
        except Exception as e:
            print(f"Failed to connect to VittoriaDB: {e}")
            return False
    
    def health_check(self) -> bool:
        """Check if VittoriaDB is running."""
        try:
            if not self.db:
                return False
            health = self.db.health()
            return health.status == "healthy"
        except:
            return False
    
    def get_supported_formats(self) -> Dict[str, Any]:
        """Get supported document formats."""
        # This would need to be implemented in the Python client
        # For now, return the known formats
        return {
            "supported_formats": [
                {"type": "txt", "description": "Plain text files", "status": "fully_implemented"},
                {"type": "md", "description": "Markdown files", "status": "fully_implemented"},
                {"type": "html", "description": "HTML files", "status": "fully_implemented"},
                {"type": "pdf", "description": "PDF documents", "status": "fully_implemented"},
                {"type": "docx", "description": "Microsoft Word documents", "status": "fully_implemented"}
            ]
        }
    
    def create_collection(self, name: str, dimensions: int = 384, metric: str = "cosine") -> bool:
        """Create a collection for documents."""
        try:
            metric_enum = getattr(vittoriadb.DistanceMetric, metric.upper(), vittoriadb.DistanceMetric.COSINE)
            collection = self.db.create_collection(
                name=name,
                dimensions=dimensions,
                metric=metric_enum
            )
            self.collections[name] = collection
            return True
        except vittoriadb.CollectionError as e:
            if "already exists" in str(e):
                collection = self.db.get_collection(name)
                self.collections[name] = collection
                return True
            return False
        except Exception:
            return False
    
    def process_document(self, file_path: str, chunk_size: int = 500, 
                        chunk_overlap: int = 50) -> Optional[Dict[str, Any]]:
        """Process a document without storing it."""
        if not os.path.exists(file_path):
            print(f"❌ File not found: {file_path}")
            return None
        
        # For this example, we'll simulate document processing
        # In a real implementation, this would use the document processing API
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            content = f.read()
        
        # Simple chunking
        chunks = []
        for i in range(0, len(content), chunk_size):
            chunk = content[i:i + chunk_size]
            if chunk.strip():
                chunks.append(chunk.strip())
        
        return {
            "title": os.path.basename(file_path),
            "type": os.path.splitext(file_path)[1][1:],
            "content": content,
            "chunks": chunks,
            "metadata": {
                "filename": os.path.basename(file_path),
                "size": len(content),
                "chunks_count": len(chunks)
            }
        }
    
    def upload_document(self, collection_name: str, file_path: str, 
                       chunk_size: int = 500, chunk_overlap: int = 50,
                       metadata: Optional[Dict[str, str]] = None) -> Optional[Dict[str, Any]]:
        """Upload and process a document to a collection."""
        # Process the document first
        doc_data = self.process_document(file_path, chunk_size, chunk_overlap)
        if not doc_data:
            return None
        
        if collection_name not in self.collections:
            print(f"❌ Collection '{collection_name}' not found")
            return None
        
        collection = self.collections[collection_name]
        
        # Insert chunks as vectors (using dummy embeddings for this example)
        chunks_inserted = 0
        for i, chunk in enumerate(doc_data["chunks"]):
            try:
                # In a real implementation, you'd generate embeddings here
                dummy_vector = [0.1] * 384  # Dummy 384-dimensional vector
                
                chunk_metadata = {
                    "content": chunk,
                    "document": doc_data["title"],
                    "chunk_index": i,
                    **(metadata or {})
                }
                
                collection.insert(
                    id=f"{doc_data['title']}_chunk_{i}",
                    vector=dummy_vector,
                    metadata=chunk_metadata
                )
                chunks_inserted += 1
            except Exception as e:
                print(f"❌ Failed to insert chunk {i}: {e}")
        
        return {
            "document_id": doc_data["title"],
            "chunks_created": len(doc_data["chunks"]),
            "chunks_inserted": chunks_inserted
        }
    
    def search_collection(self, collection_name: str, query_vector: List[float], 
                         limit: int = 5) -> List[Dict[str, Any]]:
        """Search a collection with a query vector."""
        if collection_name not in self.collections:
            return []
        
        try:
            collection = self.collections[collection_name]
            results = collection.search(
                vector=query_vector,
                limit=limit,
                include_metadata=True
            )
            
            # Convert to the expected format
            return [
                {
                    "id": result.id,
                    "score": result.score,
                    "metadata": result.metadata or {}
                }
                for result in results
            ]
        except Exception:
            return []
    
    def get_collection_info(self, collection_name: str) -> Optional[Dict[str, Any]]:
        """Get collection information."""
        try:
            if collection_name in self.collections:
                collection = self.collections[collection_name]
                info = collection.info
                return {
                    "name": info.name,
                    "dimensions": info.dimensions,
                    "metric": info.metric.value,
                    "index_type": info.index_type.value,
                    "vector_count": info.vector_count
                }
            return None
        except Exception:
            return None

def create_sample_documents():
    """Create sample documents for testing."""
    docs_dir = Path("examples/documents/samples")
    docs_dir.mkdir(exist_ok=True)
    
    # Create sample text document
    with open(docs_dir / "sample.txt", "w") as f:
        f.write("""VittoriaDB: A Simple Vector Database

VittoriaDB is a high-performance, embedded vector database designed specifically for local AI development and production deployments. Built with simplicity and performance in mind, it provides a zero-configuration solution for vector similarity search.

Key Features:
- Zero Configuration: Works immediately after installation
- High Performance: HNSW indexing for scalable similarity search
- Persistent Storage: ACID-compliant file-based storage with WAL
- Dual Interface: REST API + Native Python client
- Document Processing: Support for PDF, DOCX, TXT, MD, HTML files
- AI-Ready: Seamless integration with embedding models

Use Cases:
VittoriaDB is perfect for RAG applications, semantic search, recommendation systems, and AI prototyping. It bridges the gap between complex cloud-based vector databases and simple in-memory solutions.

Getting Started:
1. Download the binary from GitHub releases
2. Run: ./vittoriadb run
3. Create a collection via REST API
4. Insert vectors and start searching

The database stores all data locally in configurable directories, making it perfect for development, testing, and edge deployments where you need full control over your data.""")
    
    # Create sample markdown document
    with open(docs_dir / "features.md", "w") as f:
        f.write("""---
title: "VittoriaDB Features"
author: "VittoriaDB Team"
date: "2025-09-09"
tags: ["vector-database", "ai", "features"]
---

# VittoriaDB Features

## Core Capabilities

### 🎯 Zero Configuration
Works immediately after installation with sensible defaults. No complex setup, no Docker required, no cloud dependencies.

### ⚡ High Performance
- HNSW indexing provides sub-millisecond search times
- SIMD optimizations for distance calculations
- Efficient memory management and caching

### 📁 Persistent Storage
- ACID-compliant file-based storage
- Write-Ahead Log (WAL) for durability
- Crash recovery and data integrity

### 🔌 Dual Interface
- REST API for universal access
- Native Python client with auto-binary management

## Advanced Features

### Multiple Index Types
- **Flat Index**: Exact similarity search with 100% recall
- **HNSW Index**: Approximate search with sub-linear time complexity

### Distance Metrics
- Cosine similarity
- Euclidean distance
- Dot product
- Manhattan distance

### Document Processing
- PDF text extraction
- DOCX document parsing
- HTML tag stripping
- Markdown with frontmatter
- Intelligent text chunking

## Performance Metrics

| Metric | Performance |
|--------|-------------|
| Insert Speed | >10k vectors/second |
| Search Speed | <1ms for 1M vectors |
| Memory Usage | <100MB for 100k vectors |
| Binary Size | ~8MB compressed |

## Use Cases

Perfect for:
- RAG (Retrieval-Augmented Generation) applications
- Semantic search systems
- Recommendation engines
- AI prototyping and development
- Edge computing deployments""")
    
    # Create sample HTML document
    with open(docs_dir / "getting_started.html", "w") as f:
        f.write("""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="description" content="Getting started with VittoriaDB vector database">
    <meta name="keywords" content="vector database, AI, getting started, tutorial">
    <title>Getting Started with VittoriaDB</title>
</head>
<body>
    <h1>Getting Started with VittoriaDB</h1>
    
    <h2>Quick Installation</h2>
    <p>VittoriaDB can be installed in multiple ways:</p>
    
    <h3>Option 1: Pre-built Binaries</h3>
    <p>Download from GitHub releases and run immediately:</p>
    <pre><code>wget https://github.com/antonellof/VittoriaDB/releases/download/v0.1.0/vittoriadb-v0.1.0-linux-amd64.tar.gz
tar -xzf vittoriadb-v0.1.0-linux-amd64.tar.gz
./vittoriadb run</code></pre>
    
    <h3>Option 2: From Source</h3>
    <p>Build from source with Go:</p>
    <pre><code>go install github.com/antonellof/VittoriaDB/cmd/vittoriadb@latest
vittoriadb run</code></pre>
    
    <h2>First Steps</h2>
    <ol>
        <li><strong>Start the server:</strong> Run <code>vittoriadb run</code></li>
        <li><strong>Create a collection:</strong> Use the REST API to create your first collection</li>
        <li><strong>Insert vectors:</strong> Add your vector data</li>
        <li><strong>Search:</strong> Perform similarity searches</li>
    </ol>
    
    <h2>API Examples</h2>
    <p>Here are some basic API calls to get you started:</p>
    
    <h3>Create Collection</h3>
    <pre><code>curl -X POST http://localhost:8080/collections \\
  -H "Content-Type: application/json" \\
  -d '{"name": "documents", "dimensions": 384}'</code></pre>
    
    <h3>Insert Vector</h3>
    <pre><code>curl -X POST http://localhost:8080/collections/documents/vectors \\
  -H "Content-Type: application/json" \\
  -d '{"id": "doc1", "vector": [0.1, 0.2, 0.3, ...], "metadata": {"title": "Test Document"}}'</code></pre>
    
    <h3>Search</h3>
    <pre><code>curl "http://localhost:8080/collections/documents/search?vector=0.1,0.2,0.3,...&limit=5"</code></pre>
    
    <h2>Next Steps</h2>
    <p>Once you have the basics working, explore:</p>
    <ul>
        <li>Document processing capabilities</li>
        <li>Python client integration</li>
        <li>Advanced indexing options</li>
        <li>Performance tuning</li>
    </ul>
</body>
</html>""")
    
    print(f"✅ Created sample documents in {docs_dir}")
    return docs_dir

def demonstrate_document_processing():
    """Demonstrate document processing capabilities."""
    print("🚀 VittoriaDB Document Processing Example")
    print("=" * 50)
    
    # Initialize client
    client = VittoriaDBDocumentProcessor()
    
    # Connect to VittoriaDB
    if not client.connect():
        print("❌ Failed to connect to VittoriaDB. Please start it with: ./vittoriadb run")
        return
    
    # Check connection
    if not client.health_check():
        print("❌ VittoriaDB is not healthy")
        return
    
    print("✅ Connected to VittoriaDB")
    
    # Get supported formats
    print("\n📋 Supported Document Formats:")
    formats = client.get_supported_formats()
    for fmt in formats.get("supported_formats", []):
        status_emoji = "✅" if fmt["status"] == "fully_implemented" else "🚧" if fmt["status"] == "placeholder" else "❌"
        print(f"  {status_emoji} {fmt['type'].upper()}: {fmt['description']}")
    
    # Create sample documents
    docs_dir = create_sample_documents()
    
    # Create collection
    collection_name = "processed_docs"
    print(f"\n🔄 Creating collection '{collection_name}'...")
    client.create_collection(collection_name, dimensions=384)
    
    # Process documents
    sample_files = [
        "sample.txt",
        "features.md", 
        "getting_started.html"
    ]
    
    processed_docs = []
    
    for filename in sample_files:
        file_path = docs_dir / filename
        if file_path.exists():
            print(f"\n📄 Processing {filename}...")
            
            # First, process without storing to see the result
            result = client.process_document(str(file_path), chunk_size=300, chunk_overlap=50)
            if result:
                print(f"  ✅ Extracted {len(result.get('chunks', []))} chunks")
                print(f"  📊 Document: {result.get('title', 'Unknown')} ({result.get('type', 'unknown')})")
                print(f"  📝 Content length: {len(result.get('content', ''))} characters")
                
                # Show metadata
                metadata = result.get('metadata', {})
                if metadata:
                    print(f"  🏷️  Metadata: {len(metadata)} fields")
                    for key, value in list(metadata.items())[:3]:  # Show first 3 metadata fields
                        print(f"     • {key}: {value}")
                
                # Upload to collection
                upload_result = client.upload_document(
                    collection_name=collection_name,
                    file_path=str(file_path),
                    chunk_size=300,
                    chunk_overlap=50,
                    metadata={"source": "example", "processed_by": "document_processing_example"}
                )
                
                if upload_result:
                    processed_docs.append({
                        "filename": filename,
                        "document_id": upload_result.get("document_id"),
                        "chunks": upload_result.get("chunks_created", 0)
                    })
                    print(f"  📤 Uploaded {upload_result.get('chunks_inserted', 0)} chunks to collection")
    
    # Show collection information
    print(f"\n📊 Collection Information:")
    info = client.get_collection_info(collection_name)
    if info:
        print(f"  • Collection name: {info.get('name', 'unknown')}")
        print(f"  • Dimensions: {info.get('dimensions', 0)}")
        print(f"  • Distance metric: {info.get('metric', 'unknown')}")
        print(f"  • Index type: {info.get('index_type', 'unknown')}")
    
    # Summary
    print(f"\n✅ Document Processing Complete!")
    print(f"📋 Summary:")
    print(f"  • Processed {len(processed_docs)} documents")
    total_chunks = sum(doc["chunks"] for doc in processed_docs)
    print(f"  • Created {total_chunks} total chunks")
    print(f"  • Stored in collection '{collection_name}'")
    
    print(f"\n💡 Next steps:")
    print(f"  • Use the collection for semantic search")
    print(f"  • Integrate with embedding models")
    print(f"  • Build RAG applications")
    print(f"  • Explore the web dashboard at http://localhost:8080")

if __name__ == "__main__":
    demonstrate_document_processing()
