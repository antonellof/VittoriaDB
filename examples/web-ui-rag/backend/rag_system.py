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
from openai import AsyncOpenAI

# Simple embedding service using HTTP requests (no complex dependencies)
import httpx
import numpy as np
import re

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

def chunk_text(text: str, max_tokens: int = 6000, overlap: int = 200) -> List[str]:
    """
    Split text into chunks that fit within OpenAI token limits.
    
    Args:
        text: Text to chunk
        max_tokens: Maximum tokens per chunk (conservative estimate: ~4 chars = 1 token)
        overlap: Number of characters to overlap between chunks
    
    Returns:
        List of text chunks
    """
    if not text or len(text) == 0:
        return []
    
    # Conservative estimate: 4 characters ≈ 1 token
    max_chars = max_tokens * 4
    
    # If text is short enough, return as single chunk
    if len(text) <= max_chars:
        return [text]
    
    chunks = []
    start = 0
    
    while start < len(text):
        # Calculate end position
        end = start + max_chars
        
        # If this is not the last chunk, try to break at a sentence or paragraph
        if end < len(text):
            # Look for sentence endings within the last 500 characters
            search_start = max(start + max_chars - 500, start)
            
            # Try to find sentence endings
            sentence_endings = []
            for match in re.finditer(r'[.!?]\s+', text[search_start:end]):
                sentence_endings.append(search_start + match.end())
            
            if sentence_endings:
                end = sentence_endings[-1]  # Use the last sentence ending
            else:
                # Try to break at paragraph
                paragraph_breaks = []
                for match in re.finditer(r'\n\s*\n', text[search_start:end]):
                    paragraph_breaks.append(search_start + match.start())
                
                if paragraph_breaks:
                    end = paragraph_breaks[-1]
                else:
                    # Try to break at word boundary
                    word_boundary = text.rfind(' ', search_start, end)
                    if word_boundary > start:
                        end = word_boundary
        
        # Extract chunk
        chunk = text[start:end].strip()
        if chunk:
            chunks.append(chunk)
        
        # Move start position with overlap
        start = max(end - overlap, start + 1)
        
        # Prevent infinite loop
        if start >= len(text):
            break
    
    return chunks

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
            self.openai_client = AsyncOpenAI(api_key=openai_api_key)
        
        # Simple embedding service (no complex dependencies)
        self.simple_embedder = True
        logger.info("✅ Simple embedding service initialized")
        
        # Collection configurations - Use 1536D for OpenAI embeddings (much faster)
        # OpenAI embeddings are 10-50x faster than local sentence transformers
        self.collection_configs = {
            'documents': {
                'dimensions': 1536,  # OpenAI text-embedding-ada-002 dimensions
                'description': 'User uploaded documents'
            },
            'web_research': {
                'dimensions': 1536,
                'description': 'Web research results'
            },
            'github_code': {
                'dimensions': 1536,
                'description': 'GitHub repository code'
            },
            'chat_history': {
                'dimensions': 1536,
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
                    # Create collection with OpenAI vectorizer for ultra-fast text search
                    import os
                    from vittoriadb.configure import Configure
                    
                    # Use OpenAI embeddings for 10-50x speed improvement
                    openai_key = os.getenv('OPENAI_API_KEY')
                    if openai_key:
                        vectorizer_config = Configure.Vectors.openai_embeddings(api_key=openai_key)
                        logger.info(f"🚀 Using OpenAI embeddings for collection '{name}' (ultra-fast)")
                    else:
                        # Fallback to Hugging Face remote API (still faster than local)
                        vectorizer_config = Configure.Vectors.huggingface_embeddings()
                        logger.warning(f"⚠️ No OpenAI key, using Hugging Face remote API for '{name}'")
                    
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
                        vectorizer_config=vectorizer_config
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
        """Add document to vector database with automatic chunking for large texts"""
        try:
            collection = self.collections.get(collection_name)
            if not collection:
                raise ValueError(f"Collection '{collection_name}' not found")
            
            base_metadata = {
                **metadata,
                'timestamp': time.time(),
                'collection': collection_name
            }
            
            # Chunk large texts to fit within OpenAI token limits
            chunks = chunk_text(content, max_tokens=6000, overlap=200)
            
            if len(chunks) == 1:
                # Single chunk - use original approach
                doc_id = f"{collection_name}_{int(time.time())}_{hash(content) % 10000}"
                
                loop = asyncio.get_event_loop()
                await loop.run_in_executor(
                    None,
                    collection.insert_text,
                    doc_id,
                    content,
                    base_metadata
                )
                
                logger.info(f"✅ Added document {doc_id} to {collection_name}")
                return doc_id
            else:
                # Multiple chunks - insert each chunk separately
                base_doc_id = f"{collection_name}_{int(time.time())}_{hash(content) % 10000}"
                loop = asyncio.get_event_loop()
                
                for i, chunk in enumerate(chunks):
                    chunk_id = f"{base_doc_id}_chunk_{i}"
                    chunk_metadata = {
                        **base_metadata,
                        'chunk_index': i,
                        'total_chunks': len(chunks),
                        'is_chunk': True,
                        'original_doc_id': base_doc_id
                    }
                    
                    await loop.run_in_executor(
                        None,
                        collection.insert_text,
                        chunk_id,
                        chunk,
                        chunk_metadata
                    )
                
                logger.info(f"✅ Added document {base_doc_id} ({len(chunks)} chunks) to {collection_name}")
                return base_doc_id
            
        except Exception as e:
            logger.error(f"❌ Failed to add document: {e}")
            raise
    
    async def add_documents_batch(self, 
                                 documents: List[Dict[str, Any]], 
                                 collection_name: str = 'documents') -> List[str]:
        """Add multiple documents to vector database using high-performance batch processing"""
        try:
            collection = self.collections.get(collection_name)
            if not collection:
                raise ValueError(f"Collection '{collection_name}' not found")
            
            # Prepare document data with chunking for large texts
            doc_ids = []
            texts = []
            metadatas = []
            
            for doc in documents:
                content = doc['content']
                base_metadata = {
                    **doc['metadata'],
                    'timestamp': time.time(),
                    'collection': collection_name
                }
                
                # Chunk large texts to fit within OpenAI token limits
                chunks = chunk_text(content, max_tokens=6000, overlap=200)
                
                if len(chunks) == 1:
                    # Single chunk - use original document ID
                    doc_id = f"{collection_name}_{int(time.time())}_{hash(content) % 10000}"
                    doc_ids.append(doc_id)
                    texts.append(content)
                    metadatas.append(base_metadata)
                else:
                    # Multiple chunks - create separate documents for each chunk
                    base_doc_id = f"{collection_name}_{int(time.time())}_{hash(content) % 10000}"
                    for i, chunk in enumerate(chunks):
                        chunk_id = f"{base_doc_id}_chunk_{i}"
                        doc_ids.append(chunk_id)
                        texts.append(chunk)
                        
                        # Add chunk metadata
                        chunk_metadata = {
                            **base_metadata,
                            'chunk_index': i,
                            'total_chunks': len(chunks),
                            'is_chunk': True,
                            'original_doc_id': base_doc_id
                        }
                        metadatas.append(chunk_metadata)
            
            # Use VittoriaDB's built-in vectorizer for text insertion (now that we have vectorizers)
            logger.info(f"🚀 Using VittoriaDB vectorizer for batch insertion of {len(texts)} documents")
            loop = asyncio.get_event_loop()
            
            def batch_insert_native():
                """Fallback native SDK batch insertion"""
                inserted_count = 0
                for i, (doc_id, text, metadata) in enumerate(zip(doc_ids, texts, metadatas)):
                    try:
                        collection.insert_text(
                            id=doc_id,
                            text=text,
                            metadata=metadata
                        )
                        inserted_count += 1
                    except Exception as e:
                        logger.error(f"Failed to insert document {i}: {e}")
                return inserted_count
            
            # Run in thread pool to prevent blocking the event loop
            inserted_count = await loop.run_in_executor(None, batch_insert_native)
            
            logger.info(f"✅ Native SDK batch inserted {inserted_count}/{len(documents)} documents to {collection_name}")
            return doc_ids[:inserted_count]
            
        except Exception as e:
            logger.error(f"❌ Failed to batch add documents: {e}")
            raise
    
    async def _batch_insert_with_openai(self, collection, doc_ids: List[str], 
                                       texts: List[str], metadatas: List[Dict]) -> List[str]:
        """High-performance batch insertion using OpenAI embeddings"""
        try:
            # Generate embeddings in batch (much faster than individual)
            logger.info(f"🚀 Generating embeddings for {len(texts)} documents using OpenAI...")
            
            # Generate embeddings directly (no need for thread pool since OpenAI client is async)
            all_embeddings = []
            
            # Process in smaller chunks to avoid token limits
            chunk_size = 20  # Process 20 texts at a time
            for i in range(0, len(texts), chunk_size):
                chunk_texts = texts[i:i + chunk_size]
                
                # Truncate texts that are too long (rough estimate: 1 token ≈ 4 chars)
                truncated_texts = []
                for text in chunk_texts:
                    if len(text) > 6000:  # ~1500 tokens
                        truncated_texts.append(text[:6000] + "...")
                    else:
                        truncated_texts.append(text)
                
                response = await self.openai_client.embeddings.create(
                    model="text-embedding-ada-002",
                    input=truncated_texts
                )
                all_embeddings.extend([data.embedding for data in response.data])
            
            embeddings = all_embeddings
            
            # Insert pre-computed vectors (bypasses slow embedding generation)
            loop = asyncio.get_event_loop()
            
            def batch_insert_vectors():
                """Insert pre-computed vectors directly"""
                inserted_count = 0
                for doc_id, embedding, metadata in zip(doc_ids, embeddings, metadatas):
                    try:
                        collection.insert(
                            id=doc_id,
                            vector=embedding,
                            metadata=metadata
                        )
                        inserted_count += 1
                    except Exception as e:
                        logger.error(f"Failed to insert vector for {doc_id}: {e}")
                return inserted_count
            
            inserted_count = await loop.run_in_executor(None, batch_insert_vectors)
            
            logger.info(f"🚀 OpenAI batch inserted {inserted_count} documents with pre-computed embeddings")
            return doc_ids[:inserted_count]
            
        except Exception as e:
            logger.error(f"❌ OpenAI batch insertion failed: {e}")
            raise
    
    async def _batch_insert_with_simple_embedder(self, collection, doc_ids: List[str], 
                                                texts: List[str], metadatas: List[Dict]) -> List[str]:
        """High-performance batch insertion using simple random embeddings (for testing)"""
        try:
            # Generate simple embeddings (for performance testing)
            logger.info(f"🚀 Generating simple embeddings for {len(texts)} documents...")
            
            loop = asyncio.get_event_loop()
            
            def generate_simple_embeddings():
                """Generate simple hash-based embeddings for testing"""
                embeddings = []
                # Use 1536 dimensions to match OpenAI collections
                dimensions = 1536
                
                for text in texts:
                    # Create a simple deterministic embedding based on text hash
                    # This is for performance testing only - not for production!
                    text_hash = hash(text)
                    np.random.seed(abs(text_hash) % 2**32)
                    embedding = np.random.random(dimensions).astype(np.float32)
                    embedding = embedding / np.linalg.norm(embedding)  # Normalize
                    embeddings.append(embedding.tolist())
                return embeddings
            
            embeddings = await loop.run_in_executor(None, generate_simple_embeddings)
            
            # Insert pre-computed vectors (bypasses slow embedding generation)
            def batch_insert_vectors():
                """Insert pre-computed vectors directly"""
                inserted_count = 0
                for i, (doc_id, embedding, metadata) in enumerate(zip(doc_ids, embeddings, metadatas)):
                    try:
                        # Add content to metadata for search results
                        metadata_with_content = {
                            **metadata,
                            '_content': texts[i]  # Store original text
                        }
                        
                        collection.insert(
                            id=doc_id,
                            vector=embedding,
                            metadata=metadata_with_content
                        )
                        inserted_count += 1
                    except Exception as e:
                        logger.error(f"Failed to insert vector for {doc_id}: {e}")
                return inserted_count
            
            inserted_count = await loop.run_in_executor(None, batch_insert_vectors)
            
            logger.info(f"🚀 Simple embedder batch inserted {inserted_count} documents with pre-computed embeddings")
            return doc_ids[:inserted_count]
            
        except Exception as e:
            logger.error(f"❌ Simple embedder batch insertion failed: {e}")
            raise
    
    async def search_knowledge_base(self, 
                                   query: str,
                                   collections: List[str] = None,
                                   limit: int = 5,
                                   min_score: float = 0.3) -> List[SearchResult]:
        """Fast concurrent search across knowledge base collections"""
        if collections is None:
            collections = ['documents', 'web_research', 'github_code']
        
        # Special handling for knowledge base overview queries
        overview_keywords = [
            'what documents', 'list documents', 'show documents', 'documents do I have', 
            'knowledge base', 'what files', 'show files', 'what content', 'overview',
            'what information', 'what data', 'show me everything', 'all documents',
            'contents of', 'what\'s in', 'inventory', 'catalog'
        ]
        is_overview_query = any(keyword in query.lower() for keyword in overview_keywords)
        
        # For overview queries, increase limit and lower score threshold
        if is_overview_query:
            limit = min(50, limit * 10)  # Increase limit significantly but cap at 50
            min_score = max(0.1, min_score - 0.2)  # Lower score threshold for broader results
            logger.info(f"🔍 Knowledge base overview query detected - expanding search to {limit} results with min_score {min_score}")
        
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
                    limit, min_score, is_overview_query
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
                                       limit: int, min_score: float, is_overview_query: bool) -> List[SearchResult]:
        """Search a single collection (for concurrent execution)"""
        try:
            if is_overview_query:
                # For overview queries, use broader search terms to get comprehensive results
                if collection_name == 'documents':
                    search_queries = ['document', 'file', 'content', 'information', 'text']
                elif collection_name == 'github_code':
                    search_queries = ['code', 'repository', 'file', 'function', 'class']
                elif collection_name == 'web_research':
                    search_queries = ['research', 'web', 'information', 'data', 'content']
                else:
                    search_queries = ['content', 'information', 'data']
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
                    include_content=True
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
                
            response = await self.openai_client.chat.completions.create(
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
    
    async def get_stats(self) -> Dict[str, Any]:
        """Get RAG system statistics (async wrapper for get_collection_stats)"""
        return self.get_collection_stats()
    
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
