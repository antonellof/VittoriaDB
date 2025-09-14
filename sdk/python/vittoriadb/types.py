"""
Data types and structures for VittoriaDB Python client.
"""

from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from enum import Enum


class DistanceMetric(Enum):
    """Distance metrics for vector similarity calculation."""
    COSINE = 0
    EUCLIDEAN = 1
    DOT_PRODUCT = 2
    MANHATTAN = 3
    
    @classmethod
    def from_string(cls, value: str) -> 'DistanceMetric':
        """Create DistanceMetric from string value."""
        string_map = {
            "cosine": cls.COSINE,
            "euclidean": cls.EUCLIDEAN,
            "dot_product": cls.DOT_PRODUCT,
            "manhattan": cls.MANHATTAN
        }
        return string_map.get(value.lower(), cls.COSINE)
    
    def to_string(self) -> str:
        """Convert to string representation."""
        string_map = {
            self.COSINE: "cosine",
            self.EUCLIDEAN: "euclidean", 
            self.DOT_PRODUCT: "dot_product",
            self.MANHATTAN: "manhattan"
        }
        return string_map.get(self, "cosine")


class IndexType(Enum):
    """Vector index types."""
    FLAT = 0
    HNSW = 1
    IVF = 2
    
    @classmethod
    def from_string(cls, value: str) -> 'IndexType':
        """Create IndexType from string value."""
        string_map = {
            "flat": cls.FLAT,
            "hnsw": cls.HNSW,
            "ivf": cls.IVF
        }
        return string_map.get(value.lower(), cls.FLAT)
    
    def to_string(self) -> str:
        """Convert to string representation."""
        string_map = {
            self.FLAT: "flat",
            self.HNSW: "hnsw",
            self.IVF: "ivf"
        }
        return string_map.get(self, "flat")


class VectorizerType(Enum):
    """Vectorizer types for automatic embedding generation."""
    NONE = 0
    SENTENCE_TRANSFORMERS = 1
    OPENAI = 2
    HUGGINGFACE = 3
    OLLAMA = 4
    
    @classmethod
    def from_string(cls, value: str) -> 'VectorizerType':
        """Create VectorizerType from string value."""
        string_map = {
            "none": cls.NONE,
            "sentence_transformers": cls.SENTENCE_TRANSFORMERS,
            "openai": cls.OPENAI,
            "huggingface": cls.HUGGINGFACE,
            "ollama": cls.OLLAMA
        }
        return string_map.get(value.lower(), cls.NONE)
    
    def to_string(self) -> str:
        """Convert to string representation."""
        string_map = {
            self.NONE: "none",
            self.SENTENCE_TRANSFORMERS: "sentence_transformers",
            self.OPENAI: "openai",
            self.HUGGINGFACE: "huggingface",
            self.OLLAMA: "ollama"
        }
        return string_map.get(self, "none")


@dataclass
class VectorizerConfig:
    """Configuration for automatic vectorization."""
    type: VectorizerType
    model: str
    dimensions: int
    options: Dict[str, Any]
    
    def __post_init__(self):
        if self.options is None:
            self.options = {}
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for API calls."""
        return {
            "type": self.type.value,  # Send integer value, not string
            "model": self.model,
            "dimensions": self.dimensions,
            "options": self.options
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'VectorizerConfig':
        """Create VectorizerConfig from dictionary."""
        return cls(
            type=VectorizerType.from_string(data["type"]),
            model=data["model"],
            dimensions=data["dimensions"],
            options=data.get("options", {})
        )


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
