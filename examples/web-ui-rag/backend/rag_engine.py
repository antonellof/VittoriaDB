#!/usr/bin/env python3
"""
Advanced RAG Engine using VittoriaDB as Vector Database

This implements a complete RAG system following best practices:
1. Document chunking with overlap
2. Embedding generation and storage
3. Semantic retrieval with scoring
4. Context management and source attribution
5. Response generation with citations

Based on AI SDK RAG patterns and VittoriaDB examples.
"""

import asyncio
import hashlib
import json
import re
import time
from dataclasses import dataclass, asdict
from typing import List, Dict, Any, Optional, Tuple, AsyncGenerator
from pathlib import Path
import logging

# Fix tokenizers multiprocessing warning
import os
os.environ["TOKENIZERS_PARALLELISM"] = "false"

import vittoriadb
from vittoriadb import DistanceMetric, IndexType, VectorizerConfig, VectorizerType
from vittoriadb.configure import Configure
import openai
import httpx
from sentence_transformers import SentenceTransformer

logger = logging.getLogger(__name__)

@dataclass
class DocumentChunk:
    """Represents a chunk of a document with metadata."""
    id: str
    content: str
    document_id: str
    document_title: str
    chunk_index: int
    chunk_size: int
    start_char: int
    end_char: int
    metadata: Dict[str, Any]
    embedding: Optional[List[float]] = None

@dataclass
class RetrievalResult:
    """Result from semantic search with source attribution."""
    chunk: DocumentChunk
    score: float
    rank: int
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            'chunk': {
                'id': self.chunk.id,
                'content': self.chunk.content,
                'document_id': self.chunk.document_id,
                'document_title': self.chunk.document_title,
                'chunk_index': self.chunk.chunk_index,
                'chunk_size': self.chunk.chunk_size,
                'start_char': self.chunk.start_char,
                'end_char': self.chunk.end_char,
                'metadata': self.chunk.metadata
            },
            'score': self.score,
            'rank': self.rank
        }

@dataclass
class RAGResponse:
    """Complete RAG response with sources and metadata."""
    answer: str
    sources: List[RetrievalResult]
    query: str
    total_chunks_searched: int
    retrieval_time: float
    generation_time: float
    model_used: str

class DocumentChunker:
    """Advanced document chunking with overlap and semantic boundaries."""
    
    def __init__(self, chunk_size: int = 1000, chunk_overlap: int = 200):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
    
    def chunk_text(self, text: str, document_id: str, document_title: str, 
                   metadata: Dict[str, Any] = None) -> List[DocumentChunk]:
        """
        Chunk text into overlapping segments with semantic boundaries.
        
        Args:
            text: Text to chunk
            document_id: Unique document identifier
            document_title: Document title for attribution
            metadata: Additional metadata to attach to chunks
            
        Returns:
            List of DocumentChunk objects
        """
        if metadata is None:
            metadata = {}
            
        # Clean and normalize text
        text = self._clean_text(text)
        
        # Split into sentences for better semantic boundaries
        sentences = self._split_into_sentences(text)
        
        chunks = []
        current_chunk = ""
        current_start = 0
        chunk_index = 0
        
        for sentence in sentences:
            # Check if adding this sentence would exceed chunk size
            if len(current_chunk) + len(sentence) > self.chunk_size and current_chunk:
                # Create chunk
                chunk_id = self._generate_chunk_id(document_id, chunk_index)
                chunk = DocumentChunk(
                    id=chunk_id,
                    content=current_chunk.strip(),
                    document_id=document_id,
                    document_title=document_title,
                    chunk_index=chunk_index,
                    chunk_size=len(current_chunk),
                    start_char=current_start,
                    end_char=current_start + len(current_chunk),
                    metadata={**metadata, 'chunking_method': 'sentence_boundary'}
                )
                chunks.append(chunk)
                
                # Handle overlap
                overlap_text = self._get_overlap_text(current_chunk, self.chunk_overlap)
                current_chunk = overlap_text + " " + sentence
                current_start = current_start + len(current_chunk) - len(overlap_text) - 1
                chunk_index += 1
            else:
                current_chunk += " " + sentence if current_chunk else sentence
        
        # Add final chunk if there's remaining content
        if current_chunk.strip():
            chunk_id = self._generate_chunk_id(document_id, chunk_index)
            chunk = DocumentChunk(
                id=chunk_id,
                content=current_chunk.strip(),
                document_id=document_id,
                document_title=document_title,
                chunk_index=chunk_index,
                chunk_size=len(current_chunk),
                start_char=current_start,
                end_char=current_start + len(current_chunk),
                metadata={**metadata, 'chunking_method': 'sentence_boundary'}
            )
            chunks.append(chunk)
        
        logger.info(f"Chunked document '{document_title}' into {len(chunks)} chunks")
        return chunks
    
    def _clean_text(self, text: str) -> str:
        """Clean and normalize text."""
        # Remove excessive whitespace
        text = re.sub(r'\s+', ' ', text)
        # Remove special characters that might interfere with chunking
        text = re.sub(r'[^\w\s\.\!\?\,\;\:\-\(\)\[\]\{\}\"\']+', ' ', text)
        return text.strip()
    
    def _split_into_sentences(self, text: str) -> List[str]:
        """Split text into sentences using simple regex."""
        # Simple sentence splitting - could be improved with spaCy or NLTK
        sentences = re.split(r'(?<=[.!?])\s+', text)
        return [s.strip() for s in sentences if s.strip()]
    
    def _get_overlap_text(self, text: str, overlap_size: int) -> str:
        """Get the last N characters for overlap."""
        if len(text) <= overlap_size:
            return text
        return text[-overlap_size:]
    
    def _generate_chunk_id(self, document_id: str, chunk_index: int) -> str:
        """Generate unique chunk ID."""
        return f"{document_id}_chunk_{chunk_index}_{int(time.time())}"

