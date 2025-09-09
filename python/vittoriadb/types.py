"""
Data types and structures for VittoriaDB Python client.
"""

from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from enum import Enum


class DistanceMetric(Enum):
    """Distance metrics for vector similarity calculation."""
    COSINE = "cosine"
    EUCLIDEAN = "euclidean"
    DOT_PRODUCT = "dot_product"
    MANHATTAN = "manhattan"


class IndexType(Enum):
    """Vector index types."""
    FLAT = "flat"
    HNSW = "hnsw"
    IVF = "ivf"


@dataclass
class Vector:
    """Represents a vector with metadata."""
    id: str
    vector: List[float]
    metadata: Optional[Dict[str, Any]] = None

    def __post_init__(self):
        if self.metadata is None:
            self.metadata = {}


@dataclass
class SearchResult:
    """Represents a search result."""
    id: str
    score: float
    vector: Optional[List[float]] = None
    metadata: Optional[Dict[str, Any]] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'SearchResult':
        """Create SearchResult from dictionary."""
        return cls(
            id=data["id"],
            score=data["score"],
            vector=data.get("vector"),
            metadata=data.get("metadata")
        )


@dataclass
class CollectionInfo:
    """Represents collection information."""
    name: str
    dimensions: int
    metric: DistanceMetric
    index_type: IndexType
    vector_count: int
    created: str
    modified: str

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'CollectionInfo':
        """Create CollectionInfo from dictionary."""
        return cls(
            name=data["name"],
            dimensions=data["dimensions"],
            metric=DistanceMetric(data["metric"]),
            index_type=IndexType(data["index_type"]),
            vector_count=data["vector_count"],
            created=data["created"],
            modified=data["modified"]
        )


@dataclass
class HealthStatus:
    """Represents database health status."""
    status: str
    uptime: int
    collections: int
    total_vectors: int
    memory_usage: int
    disk_usage: int

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'HealthStatus':
        """Create HealthStatus from dictionary."""
        return cls(
            status=data["status"],
            uptime=data["uptime"],
            collections=data["collections"],
            total_vectors=data["total_vectors"],
            memory_usage=data["memory_usage"],
            disk_usage=data["disk_usage"]
        )


@dataclass
class DatabaseStats:
    """Represents database statistics."""
    total_vectors: int
    total_size: int
    index_size: int
    queries_total: int
    queries_per_sec: float
    avg_query_latency: float
    collections: List[Dict[str, Any]]

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'DatabaseStats':
        """Create DatabaseStats from dictionary."""
        return cls(
            total_vectors=data["total_vectors"],
            total_size=data["total_size"],
            index_size=data["index_size"],
            queries_total=data["queries_total"],
            queries_per_sec=data["queries_per_sec"],
            avg_query_latency=data["avg_query_latency"],
            collections=data["collections"]
        )


class VittoriaDBError(Exception):
    """Base exception for VittoriaDB errors."""
    pass


class ConnectionError(VittoriaDBError):
    """Raised when connection to VittoriaDB fails."""
    pass


class CollectionError(VittoriaDBError):
    """Raised when collection operations fail."""
    pass


class VectorError(VittoriaDBError):
    """Raised when vector operations fail."""
    pass


class SearchError(VittoriaDBError):
    """Raised when search operations fail."""
    pass


class BinaryError(VittoriaDBError):
    """Raised when binary management fails."""
    pass
