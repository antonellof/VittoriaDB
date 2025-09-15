"""
VittoriaDB RAG Web UI - FastAPI Backend
High-performance API with WebSocket support for real-time chat
"""

import os
import asyncio
import logging
import time
import json
import hashlib
from typing import List, Dict, Any, Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException, WebSocket, WebSocketDisconnect, UploadFile, File, BackgroundTasks
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from dotenv import load_dotenv
import structlog

# Import our modules
from models import *
from rag_system import get_rag_system, SearchResult
from rag_engine import VittoriaRAGEngine
from file_processor import get_file_processor
from web_research import get_web_researcher
from github_indexer import get_github_indexer
from notification_system import get_notification_service

# Load environment variables
load_dotenv()

# Configure structured logging
structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.stdlib.PositionalArgumentsFormatter(),
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.StackInfoRenderer(),
        structlog.processors.format_exc_info,
        structlog.processors.UnicodeDecoder(),
        structlog.processors.JSONRenderer()
    ],
    context_class=dict,
    logger_factory=structlog.stdlib.LoggerFactory(),
    wrapper_class=structlog.stdlib.BoundLogger,
    cache_logger_on_first_use=True,
)

logger = structlog.get_logger()

# Global instances
rag_system = None
rag_engine = None
file_processor = None
web_researcher = None
github_indexer = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager"""
    global rag_system, rag_engine, file_processor, web_researcher, github_indexer, notification_service
    
    logger.info("🚀 Starting VittoriaDB RAG API")
    
    # Auto-start VittoriaDB if not running (for Docker)
    try:
        import subprocess
        import time
        
        # Check if VittoriaDB is already running
        try:
            import requests
            requests.get("http://localhost:8080/health", timeout=2)
            logger.info("✅ VittoriaDB already running")
        except:
            logger.info("🔄 Starting VittoriaDB server...")
            # Start VittoriaDB in background
            subprocess.Popen([
                "vittoriadb", "run", 
                "--port", "8080", 
                "--data-dir", "/app/data",
                "--host", "0.0.0.0"
            ])
            
            # Wait for VittoriaDB to start
            for i in range(10):
                try:
                    requests.get("http://localhost:8080/health", timeout=2)
                    logger.info("✅ VittoriaDB started successfully")
                    break
                except:
                    time.sleep(1)
            else:
                logger.warning("⚠️ VittoriaDB may not have started properly")
    except Exception as e:
        logger.warning("⚠️ Could not auto-start VittoriaDB", error=str(e))
    
    # Initialize components
    try:
        rag_system = get_rag_system()
        
        # Initialize new RAG engine
        rag_engine = VittoriaRAGEngine(
            vittoriadb_url="http://localhost:8080",
            collection_name="advanced_rag_kb",
            embedding_model="all-MiniLM-L6-v2",
            openai_api_key=os.getenv('OPENAI_API_KEY')
        )
        await rag_engine.initialize()
        logger.info("✅ Advanced RAG engine initialized")
        
        file_processor = get_file_processor()
        web_researcher = get_web_researcher()
        github_indexer = get_github_indexer()
        notification_service = get_notification_service()
        
        logger.info("✅ All components initialized successfully")
    except Exception as e:
        logger.error("❌ Failed to initialize components", error=str(e))
        raise
    
    yield
    
    # Cleanup
    logger.info("🛑 Shutting down VittoriaDB RAG API")
    if rag_system:
        rag_system.close()
    if rag_engine:
        rag_engine.close()

# Create FastAPI app
app = FastAPI(
    title="VittoriaDB RAG API",
    description="Advanced RAG system with document processing, web research, and code indexing",
    version="1.0.0",
    lifespan=lifespan
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://127.0.0.1:3000"],  # React dev server
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# WebSocket connection manager
class ConnectionManager:
    def __init__(self):
        self.active_connections: List[WebSocket] = []

    async def connect(self, websocket: WebSocket):
        await websocket.accept()
        self.active_connections.append(websocket)
        logger.info("WebSocket connected", connections=len(self.active_connections))

    def disconnect(self, websocket: WebSocket):
        self.active_connections.remove(websocket)
        logger.info("WebSocket disconnected", connections=len(self.active_connections))

    async def send_personal_message(self, message: dict, websocket: WebSocket):
        await websocket.send_text(json.dumps(message))

manager = ConnectionManager()

# Health check endpoint
@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint"""
    try:
        # Check VittoriaDB connection
        vittoriadb_connected = rag_system is not None
        if vittoriadb_connected:
            try:
                rag_system.get_collection_stats()
            except:
                vittoriadb_connected = False
        
        # Check OpenAI configuration
        openai_configured = bool(os.getenv('OPENAI_API_KEY'))
        
        return HealthResponse(
            status="healthy" if vittoriadb_connected else "degraded",
            vittoriadb_connected=vittoriadb_connected,
            openai_configured=openai_configured
        )
    except Exception as e:
        logger.error("Health check failed", error=str(e))
        raise HTTPException(status_code=500, detail="Health check failed")

