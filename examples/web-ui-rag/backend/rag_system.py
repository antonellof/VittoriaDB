"""
VittoriaDB RAG System
Core RAG functionality with VittoriaDB integration
"""

import os
import logging
from typing import List, Dict, Any, Optional, Tuple
from dataclasses import dataclass
import asyncio
import time

import vittoriadb
from vittoriadb.configure import Configure
from vittoriadb.types import IndexType, DistanceMetric, ContentStorageConfig
import openai
from openai import OpenAI

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class SearchResult:
    """Search result from vector database"""
    content: str
    metadata: Dict[str, Any]
    score: float
    source: str

@dataclass
class ChatMessage:
    """Chat message structure"""
    role: str  # 'user', 'assistant', 'system'
    content: str
    timestamp: float
    sources: Optional[List[SearchResult]] = None

class RAGSystem:
    """Advanced RAG system powered by VittoriaDB"""
    
    def __init__(self, 
                 vittoriadb_url: str = "http://localhost:8080",
                 openai_api_key: Optional[str] = None,
                 use_ollama: bool = True):
        """Initialize RAG system"""
        self.vittoriadb_url = vittoriadb_url
        self.use_ollama = use_ollama
        self.db = None
        self.collections = {}
        
        # Initialize OpenAI client if API key provided
        self.openai_client = None
        if openai_api_key:
            self.openai_client = OpenAI(api_key=openai_api_key)
        
        # Collection configurations - Use 384D for sentence-transformers compatibility
        # auto_embeddings() defaults to sentence-transformers with 384 dimensions
        self.collection_configs = {
            'documents': {
                'dimensions': 384,  # sentence-transformers default
                'description': 'User uploaded documents'
            },
            'web_research': {
                'dimensions': 384,
                'description': 'Web research results'
            },
            'github_code': {
                'dimensions': 384,
                'description': 'GitHub repository code'
            },
            'chat_history': {
                'dimensions': 384,
                'description': 'Chat conversation history'
            }
        }
        
        self._initialize_db()
    
    def _initialize_db(self):
        """Initialize VittoriaDB connection and collections"""
        try:
            # Connect to VittoriaDB
            self.db = vittoriadb.connect(
                url=self.vittoriadb_url,
                auto_start=True  # Auto-start server if not running
            )
            logger.info(f"✅ Connected to VittoriaDB at {self.vittoriadb_url}")
            
            # Create collections with HNSW indexing and content storage for better performance
            for name, config in self.collection_configs.items():
                try:
                    collection = self.db.create_collection(
                        name=name,
                        dimensions=config['dimensions'],
                        metric=DistanceMetric.COSINE,
                        index_type=IndexType.HNSW,  # Use HNSW for fast search
                        config={
                            "m": 16,                # HNSW parameter: number of connections
                            "ef_construction": 200, # Build quality vs speed
                            "ef_search": 50        # Search quality vs speed
                        },
                        vectorizer_config=Configure.Vectors.auto_embeddings(),
                        content_storage=ContentStorageConfig(  # NEW: Enhanced content storage
                            enabled=True,
                            field_name="_content",
                            max_size=2097152,  # 2MB limit for documents
                            compressed=False
                        )
                    )
                    self.collections[name] = collection
                    logger.info(f"✅ Collection '{name}' ready with HNSW indexing")
                except Exception as e:
                    if "already exists" in str(e):
                        # Collection exists, get reference
                        self.collections[name] = self.db.get_collection(name)
                        logger.info(f"✅ Collection '{name}' loaded")
                    else:
                        logger.error(f"❌ Failed to create collection '{name}': {e}")
                        
        except Exception as e:
            logger.error(f"❌ Failed to initialize VittoriaDB: {e}")
            raise
    
    async def add_document(self, 
                          content: str, 
                          metadata: Dict[str, Any],
                          collection_name: str = 'documents') -> str:
        """Add document to vector database"""
        try:
            doc_id = f"{collection_name}_{int(time.time())}_{hash(content) % 10000}"
            
            collection = self.collections.get(collection_name)
            if not collection:
                raise ValueError(f"Collection '{collection_name}' not found")
            
            # Add document with automatic embedding
            collection.insert_text(
                id=doc_id,
                text=content,
                metadata={
                    **metadata,
                    'timestamp': time.time(),
                    'collection': collection_name
                }
            )
            
            logger.info(f"✅ Added document {doc_id} to {collection_name}")
            return doc_id
            
        except Exception as e:
            logger.error(f"❌ Failed to add document: {e}")
            raise
    
    async def search_knowledge_base(self, 
                                   query: str,
                                   collections: List[str] = None,
                                   limit: int = 5,
                                   min_score: float = 0.3) -> List[SearchResult]:
        """Fast concurrent search across knowledge base collections"""
        if collections is None:
            collections = ['documents', 'web_research', 'github_code']
        
        # Special handling for document listing queries
        listing_keywords = ['what documents', 'list documents', 'show documents', 'documents do I have', 'knowledge base']
        is_listing_query = any(keyword in query.lower() for keyword in listing_keywords)
        
        # Create concurrent search tasks
        search_tasks = []
        
        for collection_name in collections:
            collection = self.collections.get(collection_name)
            if not collection:
                continue
            
            # Create async task for each collection search
            task = asyncio.create_task(
                self._search_single_collection(
                    collection, collection_name, query, 
                    limit, min_score, is_listing_query
                )
            )
            search_tasks.append(task)
        
        # Wait for all searches to complete concurrently
        try:
            search_results_lists = await asyncio.gather(*search_tasks, return_exceptions=True)
        except Exception as e:
            logger.error(f"❌ Concurrent search failed: {e}")
            return []
        
        # Combine results from all collections
        all_results = []
        for results in search_results_lists:
            if isinstance(results, list):  # Skip exceptions
                all_results.extend(results)
        
        # Sort by score descending and remove duplicates
        seen_items = set()
        unique_results = []
        
        for result in sorted(all_results, key=lambda x: x.score, reverse=True):
            # Create unique identifier based on content type
            if result.metadata.get('filename'):
                # For uploaded documents: use filename + chunk_index
                item_id = result.metadata.get('filename', '') + str(result.metadata.get('chunk_index', 0))
            elif result.metadata.get('url'):
                # For web research: use URL + title
                item_id = result.metadata.get('url', '') + result.metadata.get('title', '')
            else:
                # Fallback: use title + source + timestamp
                item_id = result.metadata.get('title', '') + result.source + str(result.metadata.get('timestamp', 0))
            
            if item_id not in seen_items:
                seen_items.add(item_id)
                unique_results.append(result)
        
        return unique_results[:limit]
    
    async def _search_single_collection(self, collection, collection_name: str, query: str, 
                                       limit: int, min_score: float, is_listing_query: bool) -> List[SearchResult]:
        """Search a single collection (for concurrent execution)"""
        try:
            if is_listing_query and collection_name == 'documents':
                # For listing queries, use a broader search
                search_queries = ['document', 'file', 'content', 'information']
                for search_term in search_queries:
                    try:
                        results = collection.search_text(
                            query=search_term,
                            limit=limit * 2,
                            include_metadata=True
                        )
                        
                        search_results = []
                        for result in results:
                            if result.score >= 0.1:  # Lower threshold for listing
                                search_result = SearchResult(
                                    content=f"Document: {result.metadata.get('filename', 'Unknown')} - {result.metadata.get('content', 'No content')[:200]}...",
                                    metadata=result.metadata,
                                    score=result.score,
                                    source=collection_name
                                )
                                search_results.append(search_result)
                        return search_results
                    except:
                        continue
                return []
            else:
                # Regular semantic search with content retrieval
                results = collection.search_text(
                    query=query,
                    limit=limit,
                    include_metadata=True,
                    include_content=True  # NEW: Retrieve original content directly
                )
                
                search_results = []
                for result in results:
                    if result.score >= min_score:
                        # Get content from multiple possible sources (prioritizing new content field)
                        content = None
                        
                        # First priority: VittoriaDB's new content field (from enhanced search)
                        if hasattr(result, 'content') and result.content:
                            content = result.content
                        # Second priority: VittoriaDB's content storage in metadata
                        elif result.metadata and '_content' in result.metadata and result.metadata['_content']:
                            content = result.metadata['_content']
                        # Third priority: Legacy content field in metadata
                        elif result.metadata and 'content' in result.metadata and result.metadata['content']:
                            content = result.metadata['content']
                        # Fourth priority: Use snippet if available (for web search results)
                        elif result.metadata and 'snippet' in result.metadata and result.metadata['snippet']:
                            # For web search results, enhance snippet with title and URL
                            snippet = result.metadata['snippet']
                            title = result.metadata.get('title', '')
                            url = result.metadata.get('url', '')
                            
                            if title and url:
                                content = f"Title: {title}\nURL: {url}\n\nContent: {snippet}"
                            elif title:
                                content = f"Title: {title}\n\nContent: {snippet}"
                            else:
                                content = snippet
                        # Last resort: use title
                        elif result.metadata and result.metadata.get('title'):
                            content = f"Title: {result.metadata['title']}"
                        else:
                            content = 'No content available'
                        
                        search_result = SearchResult(
                            content=content,
                            metadata=result.metadata,
                            score=result.score,
                            source=collection_name
                        )
                        search_results.append(search_result)
                
                return search_results
                
        except Exception as e:
            logger.error(f"❌ Search failed in {collection_name}: {e}")
            return []
    
    async def generate_response(self, 
                               user_query: str,
                               context_results: List[SearchResult],
                               chat_history: List[ChatMessage] = None,
                               model: str = "gpt-3.5-turbo") -> str:
        """Generate AI response using retrieved context"""
        
        if not self.openai_client:
            return "❌ OpenAI API key not configured. Please set OPENAI_API_KEY environment variable."
        
        # Build context from search results
        context_text = ""
        sources = []
        
        for result in context_results:
            context_text += f"\n--- Source: {result.source} (Score: {result.score:.3f}) ---\n"
            context_text += result.content
            sources.append(f"{result.source}: {result.metadata.get('title', 'Unknown')}")
        
        # Build chat history context
        history_context = ""
        if chat_history:
            for msg in chat_history[-5:]:  # Last 5 messages
                history_context += f"{msg.role}: {msg.content}\n"
        
        # Create system prompt
        system_prompt = f"""You are an intelligent assistant with access to a knowledge base. 
        Use the provided context to answer questions accurately and helpfully.
        
        CONTEXT FROM KNOWLEDGE BASE:
        {context_text}
        
        RECENT CHAT HISTORY:
        {history_context}
        
        Guidelines:
        - Answer based on the provided context when relevant
        - If context doesn't contain the answer, say so clearly
        - Cite sources when using specific information
        - Be concise but comprehensive
        - If asked about sources, mention: {', '.join(sources)}
        """
        
        try:
            # Generate response using OpenAI (default to GPT-4)
            if model == 'gpt-3.5-turbo':
                model = 'gpt-4'  # Upgrade to GPT-4 for better responses
                
            response = self.openai_client.chat.completions.create(
                model=model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_query}
                ],
                temperature=0.7,
                max_tokens=1500  # Increased for GPT-4
            )
                
            return response.choices[0].message.content
            
        except Exception as e:
            logger.error(f"❌ Failed to generate response: {e}")
            return f"❌ Error generating response: {str(e)}"
    
    async def chat(self, 
                   user_message: str,
                   chat_history: List[ChatMessage] = None,
                   search_collections: List[str] = None) -> Tuple[str, List[SearchResult]]:
        """Complete chat workflow: search + generate response"""
        
        # Search knowledge base
        search_results = await self.search_knowledge_base(
            query=user_message,
            collections=search_collections,
            limit=5
        )
        
        # Generate response
        response = await self.generate_response(
            user_query=user_message,
            context_results=search_results,
            chat_history=chat_history
        )
        
        # Store chat interaction in history collection
        try:
            await self.add_document(
                content=f"User: {user_message}\nAssistant: {response}",
                metadata={
                    'type': 'chat_interaction',
                    'user_query': user_message,
                    'assistant_response': response,
                    'search_results_count': len(search_results)
                },
                collection_name='chat_history'
            )
        except Exception as e:
            logger.warning(f"Failed to store chat history: {e}")
        
        return response, search_results
    
    def get_collection_stats(self) -> Dict[str, Any]:
        """Get statistics for all collections"""
        import requests
        stats = {}
        
        for name, collection in self.collections.items():
            try:
                # Get detailed collection info directly from VittoriaDB API
                response = requests.get(f"{self.vittoriadb_url}/collections/{name}")
                if response.status_code == 200:
                    api_info = response.json()
                    
                    # Map index_type number to string
                    index_type_map = {0: "flat", 1: "hnsw"}
                    index_type = index_type_map.get(api_info.get('index_type', 0), "unknown")
                    
                    # Map metric number to string  
                    metric_map = {0: "cosine", 1: "euclidean", 2: "dot"}
                    metric = metric_map.get(api_info.get('metric', 0), "unknown")
                    
                    stats[name] = {
                        'name': api_info['name'],
                        'vector_count': api_info['vector_count'],
                        'dimensions': api_info['dimensions'],
                        'metric': metric,
                        'index_type': index_type,
                        'description': self.collection_configs[name]['description']
                    }
                else:
                    # Fallback to Python SDK info
                    collection._info = None
                    info = collection.info
                    stats[name] = {
                        'name': info.name,
                        'vector_count': info.vector_count,
                        'dimensions': info.dimensions,
                        'metric': info.metric.to_string(),
                        'index_type': 'unknown',
                        'description': self.collection_configs[name]['description']
                    }
            except Exception as e:
                stats[name] = {'error': str(e)}
        
        return stats
    
    def close(self):
        """Close database connection"""
        if self.db:
            self.db.close()
            logger.info("✅ VittoriaDB connection closed")

# Global RAG system instance
_rag_system = None

def get_rag_system() -> RAGSystem:
    """Get or create global RAG system instance"""
    global _rag_system
    if _rag_system is None:
        openai_key = os.getenv('OPENAI_API_KEY')
        _rag_system = RAGSystem(openai_api_key=openai_key)
    return _rag_system
