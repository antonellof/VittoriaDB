"""
VittoriaDB Unified API - Schema-based document database client.

This module provides a modern, schema-driven API similar to advanced vector databases,
offering full-text search, vector search, hybrid search, and advanced query capabilities.
"""

import json
import requests
from typing import Dict, List, Any, Optional, Union
from .types import VittoriaDBError, ConnectionError


class VittoriaDocument:
    """
    VittoriaDB document-oriented client with schema-based API.
    
    This client provides a modern API for document storage and search:
    - Schema-based document structure
    - Full-text search with BM25 scoring  
    - Vector similarity search
    - Hybrid search combining text and vectors
    - Advanced filtering, facets, and sorting
    """

    def __init__(self, url: str = "http://localhost:8080"):
        """
        Initialize VittoriaDB document client.
        
        Args:
            url: VittoriaDB server URL
        """
        self.url = url.rstrip('/')
        self.session = requests.Session()
        self.schema = None

    def create(self, schema: Dict[str, Any], **kwargs) -> Dict[str, Any]:
        """
        Create a document database with schema definition.
        
        Args:
            schema: Document schema definition
            **kwargs: Additional configuration options
                - language: Text processing language (default: "english")
                - fulltext_config: Full-text search configuration
                - vectorizer_configs: Vectorizer configurations for vector fields
                - content_storage: Content storage configuration
        
        Returns:
            Creation response with schema information
            
        Example:
            >>> db = VittoriaDocument()
            >>> schema = {
            ...     "name": "string",
            ...     "description": "string", 
            ...     "price": "number",
            ...     "embedding": "vector[1536]",
            ...     "meta": {
            ...         "rating": "number"
            ...     }
            ... }
            >>> response = db.create(schema)
        """
        payload = {
            "schema": schema,
            **kwargs
        }
        
        try:
            response = self.session.post(f"{self.url}/create", json=payload)
            response.raise_for_status()
            result = response.json()
            self.schema = schema  # Store schema for validation
            return result
        except requests.RequestException as e:
            raise ConnectionError(f"Failed to create document database: {e}")

    def insert(self, document: Dict[str, Any], **options) -> str:
        """
        Insert a document into the database.
        
        Args:
            document: Document to insert (must conform to schema)
            **options: Insert options
                - skip_validation: Skip schema validation
                - upsert: Update if document exists
        
        Returns:
            Document ID
            
        Example:
            >>> doc_id = db.insert({
            ...     "name": "Noise cancelling headphones",
            ...     "description": "Best headphones on market",
            ...     "price": 99.99,
            ...     "embedding": [0.1, 0.2, 0.3, ...],
            ...     "meta": {"rating": 4.5}
            ... })
        """
        payload = {
            "document": document,
            "options": options
        }
        
        try:
            response = self.session.post(f"{self.url}/documents", json=payload)
            response.raise_for_status()
            result = response.json()
            return result["id"]
        except requests.RequestException as e:
            raise VittoriaDBError(f"Failed to insert document: {e}")

    def search(self, 
               term: Optional[str] = None,
               mode: str = "fulltext",
               vector: Optional[Dict[str, Any]] = None,
               limit: int = 10,
               offset: int = 0,
               **kwargs) -> Dict[str, Any]:
        """
        Search documents with multiple modes and advanced options.
        
        Args:
            term: Search term for full-text search
            mode: Search mode ("fulltext", "vector", "hybrid")
            vector: Vector search parameters {"value": [...], "property": "field"}
            limit: Maximum results to return
            offset: Number of results to skip
            **kwargs: Additional search options
                - properties: Fields to search in (for full-text)
                - where: Filter conditions
                - facets: Facet configuration
                - sort_by: Sorting configuration
                - group_by: Grouping configuration  
                - threshold: Relevance threshold
                - similarity: Vector similarity threshold
                - hybrid_weights: Weights for hybrid search
                - include_vectors: Include vector data in results
                - boost: Field boosting for full-text search
                - relevance: BM25 parameters
                - exact: Exact matching
                - tolerance: Typo tolerance
        
        Returns:
            Search results with hits, count, elapsed time, facets
            
        Examples:
            # Full-text search
            >>> results = db.search(term="best headphones", limit=5)
            
            # Vector search  
            >>> results = db.search(
            ...     mode="vector",
            ...     vector={"value": [0.1, 0.2, ...], "property": "embedding"},
            ...     similarity=0.8
            ... )
            
            # Hybrid search
            >>> results = db.search(
            ...     term="noise cancelling",
            ...     mode="hybrid", 
            ...     vector={"value": [0.1, 0.2, ...], "property": "embedding"},
            ...     hybrid_weights={"text": 0.7, "vector": 0.3}
            ... )
            
            # Advanced search with facets and filters
            >>> results = db.search(
            ...     term="headphones",
            ...     where={"price": {"lt": 200}},
            ...     facets={"category": {"type": "string", "limit": 10}},
            ...     boost={"name": 2.0, "description": 1.0}
            ... )
        """
        params = {
            "mode": mode,
            "limit": limit,
            "offset": offset,
            **kwargs
        }
        
        if term:
            params["term"] = term
        if vector:
            params["vector"] = vector
            
        try:
            response = self.session.post(f"{self.url}/search", json=params)
            response.raise_for_status()
            return response.json()
        except requests.RequestException as e:
            raise VittoriaDBError(f"Search failed: {e}")

    def get(self, doc_id: str, include_vectors: bool = False) -> Optional[Dict[str, Any]]:
        """
        Get a document by ID.
        
        Args:
            doc_id: Document ID
            include_vectors: Include vector data in response
            
        Returns:
            Document if found, None otherwise
        """
        params = {}
        if include_vectors:
            params["include_vectors"] = "true"
            
        try:
            response = self.session.get(f"{self.url}/documents/{doc_id}", params=params)
            response.raise_for_status()
            result = response.json()
            return result["document"] if result["found"] else None
        except requests.RequestException as e:
            raise VittoriaDBError(f"Failed to get document: {e}")

    def update(self, doc_id: str, document: Dict[str, Any], **options) -> bool:
        """
        Update a document.
        
        Args:
            doc_id: Document ID
            document: Updated document data
            **options: Update options
                - skip_validation: Skip schema validation
                - partial: Partial update (merge with existing)
                
        Returns:
            True if updated successfully
        """
        payload = {
            "document": document,
            "options": options
        }
        
        try:
            response = self.session.put(f"{self.url}/documents/{doc_id}", json=payload)
            response.raise_for_status()
            result = response.json()
            return result["updated"]
        except requests.RequestException as e:
            raise VittoriaDBError(f"Failed to update document: {e}")

    def delete(self, doc_id: str) -> bool:
        """
        Delete a document.
        
        Args:
            doc_id: Document ID
            
        Returns:
            True if deleted successfully
        """
        try:
            response = self.session.delete(f"{self.url}/documents/{doc_id}")
            response.raise_for_status()
            result = response.json()
            return result["deleted"]
        except requests.RequestException as e:
            raise VittoriaDBError(f"Failed to delete document: {e}")

    def count(self, where: Optional[Dict[str, Any]] = None) -> int:
        """
        Count documents matching optional filter.
        
        Args:
            where: Filter conditions
            
        Returns:
            Number of matching documents
        """
        params = {}
        if where:
            params["where"] = json.dumps(where)
            
        try:
            response = self.session.get(f"{self.url}/count", params=params)
            response.raise_for_status()
            result = response.json()
            return result["count"]
        except requests.RequestException as e:
            raise VittoriaDBError(f"Failed to count documents: {e}")

    # Convenience methods for different search modes

    def search_text(self, 
                   term: str, 
                   properties: Optional[List[str]] = None,
                   limit: int = 10,
                   **kwargs) -> Dict[str, Any]:
        """
        Perform full-text search.
        
        Args:
            term: Search term
            properties: Fields to search in
            limit: Maximum results
            **kwargs: Additional search options
            
        Returns:
            Search results
        """
        return self.search(
            term=term,
            mode="fulltext", 
            properties=properties,
            limit=limit,
            **kwargs
        )

    def search_vector(self,
                     vector_value: List[float],
                     vector_property: str,
                     similarity: float = 0.8,
                     limit: int = 10,
                     **kwargs) -> Dict[str, Any]:
        """
        Perform vector similarity search.
        
        Args:
            vector_value: Query vector
            vector_property: Vector field name
            similarity: Minimum similarity threshold
            limit: Maximum results
            **kwargs: Additional search options
            
        Returns:
            Search results
        """
        return self.search(
            mode="vector",
            vector={"value": vector_value, "property": vector_property},
            similarity=similarity,
            limit=limit,
            **kwargs
        )

    def search_hybrid(self,
                     term: str,
                     vector_value: List[float], 
                     vector_property: str,
                     text_weight: float = 0.5,
                     vector_weight: float = 0.5,
                     limit: int = 10,
                     **kwargs) -> Dict[str, Any]:
        """
        Perform hybrid search combining text and vector search.
        
        Args:
            term: Search term
            vector_value: Query vector
            vector_property: Vector field name
            text_weight: Weight for text search results
            vector_weight: Weight for vector search results
            limit: Maximum results
            **kwargs: Additional search options
            
        Returns:
            Search results
        """
        return self.search(
            term=term,
            mode="hybrid",
            vector={"value": vector_value, "property": vector_property},
            hybrid_weights={"text": text_weight, "vector": vector_weight},
            limit=limit,
            **kwargs
        )


def create(schema: Dict[str, Any], url: str = "http://localhost:8080", **kwargs) -> VittoriaDocument:
    """
    Create a VittoriaDB document database with schema.
    
    This is a convenience function that creates and configures a document database in one call.
    
    Args:
        schema: Document schema definition
        url: VittoriaDB server URL
        **kwargs: Additional configuration options
        
    Returns:
        Configured VittoriaDocument instance
        
    Example:
        >>> db = create({
        ...     "name": "string",
        ...     "description": "string",
        ...     "price": "number", 
        ...     "embedding": "vector[1536]",
        ...     "meta": {"rating": "number"}
        ... })
        >>> 
        >>> # Insert documents
        >>> db.insert({
        ...     "name": "Noise cancelling headphones",
        ...     "description": "Best headphones on market", 
        ...     "price": 99.99,
        ...     "embedding": [0.1, 0.2, 0.3, ...],
        ...     "meta": {"rating": 4.5}
        ... })
        >>>
        >>> # Search with different modes
        >>> results = db.search(term="best headphones")
    """
    client = VittoriaDocument(url)
    client.create(schema, **kwargs)
    return client