# System statistics endpoint
@app.get("/rag/stats")
async def get_rag_stats():
    """Get advanced RAG engine statistics"""
    try:
        return rag_engine.get_stats()
    except Exception as e:
        logger.error("Failed to get RAG stats", error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to get stats: {str(e)}")

@app.get("/stats", response_model=SystemStats)
async def get_system_stats():
    """Get system statistics"""
    try:
        stats = rag_system.get_collection_stats()
        
        collections = {}
        total_vectors = 0
        
        for name, collection_stats in stats.items():
            if 'error' not in collection_stats:
                collections[name] = CollectionStats(
                    name=collection_stats['name'],
                    vector_count=collection_stats['vector_count'],
                    dimensions=collection_stats['dimensions'],
                    metric=collection_stats['metric'],
                    index_type=collection_stats.get('index_type', 'unknown'),
                    description=collection_stats.get('description', '')
                )
                total_vectors += collection_stats['vector_count']
        
        system_stats = SystemStats(
            collections=collections,
            total_vectors=total_vectors,
            vittoriadb_status="connected",
            uptime=time.time()  # Simplified uptime
        )
        
        # Notify via WebSocket
        await notification_service.notify_stats_update(system_stats.dict())
        
        return system_stats
    except Exception as e:
        logger.error("Failed to get system stats", error=str(e))
        raise HTTPException(status_code=500, detail="Failed to get system statistics")

# Chat endpoint (non-streaming)
@app.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest):
    """Process chat message and return response"""
    start_time = time.time()
    
    try:
        logger.info("Processing chat request", message=request.message[:100])
        
        # Convert Pydantic models to our internal format
        chat_history = [
            type('ChatMessage', (), {
                'role': msg.role.value,
                'content': msg.content,
                'timestamp': msg.timestamp,
                'sources': msg.sources
            })()
            for msg in request.chat_history
        ]
        
        # Generate response
        response, search_results = await rag_system.chat(
            user_message=request.message,
            chat_history=chat_history,
            search_collections=request.search_collections
        )
        
        # Convert search results
        sources = [
            SearchResult(
                content=result.content,
                metadata=result.metadata,
                score=result.score,
                source=result.source
            )
            for result in search_results
        ]
        
        processing_time = time.time() - start_time
        
        logger.info("Chat request completed", 
                   processing_time=processing_time,
                   sources_count=len(sources))
        
        return ChatResponse(
            message=response,
            sources=sources,
            processing_time=processing_time,
            search_results_count=len(sources)
        )
        
    except Exception as e:
        logger.error("Chat request failed", error=str(e))
        raise HTTPException(status_code=500, detail=f"Chat processing failed: {str(e)}")