class VittoriaRAGEngine:
    """Advanced RAG engine using VittoriaDB as the vector database."""
    
    def __init__(self, 
                 vittoriadb_url: str = "http://localhost:8080",
                 collection_name: str = "rag_knowledge_base",
                 embedding_model: str = "all-MiniLM-L6-v2",
                 openai_api_key: Optional[str] = None):
        """
        Initialize the RAG engine.
        
        Args:
            vittoriadb_url: URL of VittoriaDB server
            collection_name: Name of the collection to use
            embedding_model: SentenceTransformer model name
            openai_api_key: OpenAI API key for response generation
        """
        self.vittoriadb_url = vittoriadb_url
        self.collection_name = collection_name
        self.db = None
        self.collection = None
        
        # Initialize embedding model
        logger.info(f"Loading embedding model: {embedding_model}")
        self.embedding_model = SentenceTransformer(embedding_model)
        self.embedding_dim = self.embedding_model.get_sentence_embedding_dimension()
        logger.info(f"Embedding model loaded: {self.embedding_dim} dimensions")
        
        # Initialize chunker
        self.chunker = DocumentChunker(chunk_size=1000, chunk_overlap=200)
        
        # Initialize OpenAI client
        self.openai_client = None
        if openai_api_key:
            self.openai_client = openai.AsyncOpenAI(api_key=openai_api_key)
            logger.info("Async OpenAI client initialized")
        
        # Document store for metadata
        self.documents: Dict[str, Dict[str, Any]] = {}
    
    async def initialize(self) -> bool:
        """Initialize the RAG engine and create collections."""
        try:
            logger.info("Initializing VittoriaRAG engine...")
            
            # Connect to VittoriaDB
            self.db = vittoriadb.connect(
                url=self.vittoriadb_url,
                auto_start=True
            )
            logger.info(f"Connected to VittoriaDB at {self.vittoriadb_url}")
            
            # Create or get collection with HNSW indexing
            try:
                self.collection = self.db.create_collection(
                    name=self.collection_name,
                    dimensions=self.embedding_dim,
                    metric=DistanceMetric.COSINE,
                    index_type=IndexType.HNSW,
                    config={
                        "m": 16,
                        "ef_construction": 200,
                        "ef_search": 50
                    },
                    vectorizer_config=Configure.Vectors.auto_embeddings(
                        model="all-MiniLM-L6-v2",
                        dimensions=self.embedding_dim
                    )
                )
                logger.info(f"Created collection '{self.collection_name}' with HNSW indexing")
            except Exception as e:
                if "already exists" in str(e):
                    self.collection = self.db.get_collection(self.collection_name)
                    logger.info(f"Using existing collection '{self.collection_name}'")
                else:
                    raise e
            
            return True
            
        except Exception as e:
            logger.error(f"Failed to initialize RAG engine: {e}")
            return False
    
    async def add_document(self, 
                          content: str, 
                          document_id: str, 
                          title: str, 
                          metadata: Dict[str, Any] = None) -> int:
        """
        Add a document to the knowledge base.
        
        Args:
            content: Document content
            document_id: Unique document identifier
            title: Document title
            metadata: Additional metadata
            
        Returns:
            Number of chunks created
        """
        if metadata is None:
            metadata = {}
        
        logger.info(f"Adding document '{title}' to knowledge base")
        
        # Store document metadata
        self.documents[document_id] = {
            'title': title,
            'content_length': len(content),
            'metadata': metadata,
            'added_at': time.time()
        }
        
        # Chunk the document
        chunks = self.chunker.chunk_text(content, document_id, title, metadata)
        
        # Generate embeddings and store chunks
        stored_count = 0
        for chunk in chunks:
            try:
                # Use VittoriaDB's auto-embedding feature
                await self._store_chunk_with_auto_embedding(chunk)
                stored_count += 1
            except Exception as e:
                logger.error(f"Failed to store chunk {chunk.id}: {e}")
        
        logger.info(f"Stored {stored_count}/{len(chunks)} chunks for document '{title}'")
        return stored_count
    
    async def _store_chunk_with_auto_embedding(self, chunk: DocumentChunk):
        """Store chunk using client-side embedding generation for speed."""
        # Generate embedding client-side (FAST)
        chunk_embedding = self.embedding_model.encode(chunk.content).tolist()
        
        # Prepare metadata for storage (include content for retrieval)
        chunk_metadata = {
            'document_id': chunk.document_id,
            'document_title': chunk.document_title,
            'chunk_index': chunk.chunk_index,
            'chunk_size': chunk.chunk_size,
            'start_char': chunk.start_char,
            'end_char': chunk.end_char,
            'content': chunk.content,  # Store content in metadata for retrieval
            **chunk.metadata
        }
        
        # Use direct vector insertion with pre-computed embedding (VERY FAST)
        self.collection.insert(
            id=chunk.id,
            vector=chunk_embedding,
            metadata=chunk_metadata
        )
    
    async def search(self, 
                    query: str, 
                    limit: int = 5, 
                    min_score: float = 0.3,
                    filters: Dict[str, Any] = None) -> List[RetrievalResult]:
        """
        Perform semantic search across multiple collections (knowledge base + web search).
        
        Args:
            query: Search query
            limit: Maximum number of results
            min_score: Minimum similarity score
            filters: Optional metadata filters
            
        Returns:
            List of RetrievalResult objects from all collections
        """
        start_time = time.time()
        
        try:
            # Generate query embedding client-side for FAST search
            embed_start = time.time()
            query_embedding = self.embedding_model.encode(query).tolist()
            embed_time = time.time() - embed_start
            
            # Search across multiple collections
            all_results = []
            collections_to_search = [
                (self.collection, "knowledge_base"),  # Main RAG collection
            ]
            
            # Try to get other collections if they exist
            try:
                web_research_collection = self.db.get_collection("web_research")
                collections_to_search.append((web_research_collection, "web_search"))
            except Exception as e:
                logger.debug(f"web_research collection not available: {e}")
            
            try:
                documents_collection = self.db.get_collection("documents")
                collections_to_search.append((documents_collection, "documents"))
            except Exception as e:
                logger.debug(f"documents collection not available: {e}")
            
            search_start = time.time()
            for collection, source_type in collections_to_search:
                try:
                    # Search each collection
                    collection_results = collection.search(
                        vector=query_embedding,
                        limit=limit,
                        include_metadata=True
                    )
                    
                    # Add source type to metadata
                    for result in collection_results:
                        if result.metadata is None:
                            result.metadata = {}
                        result.metadata['source_collection'] = source_type
                        all_results.append(result)
                        
                except Exception as e:
                    logger.warning(f"Failed to search collection {source_type}: {e}")
                    continue
            
            # Sort all results by score (descending)
            all_results.sort(key=lambda x: x.score, reverse=True)
            
            # Take top results across all collections
            search_results = all_results[:limit]
            vector_search_time = time.time() - search_start
            
            # Convert to RetrievalResult objects
            results = []
            for i, result in enumerate(search_results):
                if result.score >= min_score:
                    # Reconstruct DocumentChunk from search result with enhanced content retrieval
                    # Priority: VittoriaDB v0.4.0 content field > _content metadata > legacy content > title
                    content = ""
                    if hasattr(result, 'content') and result.content:
                        content = result.content
                    elif result.metadata.get('_content'):
                        content = result.metadata['_content']
                    elif result.metadata.get('content'):
                        content = result.metadata['content']
                    else:
                        content = result.metadata.get('title', 'No content available')
                    
                    chunk = DocumentChunk(
                        id=result.id,
                        content=content,
                        document_id=result.metadata.get('document_id', ''),
                        document_title=result.metadata.get('document_title', result.metadata.get('title', '')),
                        chunk_index=result.metadata.get('chunk_index', 0),
                        chunk_size=result.metadata.get('chunk_size', len(content)),
                        start_char=result.metadata.get('start_char', 0),
                        end_char=result.metadata.get('end_char', len(content)),
                        metadata=result.metadata
                    )
                    
                    retrieval_result = RetrievalResult(
                        chunk=chunk,
                        score=result.score,
                        rank=i + 1
                    )
                    results.append(retrieval_result)
            
            total_time = time.time() - start_time
            logger.info(f"Search completed: {len(results)} results in {total_time:.3f}s "
                       f"(embed: {embed_time:.3f}s, search: {vector_search_time:.3f}s)")
            
            return results
            
        except Exception as e:
            logger.error(f"Search failed: {e}")
            return []
    
    async def generate_response(self, 
                               query: str, 
                               search_results: List[RetrievalResult],
                               model: str = "gpt-4",
                               max_context_length: int = 4000) -> str:
        """
        Generate a response using retrieved context.
        
        Args:
            query: User query
            search_results: Retrieved context chunks
            model: OpenAI model to use
            max_context_length: Maximum context length in characters
            
        Returns:
            Generated response
        """
        if not self.openai_client:
            # Fallback to a simple response using the search results
            if not search_results:
                return "I don't have enough information in my knowledge base to answer your question. OpenAI API is not available."
            
            # Generate a simple response from search results
            context_info = []
            for result in search_results[:3]:  # Use top 3 results
                source_collection = result.chunk.metadata.get('source_collection', 'unknown')
                if source_collection == 'web_search':
                    source_type = "🌐 Web Search"
                elif source_collection == 'documents':
                    source_type = "📄 Document"
                else:
                    source_type = "📚 Knowledge Base"
                
                context_info.append(f"{source_type}: {result.chunk.document_title}\nContent: {result.chunk.content[:200]}...\nRelevance Score: {result.score:.3f}")
            
            fallback_response = f"""Based on the search results I found, here's what I can tell you:

{chr(10).join(context_info)}

Note: OpenAI API is currently unavailable (quota exceeded), so I'm providing the raw search results. The search found {len(search_results)} relevant sources with scores ranging from {min(r.score for r in search_results):.3f} to {max(r.score for r in search_results):.3f}."""
            
            return fallback_response
        
        if not search_results:
            return "I don't have enough information in my knowledge base to answer your question."
        
        # Build context from search results
        context_parts = []
        current_length = 0
        
        for result in search_results:
            source_info = f"[Source: {result.chunk.document_title}, Score: {result.score:.3f}]"
            chunk_text = f"{source_info}\n{result.chunk.content}\n"
            
            if current_length + len(chunk_text) > max_context_length:
                break
                
            context_parts.append(chunk_text)
            current_length += len(chunk_text)
        
        context = "\n".join(context_parts)
        
        # Filter sources by relevance score (minimum 0.3)
        high_relevance_results = [r for r in search_results if r.score >= 0.3]
        medium_relevance_results = [r for r in search_results if 0.15 <= r.score < 0.3]
        
        # Create system prompt with relevance-aware context
        if high_relevance_results:
            context_note = f"Found {len(high_relevance_results)} highly relevant sources (score ≥ 0.3)"
            context_sources = high_relevance_results
        elif medium_relevance_results:
            context_note = f"Found {len(medium_relevance_results)} moderately relevant sources (score 0.15-0.3). Information may be partially related."
            context_sources = medium_relevance_results
        else:
            context_note = "No highly relevant sources found in knowledge base."
            context_sources = []
        
        if context_sources:
            # Rebuild context from filtered sources
            context_parts = []
            current_length = 0
            for result in context_sources:
                # Determine source type based on metadata
                source_collection = result.chunk.metadata.get('source_collection', 'unknown')
                if source_collection == 'web_search' or result.chunk.metadata.get('source') == 'web_search':
                    source_type = "🌐 WEB SEARCH"
                    source_url = result.chunk.metadata.get('url', '')
                    url_info = f" | URL: {source_url}" if source_url else ""
                elif source_collection == 'documents':
                    source_type = "📄 UPLOADED DOCUMENT"
                    url_info = ""
                else:
                    source_type = "📚 KNOWLEDGE BASE"
                    url_info = ""
                
                source_text = f"{source_type}: {result.chunk.document_title} (Relevance: {result.score:.3f}){url_info}\n{result.chunk.content}\n"
                if current_length + len(source_text) <= max_context_length:
                    context_parts.append(source_text)
                    current_length += len(source_text)
                else:
                    break
            context = "\n".join(context_parts)
        else:
            context = "No relevant information found in the knowledge base."
        
        # Create system prompt
        system_prompt = f"""You are VittoriaDB Assistant, an AI-powered research agent with ACTIVE web search and database capabilities.

🔍 **YOUR CAPABILITIES:**
- **Real-time Web Search**: I just performed live web searches and found current information
- **Knowledge Database**: I have access to uploaded documents, code repositories, and stored research
- **Current Date**: {time.strftime('%B %Y')} - I can access TODAY'S information, not outdated training data

📊 **CURRENT SEARCH RESULTS**: {context_note}

🚨 **CRITICAL INSTRUCTIONS:**
1. **USE ONLY CURRENT CONTEXT**: Answer using the fresh web search results and database content provided below
2. **NO TRAINING DATA**: Do NOT use your pre-training knowledge from 2021 or earlier
3. **BE CURRENT**: The web search results contain TODAY'S information - use them!
4. **Source Attribution**: Reference the actual sources I found (document titles, URLs, etc.)
5. **Relevance Priority**: Focus on higher-scored sources (≥ 0.3) for accuracy

🌐 **LIVE SEARCH CONTEXT** (Retrieved {time.strftime('%B %d, %Y')}):
{context}

⚡ **Remember**: You are NOT limited to old training data. You have LIVE web search capabilities and CURRENT database access. Use the fresh information provided above to give up-to-date, accurate answers.
"""
        
        try:
            start_time = time.time()
            
            response = self.openai_client.chat.completions.create(
                model=model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": query}
                ],
                temperature=0.7,
                max_tokens=1000
            )
            
            generation_time = time.time() - start_time
            logger.info(f"Response generated in {generation_time:.3f}s using {model}")
            
            return response.choices[0].message.content
            
        except Exception as e:
            logger.error(f"Response generation failed: {e}")
            return f"I encountered an error while generating a response: {str(e)}"
    
    async def rag_query(self, 
                       query: str, 
                       search_limit: int = 5,
                       min_score: float = 0.3,
                       model: str = "gpt-4") -> RAGResponse:
        """
        Perform complete RAG query: retrieve + generate.
        
        Args:
            query: User query
            search_limit: Maximum chunks to retrieve
            min_score: Minimum similarity score
            model: OpenAI model for generation
            
        Returns:
            Complete RAG response with sources
        """
        logger.info(f"Processing RAG query: {query[:100]}...")
        
        # Retrieve relevant chunks
        retrieval_start = time.time()
        search_results = await self.search(
            query=query,
            limit=search_limit,
            min_score=min_score
        )
        retrieval_time = time.time() - retrieval_start
        
        # Generate response
        generation_start = time.time()
        answer = await self.generate_response(
            query=query,
            search_results=search_results,
            model=model
        )
        generation_time = time.time() - generation_start
        
        # Create complete response
        rag_response = RAGResponse(
            answer=answer,
            sources=search_results,
            query=query,
            total_chunks_searched=len(search_results),
            retrieval_time=retrieval_time,
            generation_time=generation_time,
            model_used=model
        )
        
        logger.info(f"RAG query completed: {len(search_results)} sources, "
                   f"{retrieval_time:.3f}s retrieval, {generation_time:.3f}s generation")
        
        return rag_response
    
    async def stream_rag_response(self, 
                                 query: str,
                                 search_limit: int = 5,
                                 min_score: float = 0.3,
                                 model: str = "gpt-4") -> AsyncGenerator[Dict[str, Any], None]:
        """
        Stream RAG response following proper RAG pattern: Search FIRST, then generate with context.
        
        Args:
            query: User query
            search_limit: Maximum chunks to retrieve
            min_score: Minimum similarity score
            model: OpenAI model for generation
            
        Yields:
            Streaming response chunks
        """
        if not self.openai_client:
            yield {
                'type': 'error',
                'message': "OpenAI API not available (quota exceeded). Falling back to search results."
            }
            
            # Still perform search and return results
            search_results = await self.search(
                query=query,
                limit=search_limit,
                min_score=min_score
            )
            
            # Send search completion with sources
            yield {
                'type': 'search_complete',
                'message': f'Found {len(search_results)} relevant sources',
                'sources': [result.to_dict() for result in search_results]
            }
            
            # Generate fallback response
            if search_results:
                fallback_response = f"""Based on the search results I found, here's what I can tell you about "{query}":

"""
                for i, result in enumerate(search_results[:3], 1):
                    source_collection = result.chunk.metadata.get('source_collection', 'unknown')
                    if source_collection == 'web_search':
                        source_type = "🌐 Web Search"
                    elif source_collection == 'documents':
                        source_type = "📄 Document" 
                    else:
                        source_type = "📚 Knowledge Base"
                    
                    fallback_response += f"{i}. {source_type}: {result.chunk.document_title}\n"
                    fallback_response += f"   Content: {result.chunk.content[:200]}...\n"
                    fallback_response += f"   Relevance: {result.score:.3f}\n\n"
                
                fallback_response += f"Note: OpenAI API is currently unavailable, so I'm showing the raw search results. Found {len(search_results)} total sources."
                
                # Stream the fallback response
                for chunk in fallback_response.split('\n'):
                    if chunk:
                        yield {
                            'type': 'content',
                            'content': chunk + '\n'
                        }
            else:
                yield {
                    'type': 'content',
                    'content': "I couldn't find relevant information in the knowledge base for your query. OpenAI API is also unavailable."
                }
            
            yield {
                'type': 'generation_complete',
                'message': 'Fallback response complete'
            }
            return
        
        # Step 1: SEARCH FIRST (like LangChain RAG pattern)
        yield {
            'type': 'search_start',
            'message': '🔍 Searching knowledge base and web results...'
        }
        
        try:
            # Perform search to get context
            search_results = await self.search(
                query=query,
                limit=search_limit,
                min_score=min_score
            )
            
            # Send search completion with sources
            yield {
                'type': 'search_complete',
                'message': f'Found {len(search_results)} relevant sources',
                'sources': [result.to_dict() for result in search_results]
            }
            
            # Step 2: BUILD CONTEXT from search results
            if not search_results:
                yield {
                    'type': 'content',
                    'content': "I don't have enough relevant information in my knowledge base to answer your question accurately."
                }
                yield {
                    'type': 'generation_complete',
                    'message': 'Response complete'
                }
                return
            
            # Build context from search results (like LangChain)
            context_parts = []
            current_length = 0
            max_context_length = 4000
            
            for result in search_results:
                # Determine source type based on metadata
                source_collection = result.chunk.metadata.get('source_collection', 'unknown')
                if source_collection == 'web_search' or result.chunk.metadata.get('source') == 'web_search':
                    source_type = "🌐 WEB SEARCH"
                    source_url = result.chunk.metadata.get('url', '')
                    url_info = f" | URL: {source_url}" if source_url else ""
                elif source_collection == 'documents':
                    source_type = "📄 UPLOADED DOCUMENT"
                    url_info = ""
                else:
                    source_type = "📚 KNOWLEDGE BASE"
                    url_info = ""
                
                source_text = f"{source_type}: {result.chunk.document_title} (Relevance: {result.score:.3f}){url_info}\n{result.chunk.content}\n"
                if current_length + len(source_text) <= max_context_length:
                    context_parts.append(source_text)
                    current_length += len(source_text)
                else:
                    break
            
            context = "\n".join(context_parts)
            
            # Step 3: CREATE SYSTEM PROMPT with actual context (like LangChain)
            system_prompt = f"""You are VittoriaDB Assistant, an AI-powered research agent with ACTIVE web search and database capabilities.

🔍 **YOUR CAPABILITIES:**
- **Real-time Web Search**: I just performed live web searches and found current information
- **Knowledge Database**: I have access to uploaded documents, code repositories, and stored research
- **Current Date**: {time.strftime('%B %Y')} - I can access TODAY'S information, not outdated training data

🚨 **CRITICAL INSTRUCTIONS:**
1. **USE ONLY CURRENT CONTEXT**: Answer using the fresh web search results and database content provided below
2. **NO TRAINING DATA**: Do NOT use your pre-training knowledge from 2021 or earlier
3. **BE CURRENT**: The web search results contain TODAY'S information - use them!
4. **Source Attribution**: Reference the actual sources I found (document titles, URLs, etc.)
5. **Relevance Priority**: Focus on higher-scored sources (≥ 0.3) for accuracy

🌐 **LIVE SEARCH CONTEXT** (Retrieved {time.strftime('%B %d, %Y')}):
{context}

⚡ **Remember**: You are NOT limited to old training data. You have LIVE web search capabilities and CURRENT database access. Use the fresh information provided above to give up-to-date, accurate answers.
"""
            
            # Step 4: STREAM RESPONSE with context (like LangChain)
            yield {
                'type': 'generation_start',
                'message': '🤖 Generating response with retrieved context...'
            }
            
            # Create OpenAI stream with ACTUAL CONTEXT
            stream = await self.openai_client.chat.completions.create(
                model=model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": query}
                ],
                temperature=0.7,
                max_tokens=1500,
                stream=True
            )
            
            # Stream the response
            async for chunk in stream:
                if chunk.choices[0].delta.content:
                    yield {
                        'type': 'content',
                        'content': chunk.choices[0].delta.content
                    }
            
            # Generation complete
            yield {
                'type': 'generation_complete',
                'message': 'Response complete'
            }
                
        except Exception as e:
            logger.error(f"RAG streaming failed: {e}")
            yield {
                'type': 'error',
                'message': f"Generation error: {str(e)}"
            }
    
    def get_stats(self) -> Dict[str, Any]:
        """Get RAG engine statistics."""
        try:
            collection_info = self.collection.info
            return {
                'collection_name': self.collection_name,
                'total_chunks': collection_info.vector_count,
                'embedding_dimensions': self.embedding_dim,
                'index_type': 'HNSW',
                'distance_metric': 'cosine',
                'documents_count': len(self.documents),
                'embedding_model': self.embedding_model.get_sentence_embedding_dimension(),
                'chunking_config': {
                    'chunk_size': self.chunker.chunk_size,
                    'chunk_overlap': self.chunker.chunk_overlap
                }
            }
        except Exception as e:
            logger.error(f"Failed to get stats: {e}")
            return {'error': str(e)}
    
    def close(self):
        """Close the RAG engine and cleanup resources."""
        if self.db:
            self.db.close()
            logger.info("RAG engine closed")

