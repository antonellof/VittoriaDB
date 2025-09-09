"""
VittoriaDB - Simple Embedded Vector Database

A simple, embedded, zero-configuration vector database
for local AI development.

Example usage:
    import vittoriadb
    
    # Auto-starts binary, manages lifecycle
    db = vittoriadb.connect()
    
    # Create collection
    collection = db.create_collection("documents", dimensions=384)
    
    # Insert vector
    collection.insert("doc1", [0.1, 0.2, 0.3] * 128, {"title": "Test"})
    
    # Search
    results = collection.search([0.1, 0.2, 0.3] * 128, limit=5)
    
    # Close
    db.close()
"""

from .client import VittoriaDB, Collection, connect
from .types import (
    Vector,
    SearchResult,
    CollectionInfo,
    DistanceMetric,
    IndexType,
    VittoriaDBError,
    ConnectionError,
    CollectionError,
    VectorError,
    SearchError,
    BinaryError
)

__version__ = "0.1.0"
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
    "VittoriaDBError",
    "ConnectionError",
    "CollectionError",
    "VectorError",
    "SearchError",
    "BinaryError",
]