# Real-time streaming chat endpoint with web search in thinking area
@app.post("/rag/stream")
async def advanced_rag_stream(request: ChatRequest):
    """Advanced RAG streaming endpoint with web search in reasoning area"""
    async def generate_rag_stream():
        try:
            logger.info("Starting advanced RAG streaming", message=request.message[:100])
            
            # Start thinking/reasoning area with structured steps
            thinking_steps = []
            yield f"data: {json.dumps({'type': 'reasoning_start', 'message': '🧠 Thinking...'})}\n\n"
            await asyncio.sleep(0.001)
            
            # Check if web research is requested
            should_research = request.web_search
            
            web_search_results = []
            
            if should_research:
                # Add web search step
                web_search_step = {
                    'icon': 'Search',
                    'label': 'Searching the web for latest information...',
                    'status': 'active'
                }
                thinking_steps.append(web_search_step)
                
                yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                await asyncio.sleep(0.001)
                
                try:
                    # Stream web research results as they're found
                    async for progress in web_researcher.stream_research_and_store(
                        query=request.message,
                        rag_system=rag_system
                    ):
                        if progress['type'] == 'search_results':
                            # Update step with streaming results
                            web_search_step['status'] = 'active'
                            web_search_step['label'] = f"Found {len(progress['results'])} web results"
                            web_search_step['searchResults'] = [
                                {
                                    'title': result['title'], 
                                    'url': result['url'],
                                    'type': 'web_search',
                                    'favicon': f"https://www.google.com/s2/favicons?domain={result['url']}"
                                }
                                for result in progress['results']
                            ]
                            yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                            await asyncio.sleep(0.001)
                        elif progress['type'] == 'complete':
                            web_search_step['status'] = 'complete'
                            web_search_step['label'] = f"Completed: {progress['total_results']} web results stored"
                            yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                            await asyncio.sleep(0.001)
                        elif progress['type'] == 'error':
                            web_search_step['status'] = 'complete'
                            web_search_step['label'] = f"Web research error: {progress['message']}"
                            yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                            await asyncio.sleep(0.001)
                        
                except Exception as e:
                    web_search_step['status'] = 'complete'
                    web_search_step['label'] = f'Web research error: {str(e)}'
                    yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                    await asyncio.sleep(0.001)
            
            # Add knowledge base search step
            kb_search_step = {
                'icon': 'Database',
                'label': 'Searching knowledge base...',
                'status': 'active'
            }
            thinking_steps.append(kb_search_step)
            
            yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
            await asyncio.sleep(0.001)
            
            # Use the advanced RAG engine for streaming response
            async for chunk in rag_engine.stream_rag_response(
                query=request.message,
                search_limit=request.search_limit,
                min_score=0.3,  # Improved minimum score for better relevance
                model=request.model
            ):
                if chunk['type'] == 'search_complete':
                    # Update knowledge base search step with results
                    kb_search_step['status'] = 'complete'
                    kb_search_step['label'] = f"Found {len(chunk.get('sources', []))} relevant sources"
                    
                    if chunk.get('sources'):
                        kb_search_step['searchResults'] = [
                            {
                                'title': source['chunk']['document_title'],
                                'url': f"Score: {source['score']:.3f}",
                                'type': 'knowledge_base'
                            }
                            for source in chunk['sources']
                        ]
                    
                    yield f"data: {json.dumps({'type': 'reasoning_complete', 'steps': thinking_steps})}\n\n"
                    await asyncio.sleep(0.001)
                
                # Forward other chunk types
                yield f"data: {json.dumps(chunk)}\n\n"
                await asyncio.sleep(0.001)
                
        except Exception as e:
            logger.error("Advanced RAG streaming error", error=str(e))
            yield f"data: {json.dumps({'type': 'error', 'message': str(e)})}\n\n"
    
    return StreamingResponse(
        generate_rag_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
            "Access-Control-Allow-Headers": "Content-Type",
            "X-Accel-Buffering": "no",
        }
    )

