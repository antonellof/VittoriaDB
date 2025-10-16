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
    
    async def add_documents_batch(
        self,
        collection_name: str,
        documents: List[Dict[str, Any]]
    ):
        """
        Add multiple documents in batch.
        Compatible with old interface.
        """
        for doc in documents:
            await self.add_document(
                collection_name=collection_name,
                doc_id=doc.get('id', doc.get('doc_id', f"doc_{hash(doc['content'])}")),
                content=doc['content'],
                metadata=doc.get('metadata', {})
            )
    
    def get_collection_stats(self) -> Dict[str, Any]:
        """Get statistics for all collections"""
        stats = {}
        for collection_name in self.collections:
            try:
                # VittoriaDB doesn't have a direct count method
                # We'll return the collection names for now
                stats[collection_name] = {
                    'name': collection_name,
                    'status': 'active'
                }
            except Exception as e:
                logger.warning(f"⚠️ Could not get stats for '{collection_name}': {e}")
        
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
    
    async def delete_document(self, collection_name: str, doc_id: str):
        """Delete a document from a collection"""
        try:
            # VittoriaDB delete implementation
            collection = self.vectorstore.db.get_collection(collection_name)
            collection.delete(doc_id)
            logger.info(f"✅ Deleted document '{doc_id}' from '{collection_name}'")
        except Exception as e:
            logger.error(f"❌ Failed to delete document: {e}")
            raise


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

