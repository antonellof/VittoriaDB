"""
RAG System V2 - Datapizza AI Pipeline Wrapper
Provides backward compatibility with the old RAG system interface while using Datapizza pipelines internally
"""

import os
import logging
from typing import List, Dict, Any, Optional
from dataclasses import dataclass

from datapizza_rag_pipeline import DatapizzaRAGPipeline, DatapizzaRAGConfig
from vittoriadb_vectorstore import VittoriaDBVectorstore
from models import SearchResult

logger = logging.getLogger(__name__)


class RAGSystemV2:
    """
    RAG System V2 using Datapizza AI pipelines.
    Compatible interface with the old RAGSystem for easy migration.
    """
    
    def __init__(
        self,
        vittoriadb_url: str = "http://localhost:8080",
        openai_api_key: Optional[str] = None,
        use_ollama: bool = False
    ):
        """Initialize RAG system with Datapizza pipeline"""
        self.vittoriadb_url = vittoriadb_url
        self.openai_api_key = openai_api_key or os.getenv('OPENAI_API_KEY')
        self.use_ollama = use_ollama
        
        # Create Datapizza RAG pipeline
        config = DatapizzaRAGConfig(
            openai_api_key=self.openai_api_key,
            vittoriadb_url=vittoriadb_url,
            embedding_model=os.getenv('OPENAI_EMBED_MODEL', 'text-embedding-ada-002'),
            embedding_dimensions=int(os.getenv('OPENAI_EMBED_DIMENSIONS', '1536')),
            llm_model=os.getenv('LLM_MODEL', 'gpt-4o-mini'),
            chunk_size=int(os.getenv('CHUNK_SIZE', '1000')),
            chunk_overlap=int(os.getenv('CHUNK_OVERLAP', '200')),
            retrieval_k=int(os.getenv('RETRIEVAL_K', '5'))
        )
        
        self.pipeline = DatapizzaRAGPipeline(config)
        self.vectorstore = self.pipeline.vectorstore
        
        # Keep track of collections
        self.collections = {}
        self.default_collections = ['documents', 'web_research', 'github_code', 'chat_history']
        
        # Collection configurations for stats display
        self.collection_configs = {
            'documents': {'description': 'User uploaded documents'},
            'web_research': {'description': 'Web research results'},
            'github_code': {'description': 'GitHub repository code'},
            'chat_history': {'description': 'Chat conversation history'}
        }
        
        # For compatibility with old interface
        self.openai_client = self.pipeline.llm_client
        self.embedder = self.pipeline.embedder
        
        # Initialize default collections
        self._initialize_collections()
        
        logger.info("✅ RAG System V2 initialized with Datapizza AI pipelines")
    
    def _initialize_collections(self):
        """Initialize default collections"""
        for collection_name in self.default_collections:
            try:
                self.pipeline.create_collection(collection_name, replace_existing=False)
                self.collections[collection_name] = collection_name
                logger.info(f"✅ Collection '{collection_name}' initialized")
            except Exception as e:
                logger.warning(f"⚠️ Could not initialize collection '{collection_name}': {e}")
    
    async def search_knowledge_base(
        self,
        query: str,
        collections: List[str],
        limit: int = 5,
        min_score: float = 0.3
    ) -> List[SearchResult]:
        """
        Search knowledge base across collections.
        Compatible with old interface.
        
        Args:
            query: Search query
            collections: List of collection names
            limit: Maximum results
            min_score: Minimum relevance score
            
        Returns:
            List of SearchResult objects
        """
        try:
            # Embed the query
            query_embedding = self.embedder.embed(text=query)
            
            all_results = []
            for collection_name in collections:
                try:
                    # Search using VittoriaDB vectorstore
                    results = self.vectorstore.search(
                        collection_name=collection_name,
                        query_vector=query_embedding,
                        k=limit
                    )
                    
                    # Convert to SearchResult objects
                    for result in results:
                        if result['score'] >= min_score:
                            search_result = SearchResult(
                                id=result['id'],
                                score=result['score'],
                                content=result['text'],
                                metadata=result['metadata'],
                                source=collection_name
                            )
                            all_results.append(search_result)
                            
                except Exception as e:
                    logger.warning(f"⚠️ Search failed for collection '{collection_name}': {e}")
            
            # Sort by score and limit
            all_results.sort(key=lambda x: x.score, reverse=True)
            return all_results[:limit]
            
        except Exception as e:
            logger.error(f"❌ Knowledge base search failed: {e}")
            return []
    
    async def add_document(
        self,
        collection_name: str,
        doc_id: str,
        content: str,
        metadata: Dict[str, Any]
    ):
        """
        Add a document to a collection.
        Compatible with old interface.
        """
        try:
            # Ensure collection exists
            if collection_name not in self.collections:
                self.pipeline.create_collection(collection_name, replace_existing=False)
                self.collections[collection_name] = collection_name
            
            # Ingest the document
            metadata['doc_id'] = doc_id
            self.pipeline.ingest_text(
                text=content,
                collection_name=collection_name,
                metadata=metadata
            )
            
            logger.info(f"✅ Added document '{doc_id}' to '{collection_name}'")
            
        except Exception as e:
            logger.error(f"❌ Failed to add document: {e}")
            raise
    
    async def add_prechunked_document(
        self,
        collection_name: str,
        doc_id: str,
        content: str,
        metadata: Dict[str, Any]
    ):
        """
        Add a pre-chunked document directly to VittoriaDB without re-chunking.
        This is used when documents are already processed by file_processor.
        """
        try:
            # Ensure collection exists
            if collection_name not in self.collections:
                self.pipeline.create_collection(collection_name, replace_existing=False)
                self.collections[collection_name] = collection_name
            
            # Generate embedding for the content
            embedding = self.embedder.embed(text=content)
            
            # Store content in metadata for retrieval
            metadata_with_content = {**metadata, 'content': content}
            
            # Insert directly into VittoriaDB using correct API signature
            collection = self.vectorstore.db.get_collection(collection_name)
            collection.insert(
                id=doc_id,
                vector=embedding,
                metadata=metadata_with_content
            )
            
            logger.info(f"✅ Added pre-chunked document '{doc_id}' to '{collection_name}'")
            
        except Exception as e:
            logger.error(f"❌ Failed to add pre-chunked document: {e}")
            raise
    
    async def add_documents_batch(
        self,
        collection_name: str,
        documents: List[Dict[str, Any]]
    ):
        """
        Add multiple pre-chunked documents in batch.
        Compatible with old interface (expects already-chunked content from file_processor).
        """
        for doc in documents:
            await self.add_prechunked_document(
                collection_name=collection_name,
                doc_id=doc.get('id', doc.get('doc_id', f"doc_{hash(doc['content'])}")),
                content=doc['content'],
                metadata=doc.get('metadata', {})
            )
    
    def get_collection_stats(self) -> Dict[str, Any]:
        """Get statistics for all collections"""
        import requests
        stats = {}
        
        for collection_name in self.collections:
            try:
                # Get detailed collection info directly from VittoriaDB API
                response = requests.get(f"{self.vittoriadb_url}/collections/{collection_name}")
                if response.status_code == 200:
                    api_info = response.json()
                    
                    # Map index_type number to string
                    index_type_map = {0: "flat", 1: "hnsw"}
                    index_type = index_type_map.get(api_info.get('index_type', 0), "unknown")
                    
                    # Map metric number to string  
                    metric_map = {0: "cosine", 1: "euclidean", 2: "dot"}
                    metric = metric_map.get(api_info.get('metric', 0), "unknown")
                    
                    stats[collection_name] = {
                        'name': api_info['name'],
                        'vector_count': api_info['vector_count'],
                        'dimensions': api_info['dimensions'],
                        'metric': metric,
                        'index_type': index_type,
                        'description': self.collection_configs.get(collection_name, {}).get('description', ''),
                        'status': 'active'
                    }
                else:
                    # Fallback to placeholder if API call fails
                    logger.warning(f"⚠️ Could not get stats for '{collection_name}': HTTP {response.status_code}")
                    stats[collection_name] = {
                        'name': collection_name,
                        'vector_count': 0,
                        'dimensions': self.pipeline.config.embedding_dimensions,
                        'metric': 'cosine',
                        'index_type': 'unknown',
                        'description': self.collection_configs.get(collection_name, {}).get('description', ''),
                        'status': 'error'
                    }
                    
            except Exception as e:
                logger.error(f"❌ Failed to get stats for '{collection_name}': {e}")
                stats[collection_name] = {
                    'name': collection_name,
                    'vector_count': 0,
                    'dimensions': self.pipeline.config.embedding_dimensions,
                    'metric': 'cosine',
                    'index_type': 'unknown',
                    'description': self.collection_configs.get(collection_name, {}).get('description', ''),
                    'status': 'error',
                    'error': str(e)
                }
        
        return stats
    
    async def generate_response(
        self,
        user_query: str,
        system_prompt: str,
        model: str = None
    ) -> str:
        """
        Generate a response using the LLM.
        Compatible with old interface.
        """
        try:
            model = model or self.pipeline.config.llm_model
            
            response = await self.openai_client.a_invoke(
                input=user_query,
                system_prompt=system_prompt,
                model=model,
                temperature=0.7,
                max_tokens=1500
            )
            
            return response.text
            
        except Exception as e:
            logger.error(f"❌ Response generation failed: {e}")
            return f"❌ Error: {str(e)}"
    
    async def list_documents(self, collection_name: str = 'documents', limit: int = 200) -> List[Dict[str, Any]]:
        """List all documents in a collection WITHOUT using embeddings (direct VittoriaDB access)"""
        try:
            if collection_name not in self.collections:
                raise ValueError(f"Collection '{collection_name}' not found")
            
            logger.info(f"🔍 Listing documents from {collection_name} using direct VittoriaDB access (NO embeddings)")
            
            # Use VittoriaDB Python SDK directly to avoid embedding generation
            collection = self.vectorstore.db.get_collection(collection_name)
            
            # Try to get all vectors using the VittoriaDB SDK
            try:
                import numpy as np
                
                # Create a zero vector with the right dimensions
                dimensions = self.pipeline.config.embedding_dimensions
                zero_vector = np.zeros(dimensions).tolist()
                
                # Search with zero vector to get everything
                results = collection.search(
                    vector=zero_vector,
                    limit=limit
                )
                
                documents = []
                for result in results:
                    # Extract data from VittoriaDB result
                    result_dict = result.__dict__ if hasattr(result, '__dict__') else {}
                    
                    documents.append({
                        'id': result_dict.get('id', 'unknown'),
                        'metadata': result_dict.get('metadata', {}),
                        'score': result_dict.get('score', 0.0),
                        'content': result_dict.get('content', '')
                    })
                
                logger.info(f"📋 Listed {len(documents)} documents using VittoriaDB SDK (no embeddings)")
                return documents
                
            except Exception as sdk_error:
                logger.error(f"❌ Failed to list documents using SDK: {sdk_error}")
                return []
                
        except Exception as e:
            logger.error(f"❌ Failed to list documents: {str(e)}")
            return []
    
    async def get_original_documents(self, collection_name: str = 'documents') -> List[Dict[str, Any]]:
        """Get original documents (grouped by source file) instead of individual chunks"""
        try:
            if collection_name not in self.collections:
                raise ValueError(f"Collection '{collection_name}' not found")
            
            # Get all chunks from the collection
            documents = await self.list_documents(collection_name, 1000)
            logger.info(f"🔍 Retrieved {len(documents)} raw documents from list_documents")
            
            # Group chunks by original document
            document_groups = {}
            
            for doc in documents:
                metadata = doc.get('metadata', {})
                
                # Use document_id as the primary grouping key, but try multiple fields
                doc_id = (metadata.get('document_id') or 
                         metadata.get('content_hash') or 
                         metadata.get('filename') or
                         doc.get('id'))
                
                if not doc_id:
                    logger.warning(f"⚠️ No document ID found for chunk: {doc}")
                    continue
                
                # Extract original document info
                filename = metadata.get('filename', 'Unknown File')
                title = metadata.get('title', filename)
                file_type = metadata.get('file_type', 'unknown')
                upload_timestamp = metadata.get('upload_timestamp', metadata.get('timestamp', 0))
                content_hash = metadata.get('content_hash', '')
                
                if doc_id not in document_groups:
                    document_groups[doc_id] = {
                        'document_id': doc_id,
                        'filename': filename,
                        'title': title,
                        'file_type': file_type,
                        'upload_timestamp': upload_timestamp,
                        'content_hash': content_hash,
                        'collection': collection_name,
                        'chunks': [],
                        'total_chunks': 0,
                        'total_size': 0
                    }
                
                # Add chunk info
                document_groups[doc_id]['chunks'].append({
                    'chunk_id': doc['id'],
                    'content_preview': doc.get('content', '')[:200] + '...' if doc.get('content') else 'No content',
                    'score': doc.get('score', 0.0),
                    'metadata': metadata
                })
                
                document_groups[doc_id]['total_chunks'] += 1
                if doc.get('content'):
                    document_groups[doc_id]['total_size'] += len(doc.get('content', ''))
            
            # Convert to list and sort by upload time (newest first)
            original_documents = list(document_groups.values())
            original_documents.sort(key=lambda x: x['upload_timestamp'], reverse=True)
            
            logger.info(f"📋 Found {len(original_documents)} original documents with {sum(d['total_chunks'] for d in original_documents)} total chunks")
            
            return original_documents
            
        except Exception as e:
            logger.error(f"❌ Failed to get original documents: {str(e)}")
            return []
    
    async def delete_document(self, collection_name: str, doc_id: str):
        """Delete a single chunk from a collection by its chunk ID"""
        try:
            # VittoriaDB delete implementation
            collection = self.vectorstore.db.get_collection(collection_name)
            collection.delete(doc_id)
            logger.info(f"✅ Deleted chunk '{doc_id}' from '{collection_name}'")
        except Exception as e:
            logger.error(f"❌ Failed to delete chunk: {e}")
            raise
    
    async def delete_document_by_id(self, document_id: str, collection_name: str):
        """Delete a document and all its chunks by document_id"""
        try:
            # Get all chunks for this document
            original_docs = await self.get_original_documents(collection_name)
            
            # Debug: Log available document IDs
            available_ids = [doc['document_id'] for doc in original_docs]
            logger.info(f"🔍 Looking for document_id='{document_id}' in collection '{collection_name}'")
            logger.info(f"🔍 Available document_ids: {available_ids}")
            
            # Find the document
            target_doc = None
            for doc in original_docs:
                if doc['document_id'] == document_id:
                    target_doc = doc
                    break
            
            if not target_doc:
                logger.warning(f"⚠️ Document '{document_id}' not found in '{collection_name}'")
                logger.warning(f"⚠️ Available documents: {[{'id': d['document_id'], 'file': d['filename']} for d in original_docs]}")
                return {
                    'success': False,
                    'message': f"Document not found. Available IDs: {', '.join(available_ids)}",
                    'deleted_count': 0
                }
            
            # Delete all chunks
            deleted_count = 0
            failed_chunks = []
            
            for chunk in target_doc['chunks']:
                try:
                    await self.delete_document(collection_name, chunk['chunk_id'])
                    deleted_count += 1
                except Exception as e:
                    logger.error(f"❌ Failed to delete chunk {chunk['chunk_id']}: {e}")
                    failed_chunks.append(chunk['chunk_id'])
            
            logger.info(f"✅ Deleted document '{document_id}': {deleted_count} chunks, {len(failed_chunks)} failed")
            
            return {
                'success': len(failed_chunks) == 0,
                'message': f"Deleted {deleted_count} chunks" + (f", {len(failed_chunks)} failed" if failed_chunks else ""),
                'deleted_count': deleted_count,
                'failed_count': len(failed_chunks),
                'failed_chunks': failed_chunks
            }
            
        except Exception as e:
            logger.error(f"❌ Failed to delete document by ID: {e}")
            raise
    
    def close(self):
        """Close connections and cleanup resources"""
        try:
            # VittoriaDB SDK handles cleanup internally
            logger.info("✅ RAG System V2 closed successfully")
        except Exception as e:
            logger.error(f"❌ Failed to close RAG system: {e}")


# Global instance
_rag_system: Optional[RAGSystemV2] = None


def get_rag_system() -> RAGSystemV2:
    """Get or create global RAG system instance"""
    global _rag_system
    if _rag_system is None:
        openai_key = os.getenv('OPENAI_API_KEY')
        vittoriadb_url = os.getenv('VITTORIADB_URL', 'http://localhost:8080')
        _rag_system = RAGSystemV2(
            vittoriadb_url=vittoriadb_url,
            openai_api_key=openai_key
        )
    return _rag_system