@app.post("/chat/stream")
async def chat_stream(request: ChatRequest):
    """Process chat message with immediate streaming response"""
    async def generate_stream():
        try:
            logger.info("Starting real-time streaming chat", message=request.message[:100])
            
            # Send immediate status and start search progress streaming
            yield f"data: {json.dumps({'type': 'status', 'message': 'Searching knowledge base...'})}\n\n"
            
            # Check if OpenAI is configured
            if not rag_system.openai_client:
                yield f"data: {json.dumps({'type': 'error', 'message': 'OpenAI not configured'})}\n\n"
                return
            
            # Perform knowledge base search first
            search_results = []
            context_text = ""
            
            if request.search_collections and len(request.search_collections) > 0:
                try:
                    yield f"data: {json.dumps({'type': 'search_progress', 'message': f'🔍 Searching {len(request.search_collections)} collections...'})}\n\n"
                    
                    search_results = await rag_system.search_knowledge_base(
                        query=request.message,
                        collections=request.search_collections,
                        limit=request.search_limit,
                        min_score=0.3  # Increased minimum score for better relevance
                    )
                    
                    if search_results:
                        # Filter results by relevance score
                        high_relevance = [r for r in search_results if r.score >= 0.5]
                        medium_relevance = [r for r in search_results if 0.3 <= r.score < 0.5]
                        
                        yield f"data: {json.dumps({'type': 'search_progress', 'message': f'✅ Found {len(search_results)} relevant documents (High: {len(high_relevance)}, Medium: {len(medium_relevance)})'})}\n\n"
                        
                        # Build enhanced context from search results
                        context_parts = []
                        for i, result in enumerate(search_results[:5]):  # Use top 5 results
                            # Determine source type and add metadata
                            source_collection = result.metadata.get('source_collection', result.source)
                            if source_collection == 'web_research' or result.metadata.get('source') == 'web_search':
                                source_type = "🌐 WEB SEARCH"
                                source_url = result.metadata.get('url', '')
                                url_info = f" | URL: {source_url}" if source_url else ""
                            elif source_collection == 'documents':
                                source_type = "📄 UPLOADED DOCUMENT"
                                url_info = ""
                            else:
                                source_type = "📚 KNOWLEDGE BASE"
                                url_info = ""
                            
                            # Enhanced context with relevance scoring
                            relevance_indicator = "🔥 HIGH" if result.score >= 0.5 else "📊 MEDIUM" if result.score >= 0.3 else "📉 LOW"
                            
                            context_parts.append(f"""
{source_type}: {result.metadata.get('title', 'Unknown Document')} 
Relevance: {relevance_indicator} ({result.score:.3f}){url_info}
Content: {result.content}
---""")
                        
                        context_text = "\n".join(context_parts)
                    else:
                        yield f"data: {json.dumps({'type': 'search_progress', 'message': '⚠️ No highly relevant documents found (score < 0.3)'})}\n\n"
                        
                except Exception as e:
                    yield f"data: {json.dumps({'type': 'search_progress', 'message': f'❌ Search error: {str(e)}'})}\n\n"
            
            # Create system prompt with context
            if context_text:
                system_prompt = f"""You are VittoriaDB Assistant, an AI-powered research agent with ACTIVE web search and database capabilities.

🔍 **YOUR CAPABILITIES:**
- **Real-time Web Search**: I just performed live web searches and found current information
- **Knowledge Database**: I have access to uploaded documents, code repositories, and stored research
- **Current Date**: {time.strftime('%B %Y')} - I can access TODAY'S information, not outdated training data

🚨 **CRITICAL INSTRUCTIONS:**
1. **USE ONLY PROVIDED CONTEXT**: Answer using ONLY the search results and database content provided below
2. **NO TRAINING DATA**: Do NOT use your pre-training knowledge - use ONLY the context provided
3. **BE CURRENT**: The web search results contain TODAY'S information - prioritize them!
4. **Source Attribution**: Always reference the actual sources (document titles, URLs, relevance scores)
5. **Relevance Priority**: Focus on 🔥 HIGH relevance sources (≥ 0.5) for accuracy

🌐 **RETRIEVED CONTEXT** (Retrieved {time.strftime('%B %d, %Y')}):
{context_text}

⚡ **Remember**: You have LIVE web search capabilities and CURRENT database access. Use the fresh information provided above to give up-to-date, accurate answers. If the context doesn't contain sufficient information to answer the question, clearly state what information is missing and what was found instead."""
            else:
                system_prompt = f"""You are VittoriaDB Assistant, an AI-powered research agent.

❌ **NO RELEVANT CONTEXT FOUND**: No documents in the knowledge base matched the user's query with sufficient relevance (score ≥ 0.3).

🔍 **WHAT TO DO:**
1. Clearly explain that no relevant information was found in the knowledge base
2. Suggest the user try:
   - Different search terms or keywords
   - Adding relevant documents to the knowledge base
   - Using the web research feature to find current information
3. Provide general guidance if appropriate, but make it clear it's not from the knowledge base

⚠️ **IMPORTANT**: Do not make up information or use outdated training data. Be honest about the lack of relevant context."""
            
            # Start OpenAI streaming with context
            yield f"data: {json.dumps({'type': 'search_progress', 'message': '🤖 Generating response...'})}\n\n"
            
            stream_response = rag_system.openai_client.chat.completions.create(
                model=request.model,
                messages=[
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": request.message}
                ],
                temperature=0.7,
                max_tokens=1500,
                stream=True
            )
            
            # Stream OpenAI response immediately while search happens in background
            response_chunks = []
            for chunk in stream_response:
                if chunk.choices[0].delta.content:
                    content = chunk.choices[0].delta.content
                    response_chunks.append(content)
                    yield f"data: {json.dumps({'type': 'content', 'content': content})}\n\n"
            
            # Note: Search results would be processed here in a full implementation
            # For now, we're prioritizing immediate streaming response
            
            # Send completion
            yield f"data: {json.dumps({'type': 'done'})}\n\n"
            
        except Exception as e:
            logger.error("Streaming chat error", error=str(e))
            yield f"data: {json.dumps({'type': 'error', 'message': str(e)})}\n\n"
    
    return StreamingResponse(
        generate_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
            "Access-Control-Allow-Headers": "Content-Type",
            "X-Accel-Buffering": "no",  # Disable nginx buffering
        }
    )

