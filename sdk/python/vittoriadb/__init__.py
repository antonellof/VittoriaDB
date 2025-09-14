"""
VittoriaDB Python SDK - Vector Database Client

A Python client library for VittoriaDB, a high-performance embedded vector database.
This SDK connects to and manages VittoriaDB Go server instances with automatic 
binary management and lifecycle control.

The VittoriaDB server is a single Go binary that provides:
- Zero-configuration vector database
- High-performance HNSW indexing  
- Server-side automatic embeddings
- Document processing capabilities
- RESTful API with native Python client

Example usage:
    import vittoriadb
    from vittoriadb.configure import Configure
    
    # Auto-starts VittoriaDB server binary and connects
    db = vittoriadb.connect()
    
    # Create collection with automatic embeddings
    collection = db.create_collection(
        name="documents", 
        dimensions=384,
        vectorizer_config=Configure.Vectors.auto_embeddings()
    )
    
    # Insert text (server generates embeddings automatically)
    collection.insert_text("doc1", "Your text content", {"title": "Test"})
    
    # Search with text (server vectorizes query automatically)  
    results = collection.search_text("search query", limit=5)
    
    # Close connection
    db.close()

Server Management:
    # Manual server control
    ./vittoriadb run                    # Start server manually
    db = vittoriadb.connect(auto_start=False)  # Connect to existing server
    
    # Or let the SDK manage the server
    db = vittoriadb.connect(auto_start=True)   # Auto-start and manage server
"""

from .client import VittoriaDB, Collection, connect
from .types import (
    Vector,
    SearchResult,
    CollectionInfo,
    DistanceMetric,
    IndexType,
    VectorizerType,
    VectorizerConfig,
    VittoriaDBError,
    ConnectionError,
    CollectionError,
    VectorError,
    SearchError,
    BinaryError
)
from . import configure

__version__ = "0.3.0"
__author__ = "VittoriaDB Team"
__email__ = "team@vittoriadb.dev"

__all__ = [
    "VittoriaDB",
    "Collection", 
    "connect",
    "Vector",
    "SearchResult",
    "CollectionInfo",
    "DistanceMetric",
    "IndexType",
    "VectorizerType",
    "VectorizerConfig",
    "VittoriaDBError",
    "ConnectionError",
    "CollectionError",
    "VectorError",
    "SearchError",
    "BinaryError",
    "configure",
]