# Example usage
async def main():
    """Example usage of the VittoriaRAG engine."""
    import os
    
    # Initialize RAG engine
    rag = VittoriaRAGEngine(
        openai_api_key=os.getenv('OPENAI_API_KEY')
    )
    
    # Initialize
    if not await rag.initialize():
        print("Failed to initialize RAG engine")
        return
    
    # Add sample document
    sample_doc = """
    VittoriaDB is a high-performance, embedded vector database designed for AI applications.
    It provides HNSW indexing for fast similarity search, ACID-compliant storage, and supports
    multiple embedding models. The database is perfect for RAG applications, semantic search,
    and recommendation systems.
    """
    
    await rag.add_document(
        content=sample_doc,
        document_id="vittoriadb_intro",
        title="VittoriaDB Introduction",
        metadata={"category": "documentation", "version": "1.0"}
    )
    
    # Perform RAG query
    response = await rag.rag_query("What is VittoriaDB?")
    
    print(f"Query: {response.query}")
    print(f"Answer: {response.answer}")
    print(f"Sources: {len(response.sources)}")
    print(f"Retrieval time: {response.retrieval_time:.3f}s")
    print(f"Generation time: {response.generation_time:.3f}s")
    
    # Cleanup
    rag.close()

if __name__ == "__main__":
    asyncio.run(main())