# WebSocket endpoint for streaming chat
@app.websocket("/ws/chat")
async def websocket_chat(websocket: WebSocket):
    """WebSocket endpoint for real-time streaming chat"""
    await manager.connect(websocket)
    
    try:
        while True:
            # Receive message from client
            data = await websocket.receive_text()
            request_data = json.loads(data)
            
            logger.info("WebSocket chat request", message=request_data.get('message', '')[:100])
            
            try:
                # Parse request
                request = ChatRequest(**request_data)
                
                # Send typing indicator
                await manager.send_personal_message({
                    "type": "typing",
                    "content": "Assistant is thinking..."
                }, websocket)
                
                # Convert chat history
                chat_history = [
                    type('ChatMessage', (), {
                        'role': msg.role.value,
                        'content': msg.content,
                        'timestamp': msg.timestamp,
                        'sources': msg.sources
                    })()
                    for msg in request.chat_history
                ]
                
                # Check if web research is needed
                research_keywords = ['research', 'search web', 'look up', 'find information about']
                should_research = any(keyword in request.message.lower() for keyword in research_keywords)
                
                if should_research and 'web' in request.message.lower():
                    await manager.send_personal_message({
                        "type": "status",
                        "content": "🔍 Researching on the web..."
                    }, websocket)
                    
                    # Perform web research
                    research_result = await web_researcher.research_and_store(
                        query=request.message,
                        rag_system=rag_system
                    )
                    
                    if research_result['success']:
                        await manager.send_personal_message({
                            "type": "status",
                            "content": f"✅ Found {research_result['results_count']} web results"
                        }, websocket)
                
                # Generate response
                await manager.send_personal_message({
                    "type": "status",
                    "content": "🧠 Generating response..."
                }, websocket)
                
                start_time = time.time()
                response, search_results = await rag_system.chat(
                    user_message=request.message,
                    chat_history=chat_history,
                    search_collections=request.search_collections
                )
                
                # Convert search results
                sources = [
                    SearchResult(
                        content=result.content,
                        metadata=result.metadata,
                        score=result.score,
                        source=result.source
                    )
                    for result in search_results
                ]
                
                processing_time = time.time() - start_time
                
                # Send response in chunks for streaming effect
                words = response.split()
                chunk_size = 3  # Words per chunk
                
                for i in range(0, len(words), chunk_size):
                    chunk = ' '.join(words[i:i + chunk_size])
                    await manager.send_personal_message({
                        "type": "content",
                        "content": chunk + (' ' if i + chunk_size < len(words) else '')
                    }, websocket)
                    await asyncio.sleep(0.05)  # Small delay for streaming effect
                
                # Send sources
                await manager.send_personal_message({
                    "type": "sources",
                    "sources": [source.dict() for source in sources]
                }, websocket)
                
                # Send completion
                await manager.send_personal_message({
                    "type": "done",
                    "processing_time": processing_time
                }, websocket)
                
                logger.info("WebSocket chat completed", processing_time=processing_time)
                
            except Exception as e:
                logger.error("WebSocket chat error", error=str(e))
                await manager.send_personal_message({
                    "type": "error",
                    "error": f"Error: {str(e)}"
                }, websocket)
                
    except WebSocketDisconnect:
        manager.disconnect(websocket)

