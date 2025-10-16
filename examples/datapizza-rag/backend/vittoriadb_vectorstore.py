"""
VittoriaDB Vectorstore Adapter for Datapizza AI
Implements the Datapizza vectorstore interface for VittoriaDB
"""

import logging
from typing import List, Dict, Any, Optional, Union
import vittoriadb
from vittoriadb.types import IndexType, DistanceMetric
from vittoriadb.configure import Configure

logger = logging.getLogger(__name__)


class VittoriaDBVectorstore:
    """
    VittoriaDB adapter for Datapizza AI pipelines.
    Implements the vectorstore interface compatible with Datapizza's IngestionPipeline and DagPipeline.
    """
    
    def __init__(
        self,
        url: str = "http://localhost:8080",
        auto_start: bool = False
    ):
        """
        Initialize VittoriaDB vectorstore.
        
        Args:
            url: VittoriaDB server URL
            auto_start: Whether to auto-start VittoriaDB (set False in Docker)
        """
        self.url = url
        self.auto_start = auto_start
        self.db = None
        self._connect()
        
    def _connect(self):
        """Connect to VittoriaDB server"""
        try:
            self.db = vittoriadb.connect(
                url=self.url,
                auto_start=self.auto_start
            )
            logger.info(f"✅ Connected to VittoriaDB at {self.url}")
        except Exception as e:
            logger.error(f"❌ Failed to connect to VittoriaDB: {e}")
            raise
    
    def create_collection(
        self,
        collection_name: str,
        vector_config: List[Dict[str, Any]],
        metric: str = "cosine",
        index_type: str = "hnsw",
        **kwargs
    ):
        """
        Create a collection in VittoriaDB.
        
        Args:
            collection_name: Name of the collection
            vector_config: List of vector configurations (Datapizza format)
                          Example: [VectorConfig(name="embedding", dimensions=1536)]
            metric: Distance metric ("cosine", "euclidean", "dot")
            index_type: Index type ("hnsw", "flat")
            **kwargs: Additional collection configuration
        """
        try:
            # Extract dimensions from vector_config
            # Datapizza format: [VectorConfig(name="embedding", dimensions=1536)]
            if vector_config and len(vector_config) > 0:
                config = vector_config[0]
                dimensions = config.get('dimensions') if isinstance(config, dict) else config.dimensions
            else:
                raise ValueError("vector_config must contain at least one configuration")
            
            # Map metric to VittoriaDB DistanceMetric
            metric_map = {
                "cosine": DistanceMetric.COSINE,
                "euclidean": DistanceMetric.EUCLIDEAN,
                "dot": DistanceMetric.DOT_PRODUCT
            }
            distance_metric = metric_map.get(metric.lower(), DistanceMetric.COSINE)
            
            # Map index_type to VittoriaDB IndexType
            index_map = {
                "hnsw": IndexType.HNSW,
                "flat": IndexType.FLAT
            }
            index = index_map.get(index_type.lower(), IndexType.HNSW)
            
            # HNSW configuration
            hnsw_config = {
                "m": 16,
                "ef_construction": 200,
                "ef_search": 50
            }
            
            # Check if collection exists
            try:
                existing = self.db.get_collection(collection_name)
                logger.info(f"✅ Collection '{collection_name}' already exists")
                return existing
            except:
                pass
            
            # Create collection
            collection = self.db.create_collection(
                name=collection_name,
                dimensions=dimensions,
                metric=distance_metric,
                index_type=index,
                config=hnsw_config if index == IndexType.HNSW else {}
            )
            
            logger.info(f"✅ Created collection '{collection_name}' ({dimensions}D, {metric}, {index_type})")
            return collection
            
        except Exception as e:
            logger.error(f"❌ Failed to create collection '{collection_name}': {e}")
            raise
    
    def upsert(
        self,
        collection_name: str,
        chunks: List[Dict[str, Any]],
        **kwargs
    ):
        """
        Insert or update chunks in the collection.
        Compatible with Datapizza IngestionPipeline.
        
        Args:
            collection_name: Name of the collection
            chunks: List of chunk dictionaries with 'id', 'embedding', 'text', and 'metadata'
        """
        try:
            collection = self.db.get_collection(collection_name)
            
            for chunk in chunks:
                # Extract chunk data (Datapizza format)
                chunk_id = chunk.get('id', chunk.get('chunk_id', f"chunk_{hash(chunk.get('text', ''))}"))
                embedding = chunk.get('embedding', chunk.get('vector'))
                text = chunk.get('text', chunk.get('content', ''))
                metadata = chunk.get('metadata', {})
                
                # Add text to metadata for storage
                metadata['text'] = text
                metadata['chunk_id'] = chunk_id
                
                # Insert into VittoriaDB
                if embedding:
                    collection.insert(
                        id=chunk_id,
                        vector=embedding,
                        metadata=metadata
                    )
                else:
                    logger.warning(f"⚠️ Chunk {chunk_id} has no embedding, skipping")
            
            logger.info(f"✅ Upserted {len(chunks)} chunks to '{collection_name}'")
            
        except Exception as e:
            logger.error(f"❌ Failed to upsert chunks to '{collection_name}': {e}")
            raise
    
    def search(
        self,
        collection_name: str,
        query_vector: Union[List[float], List[List[float]]],
        k: int = 5,
        filters: Optional[Dict[str, Any]] = None,
        **kwargs
    ) -> List[Dict[str, Any]]:
        """
        Search for similar vectors in the collection.
        Compatible with Datapizza DagPipeline.
        
        Args:
            collection_name: Name of the collection
            query_vector: Query vector(s) - single vector or list of vectors
            k: Number of results to return
            filters: Optional metadata filters
            **kwargs: Additional search parameters
            
        Returns:
            List of search results in Datapizza format
        """
        try:
            collection = self.db.get_collection(collection_name)
            
            # Handle single vector or list of vectors
            if isinstance(query_vector[0], list):
                # Multiple query vectors - use first one
                query_vector = query_vector[0]
            
            # Search VittoriaDB
            results = collection.search(
                vector=query_vector,
                limit=k
            )
            
            # Convert to Datapizza format
            datapizza_results = []
            for result in results:
                datapizza_results.append({
                    'id': result.id,
                    'score': result.score,
                    'text': result.metadata.get('text', result.metadata.get('content', '')),
                    'metadata': result.metadata,
                    'embedding': result.vector if hasattr(result, 'vector') else None
                })
            
            logger.info(f"✅ Found {len(datapizza_results)} results in '{collection_name}'")
            return datapizza_results
            
        except Exception as e:
            logger.error(f"❌ Failed to search '{collection_name}': {e}")
            raise
    
    def delete_collection(self, collection_name: str):
        """Delete a collection"""
        try:
            self.db.delete_collection(collection_name)
            logger.info(f"✅ Deleted collection '{collection_name}'")
        except Exception as e:
            logger.error(f"❌ Failed to delete collection '{collection_name}': {e}")
            raise
    
    def list_collections(self) -> List[str]:
        """List all collections"""
        try:
            # VittoriaDB doesn't have a direct list_collections method
            # We'll need to maintain a list or query the API
            logger.warning("list_collections not fully implemented for VittoriaDB")
            return []
        except Exception as e:
            logger.error(f"❌ Failed to list collections: {e}")
            return []