# File upload endpoint
@app.post("/rag/document")
async def add_document_to_rag(request: Dict[str, Any]):
    """Add a document to the advanced RAG engine"""
    try:
        content = request.get('content', '')
        document_id = request.get('document_id', f"doc_{int(time.time())}")
        title = request.get('title', 'Untitled Document')
        metadata = request.get('metadata', {})
        
        logger.info("Adding document to advanced RAG", title=title)
        
        chunks_created = await rag_engine.add_document(
            content=content,
            document_id=document_id,
            title=title,
            metadata=metadata
        )
        
        return {
            "success": True,
            "filename": title,
            "document_id": document_id,
            "chunks_created": chunks_created,
            "processing_time": 0.0,
            "message": f"Document '{title}' added to advanced RAG with {chunks_created} chunks",
            "metadata": metadata
        }
        
    except Exception as e:
        logger.error("Failed to add document to RAG", error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to add document: {str(e)}")

async def process_file_background(file_content: bytes, filename: str, document_id: str):
    """Background task to process file and store embeddings"""
    try:
        logger.info("Starting background processing", filename=filename, document_id=document_id)
        
        # Notify processing start
        await notification_service.notify_processing_start(
            document_id=document_id,
            filename=filename,
            file_size=len(file_content)
        )
        
        # Process file
        await notification_service.notify_processing_progress(
            document_id=document_id,
            progress=25,
            message="Processing file content..."
        )
        
        processed_doc = await file_processor.process_uploaded_file(
            file_content=file_content,
            filename=filename,
            additional_metadata={'uploaded_via': 'api', 'document_id': document_id}
        )
        
        await notification_service.notify_processing_progress(
            document_id=document_id,
            progress=50,
            message="Generating embeddings..."
        )
        
        # Store chunks in VittoriaDB (this is the slow part)
        chunks_stored = 0
        total_chunks = len(processed_doc.chunks)
        
        for i, chunk in enumerate(processed_doc.chunks):
            await rag_system.add_document(
                content=chunk.content,
                metadata=chunk.metadata,
                collection_name='documents'
            )
            chunks_stored += 1
            
            # Update progress
            progress = 50 + int((i + 1) / total_chunks * 45)  # 50-95%
            await notification_service.notify_processing_progress(
                document_id=document_id,
                progress=progress,
                message=f"Storing chunk {i + 1}/{total_chunks}..."
            )
        
        # Notify completion
        processing_time = time.time() - notification_service.processing_status[document_id]['start_time']
        await notification_service.notify_processing_complete(
            document_id=document_id,
            chunks_created=chunks_stored,
            processing_time=processing_time
        )
        
        # Update stats
        stats = await rag_system.get_stats()
        await notification_service.notify_stats_update(stats)
        
        logger.info("Background processing completed", 
                   filename=filename,
                   document_id=document_id,
                   chunks=chunks_stored)
        
    except Exception as e:
        logger.error("Background processing failed", 
                    filename=filename, 
                    document_id=document_id, 
                    error=str(e))
        
        # Notify error
        await notification_service.notify_processing_error(
            document_id=document_id,
            error=str(e)
        )

@app.post("/upload", response_model=FileUploadResponse)
async def upload_file(background_tasks: BackgroundTasks, file: UploadFile = File(...)):
    """Upload file and process in background"""
    upload_start = time.time()
    
    try:
        logger.info("Receiving file upload", filename=file.filename, size=file.size)
        
        # Read file content (fast)
        file_content = await file.read()
        upload_time = time.time() - upload_start
        
        # Generate document ID
        document_id = f"doc_{int(time.time())}_{hash(file.filename) % 10000}"
        
        # Quick file validation (just check file type and create basic metadata)
        file_ext = os.path.splitext(file.filename)[1].lower()
        supported_types = {'.pdf', '.docx', '.doc', '.txt', '.md', '.html', '.htm'}
        
        if file_ext not in supported_types:
            raise HTTPException(status_code=400, detail=f"Unsupported file type: {file_ext}")
        
        # Create basic metadata for immediate response
        content_hash = hashlib.md5(file_content).hexdigest()[:8]
        basic_metadata = {
            'filename': file.filename,
            'file_type': file_ext,
            'file_size': len(file_content),
            'content_hash': content_hash,
            'uploaded_via': 'api',
            'document_id': document_id,
            'upload_timestamp': time.time()
        }
        
        # Add background task for embedding processing
        background_tasks.add_task(
            process_file_background,
            file_content,
            file.filename,
            document_id
        )
        
        logger.info("File upload received, processing in background", 
                   filename=file.filename,
                   document_id=document_id,
                   upload_time=upload_time)
        
        return FileUploadResponse(
            success=True,
            message=f"File '{file.filename}' uploaded successfully. Processing embeddings in background...",
            filename=file.filename,
            document_id=document_id,
            chunks_created=0,  # Will be processed in background
            processing_time=upload_time,
            metadata={
                **basic_metadata,
                'status': 'processing',
                'background_processing': True
            }
        )
        
    except Exception as e:
        logger.error("File upload failed", filename=file.filename, error=str(e))
        raise HTTPException(status_code=500, detail=f"File upload failed: {str(e)}")

# Web research endpoint
@app.post("/research", response_model=WebResearchResponse)
async def web_research(request: WebResearchRequest, background_tasks: BackgroundTasks):
    """Perform web research and store results"""
    try:
        logger.info("Starting web research", query=request.query)
        
        result = await web_researcher.research_and_store(
            query=request.query,
            rag_system=rag_system,
            search_engine=request.search_engine
        )
        
        logger.info("Web research completed", 
                   query=request.query,
                   results_count=result.get('results_count', 0))
        
        return WebResearchResponse(**result)
        
    except Exception as e:
        logger.error("Web research failed", query=request.query, error=str(e))
        raise HTTPException(status_code=500, detail=f"Web research failed: {str(e)}")

# GitHub indexing endpoint
@app.post("/github/index", response_model=GitHubIndexResponse)
async def index_github_repo(request: GitHubIndexRequest):
    """Index GitHub repository"""
    try:
        logger.info("Starting GitHub indexing", repo_url=request.repository_url)
        
        result = await github_indexer.index_and_store(
            repo_url=request.repository_url,
            rag_system=rag_system
        )
        
        logger.info("GitHub indexing completed", 
                   repo_url=request.repository_url,
                   success=result.get('success', False))
        
        return GitHubIndexResponse(**result)
        
    except Exception as e:
        logger.error("GitHub indexing failed", 
                    repo_url=request.repository_url, 
                    error=str(e))
        raise HTTPException(status_code=500, detail=f"GitHub indexing failed: {str(e)}")

# Search endpoint
@app.post("/search", response_model=SearchResponse)
async def search_knowledge_base(request: SearchRequest):
    """Search across knowledge base"""
    start_time = time.time()
    
    try:
        logger.info("Knowledge base search", query=request.query)
        
        search_results = await rag_system.search_knowledge_base(
            query=request.query,
            collections=request.collections,
            limit=request.limit,
            min_score=request.min_score
        )
        
        # Convert search results to Pydantic models
        from models import SearchResult as PydanticSearchResult
        results = [
            PydanticSearchResult(
                content=result.content,
                metadata=result.metadata,
                score=result.score,
                source=result.source
            )
            for result in search_results
        ]
        
        processing_time = time.time() - start_time
        
        logger.info("Search completed", 
                   query=request.query,
                   results_count=len(results),
                   processing_time=processing_time)
        
        return SearchResponse(
            query=request.query,
            results=results,
            total_results=len(results),
            processing_time=processing_time
        )
        
    except Exception as e:
        logger.error("Search failed", query=request.query, error=str(e))
        raise HTTPException(status_code=500, detail=f"Search failed: {str(e)}")

# Configuration endpoints
@app.get("/config", response_model=ConfigResponse)
async def get_config():
    """Get current configuration"""
    return ConfigResponse(
        openai_configured=bool(os.getenv('OPENAI_API_KEY')),
        github_configured=bool(os.getenv('GITHUB_TOKEN')),
        current_model="gpt-3.5-turbo",  # Default model
        search_limit=5,
        vittoriadb_url=os.getenv('VITTORIADB_URL', 'http://localhost:8080')
    )

# WebSocket endpoints
@app.websocket("/ws/notifications")
async def websocket_notifications(websocket: WebSocket):
    """WebSocket endpoint for real-time notifications"""
    try:
        await notification_service.connect_websocket(websocket)
        
        # Keep connection alive and handle client messages
        while True:
            try:
                # Receive any client messages (like ping/pong)
                data = await websocket.receive_text()
                message_data = json.loads(data)
                
                # Handle client requests
                if message_data.get("type") == "ping":
                    await websocket.send_text(json.dumps({
                        "type": "pong",
                        "timestamp": time.time()
                    }))
                elif message_data.get("type") == "get_processing_status":
                    # Send current processing status
                    status = notification_service.get_all_processing_status()
                    await websocket.send_text(json.dumps({
                        "type": "processing_status",
                        "data": status,
                        "timestamp": time.time()
                    }))
                elif message_data.get("type") == "get_stats":
                    # Send current system stats
                    try:
                        stats = await get_system_stats()
                        await websocket.send_text(json.dumps({
                            "type": "stats_update",
                            "data": stats,
                            "timestamp": time.time()
                        }))
                    except Exception as e:
                        logger.error(f"Failed to get stats for WebSocket: {e}")
                elif message_data.get("type") == "get_health":
                    # Send current health status
                    try:
                        health = await get_health_status()
                        await websocket.send_text(json.dumps({
                            "type": "health_update", 
                            "data": health,
                            "timestamp": time.time()
                        }))
                    except Exception as e:
                        logger.error(f"Failed to get health for WebSocket: {e}")
                    
            except WebSocketDisconnect:
                break
            except Exception as e:
                logger.warning(f"WebSocket message handling error: {e}")
                break
                
    except Exception as e:
        logger.error(f"WebSocket connection error: {e}")
    finally:
        notification_service.disconnect_websocket(websocket)

if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv('PORT', 8501))
    
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=True,
        log_level="info"
    )
