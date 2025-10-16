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
import re
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
from file_processor import get_file_processor
from web_research import get_web_researcher
from web_research_crawl4ai import AdvancedWebResearcher, research_with_crawl4ai
from github_indexer import get_github_indexer
from notification_system import get_notification_service

# Load environment variables
load_dotenv()

# Fix tokenizers multiprocessing warning
os.environ["TOKENIZERS_PARALLELISM"] = "false"

# Overview query keywords (used across multiple endpoints)
OVERVIEW_KEYWORDS = [
    'what documents', 'list documents', 'show documents', 'documents do I have', 
    'knowledge base', 'what files', 'show files', 'what content', 'overview',
    'what information', 'what data', 'show me everything', 'all documents',
    'contents of', 'what\'s in', 'inventory', 'catalog'
]

def extract_suggestions_from_response(response_text: str) -> List[str]:
    """Extract suggestions from AI response text"""
    suggestions = []
    
    # Look for suggestions in code blocks
    suggestions_pattern = r'```suggestions\n(.*?)\n```'
    match = re.search(suggestions_pattern, response_text, re.DOTALL)
    
    if match:
        suggestions_text = match.group(1)
        # Split by lines and clean up
        for line in suggestions_text.split('\n'):
            line = line.strip()
            if line and not line.startswith('#') and '?' in line:
                suggestions.append(line)
    
    # Fallback: look for questions ending with ?
    if not suggestions:
        # Look for lines that end with question marks
        question_pattern = r'^[•\-\*]?\s*(.+\?)\s*$'
        for line in response_text.split('\n'):
            match = re.match(question_pattern, line.strip())
            if match:
                question = match.group(1).strip()
                if len(question) > 10 and len(question) < 100:  # Reasonable length
                    suggestions.append(question)
    
    # Limit to 6 suggestions and ensure they're unique
    return list(dict.fromkeys(suggestions))[:6]

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
file_processor = None
web_researcher = None
advanced_web_researcher = None
github_indexer = None

@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager"""
    global rag_system, rag_engine, file_processor, web_researcher, advanced_web_researcher, github_indexer, notification_service
    
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
        file_processor = get_file_processor()
        web_researcher = get_web_researcher()
        advanced_web_researcher = AdvancedWebResearcher(max_results=3, max_content_length=2000)
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

def estimate_tokens(text: str) -> int:
    """Rough token estimation (1 token ≈ 4 characters for English)"""
    return len(text) // 4

def truncate_for_model(context_text: str, system_prompt: str, user_message: str, model: str) -> str:
    """Intelligently truncate context to fit model limits"""
    # Model context limits
    model_limits = {
        "gpt-4": 8192,
        "gpt-4-turbo": 128000,
        "gpt-4o": 128000,
        "gpt-4o-mini": 128000
    }
    
    max_tokens = model_limits.get(model, 8192)
    
    # Reserve tokens for system prompt, user message, and response
    system_tokens = estimate_tokens(system_prompt)
    user_tokens = estimate_tokens(user_message)
    response_tokens = 1500  # Reserve for response
    
    available_tokens = max_tokens - system_tokens - user_tokens - response_tokens - 200  # Safety buffer
    
    if available_tokens <= 0:
        return "[Context too large for model]"
    
    max_context_chars = available_tokens * 4
    
    if len(context_text) > max_context_chars:
        # Smart truncation for GitHub code content
        if "Repository:" in context_text and ("README" in context_text or "Code:" in context_text):
            # For GitHub code, prioritize documentation over code
            lines = context_text.split('\n')
            important_content = []
            current_length = 0
            
            # Priority order: Repository info, README sections, then code
            for line in lines:
                line_length = len(line) + 1  # +1 for newline
                if current_length + line_length > max_context_chars:
                    break
                    
                # Keep high-priority content
                if any(keyword in line for keyword in [
                    'Repository:', 'RAGFlow', '## What is', '## Key Features', 
                    '## Get Started', 'Description:', 'README.md'
                ]):
                    important_content.append(line)
                    current_length += line_length
                elif line.strip().startswith('#') and len(line) < 100:  # Headers
                    important_content.append(line)
                    current_length += line_length
                elif 'Code:' in line and current_length < max_context_chars * 0.7:  # Some code if space
                    important_content.append(line)
                    current_length += line_length
            
            truncated = '\n'.join(important_content)
            return truncated + "\n\n[GitHub content truncated - showing key documentation]"
        else:
            # Standard truncation for non-GitHub content
            truncated = context_text[:max_context_chars]
            # Try to end at a complete section
            last_section = truncated.rfind("----")
            if last_section > max_context_chars * 0.8:  # If we can keep 80% of content
                truncated = truncated[:last_section + 4]
            
            return truncated + "\n\n[Content truncated to fit model limits]"
    
    return context_text

async def get_health_status():
    """Get health status for internal use (JSON serializable)"""
    try:
        # Check VittoriaDB connection
        vittoriadb_connected = rag_system is not None
        if vittoriadb_connected:
            try:
                rag_system.get_collection_stats()
            except:
                vittoriadb_connected = False
        
        # Check OpenAI configuration (check if key exists and is not empty/placeholder)
        openai_key = os.getenv('OPENAI_API_KEY', '')
        openai_configured = bool(openai_key and openai_key.strip() and 
                                openai_key != 'your-openai-api-key-here' and
                                len(openai_key) > 10)
        
        return {
            "status": "healthy" if vittoriadb_connected else "degraded",
            "vittoriadb_connected": vittoriadb_connected,
            "openai_configured": openai_configured,
            "timestamp": time.time()
        }
    except Exception as e:
        logger.error("Health status check failed", error=str(e))
        return {
            "status": "error",
            "error": str(e),
            "timestamp": time.time()
        }

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
        
        # Check OpenAI configuration (check if key exists and is not empty/placeholder)
        openai_key = os.getenv('OPENAI_API_KEY', '')
        openai_configured = bool(openai_key and openai_key.strip() and 
                                openai_key != 'your-openai-api-key-here' and
                                len(openai_key) > 10)
        
        return HealthResponse(
            status="healthy" if vittoriadb_connected else "degraded",
            vittoriadb_connected=vittoriadb_connected,
            openai_configured=openai_configured
        )
    except Exception as e:
        logger.error("Health check failed", error=str(e))
        raise HTTPException(status_code=500, detail="Health check failed")

@app.get("/collections/stats")
async def get_collection_stats():
    """Get detailed statistics for all collections"""
    try:
        stats = rag_system.get_collection_stats()
        
        # Add more detailed information
        detailed_stats = {
            "total_collections": len(stats),
            "collections": {},
            "summary": {
                "total_documents": 0,
                "by_type": {}
            }
        }
        
        for collection_name, collection_info in stats.items():
            detailed_stats["collections"][collection_name] = {
                "document_count": collection_info.get("document_count", 0),
                "description": rag_system.collection_configs.get(collection_name, {}).get("description", ""),
                "dimensions": rag_system.collection_configs.get(collection_name, {}).get("dimensions", 0)
            }
            
            # Add to summary
            doc_count = collection_info.get("document_count", 0)
            detailed_stats["summary"]["total_documents"] += doc_count
            detailed_stats["summary"]["by_type"][collection_name] = doc_count
        
        return detailed_stats
        
    except Exception as e:
        logger.error("Failed to get collection stats", error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to get collection stats: {str(e)}")

# System statistics endpoint
# Removed /rag/stats endpoint - stats are available via /stats endpoint

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
        
        # Convert search results to Pydantic models
        sources = []
        for result in search_results:
            sources.append(
                SearchResult(
                    content=str(result.content),
                    metadata=dict(result.metadata) if result.metadata else {},
                    score=float(result.score),
                    source=str(result.source)
                )
            )
        
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
            web_search_content = []  # Store actual content for immediate use
            
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
                    # Stream web research results using Crawl4AI with duplicate detection
                    async for progress in advanced_web_researcher.stream_research_and_store(
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
                                    'status': 'pending',
                                    'favicon': f"https://www.google.com/s2/favicons?domain={result['url']}"
                                }
                                for result in progress['results']
                            ]
                            yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                            await asyncio.sleep(0.001)
                        elif progress['type'] == 'progress':
                            # Update specific URL status to "reading"
                            if web_search_step.get('searchResults'):
                                for result in web_search_step['searchResults']:
                                    if result['url'] == progress.get('url'):
                                        result['status'] = 'reading'
                                        result['message'] = progress['message']
                                        break
                            web_search_step['label'] = progress['message']
                            yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                            await asyncio.sleep(0.001)
                        elif progress['type'] == 'stored':
                            # Update specific URL status to "complete" and collect content
                            if web_search_step.get('searchResults'):
                                for result in web_search_step['searchResults']:
                                    if result['url'] == progress.get('url'):
                                        result['status'] = 'complete'
                                        result['message'] = progress['message']
                                        result['features'] = progress.get('features', {})
                                        break
                            
                            # Store the web search content for immediate LLM use
                            if progress.get('content'):
                                web_search_content.append({
                                    'title': progress.get('title', 'Web Result'),
                                    'url': progress.get('url', ''),
                                    'content': progress.get('content', ''),
                                    'features': progress.get('features', {}),
                                    'from_cache': progress.get('from_cache', False)
                                })
                            
                            # Update frontend with cache status
                            if web_search_step.get('searchResults'):
                                for result in web_search_step['searchResults']:
                                    if result['url'] == progress.get('url'):
                                        result['from_cache'] = progress.get('from_cache', False)
                                        break
                            
                            yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                            await asyncio.sleep(0.001)
                        elif progress['type'] == 'warning':
                            # Update specific URL status to "error"
                            if web_search_step.get('searchResults'):
                                for result in web_search_step['searchResults']:
                                    if result['url'] == progress.get('url'):
                                        result['status'] = 'error'
                                        result['message'] = progress['message']
                                        break
                            yield f"data: {json.dumps({'type': 'reasoning_update', 'steps': thinking_steps})}\n\n"
                            await asyncio.sleep(0.001)
                        elif progress['type'] == 'complete':
                            web_search_step['status'] = 'complete'
                            web_search_step['label'] = f"Completed: {progress['total_results']} pages read"
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
            
            # Search existing knowledge base
            try:
                # Use model with larger context window (define early)
                model_to_use = "gpt-4o" if request.model == "gpt-4" else request.model
                
                # Auto-adjust parameters for knowledge base overview queries
                search_limit = request.search_limit
                min_score = 0.3
                
                overview_keywords = OVERVIEW_KEYWORDS
                
                if any(keyword in request.message.lower() for keyword in overview_keywords):
                    search_limit = min(10, search_limit * 10)  # Increase limit for overview queries
                    min_score = max(0.1, min_score - 0.2)  # Lower threshold
                    logger.info(f"🔍 Knowledge base overview query detected in advanced RAG - using limit={search_limit}, min_score={min_score}")
                
                # Use rag_system.search_knowledge_base to respect search_collections parameter
                db_search_results_raw = await rag_system.search_knowledge_base(
                    query=request.message,
                    collections=request.search_collections,
                    limit=search_limit,
                    min_score=min_score
                )
                
                # Use search results directly (SearchResult objects)
                db_search_results = db_search_results_raw
                
                # Update knowledge base search step
                kb_search_step['status'] = 'complete'
                kb_search_step['label'] = f"Found {len(db_search_results)} relevant sources"
                
                if db_search_results:
                    kb_search_step['searchResults'] = [
                        {
                            'title': result.metadata.get('title', result.metadata.get('document_title', 'Unknown')),
                            'url': f"Score: {result.score:.3f}",
                            'type': 'knowledge_base'
                        }
                        for result in db_search_results
                    ]
                
                yield f"data: {json.dumps({'type': 'reasoning_complete', 'steps': thinking_steps})}\n\n"
                await asyncio.sleep(0.001)
                
                # Combine web search content with database results
                combined_context = []
                
                # Add fresh web search results first (highest priority) - truncated for context
                for i, web_result in enumerate(web_search_content[:2]):  # Limit to 2 fresh results
                    combined_context.append(f"""
🌐 WEB SEARCH (FRESH): {web_result['title']}
URL: {web_result['url']}
Content: {web_result['content'][:800]}...
Features: {web_result.get('features', {})}
----""")
                
                # Add existing database results - more for overview queries
                max_db_results = min(10, len(db_search_results)) if any(keyword in request.message.lower() for keyword in overview_keywords) else 2
                for result in db_search_results[:max_db_results]:
                    source_collection = result.metadata.get('source_collection', result.source)
                    if source_collection == 'web_research':
                        source_type = "🌐 WEB SEARCH (STORED)"
                        source_url = result.metadata.get('url', '')
                        url_info = f" | URL: {source_url}" if source_url else ""
                    elif source_collection == 'documents':
                        source_type = "📄 UPLOADED DOCUMENT"
                        url_info = ""
                    else:
                        source_type = "📚 KNOWLEDGE BASE"
                        url_info = ""
                    
                    doc_title = result.metadata.get('title', result.metadata.get('document_title', 'Unknown'))
                    combined_context.append(f"""
{source_type}: {doc_title}
Relevance: {result.score:.3f}{url_info}
Content: {result.content[:600]}...
----""")
                
                # Create enhanced system prompt with combined context
                context_text = "\n".join(combined_context)
                
                # Smart truncation based on model limits
                base_system_prompt = """You are VittoriaDB Assistant with LIVE web search and database access.

🔍 **CURRENT CAPABILITIES:**
- **Fresh Web Search**: I just performed live web searches with current information
- **Knowledge Database**: I have access to stored documents and previous research
- **Current Date**: {time.strftime('%B %d, %Y')} - Today's information available!

🚨 **CRITICAL INSTRUCTIONS:**
1. **PRIORITIZE FRESH WEB RESULTS**: Use the "🌐 WEB SEARCH (FRESH)" results first - they contain TODAY'S information
2. **USE ALL PROVIDED CONTEXT**: Combine fresh web results with stored knowledge
3. **NO TRAINING DATA**: Answer using ONLY the context provided below
4. **BE COMPREHENSIVE**: Synthesize information from multiple sources when available
5. **Source Attribution**: Reference the sources and their relevance scores

🌐 **RETRIEVED CONTEXT** (Retrieved {time.strftime('%B %d, %Y at %H:%M')}):
{context_placeholder}

⚡ **Remember**: You have FRESH web search results AND stored knowledge. Use both to provide comprehensive, up-to-date answers!"""
                
                context_text = truncate_for_model(context_text, base_system_prompt, request.message, model_to_use)
                
                if context_text:
                    system_prompt = f"""You are VittoriaDB Assistant with LIVE web search and database access.

🔍 **CURRENT CAPABILITIES:**
- **Fresh Web Search**: I just performed live web searches with current information
- **Knowledge Database**: I have access to stored documents and previous research
- **Current Date**: {time.strftime('%B %d, %Y')} - Today's information available!

🚨 **CRITICAL INSTRUCTIONS:**
1. **PRIORITIZE FRESH WEB RESULTS**: Use the "🌐 WEB SEARCH (FRESH)" results first - they contain TODAY'S information
2. **USE ALL PROVIDED CONTEXT**: Combine fresh web results with stored knowledge
3. **NO TRAINING DATA**: Answer using ONLY the context provided below
4. **BE COMPREHENSIVE**: Synthesize information from multiple sources when available
5. **Source Attribution**: Reference the sources and their relevance scores

🌐 **RETRIEVED CONTEXT** (Retrieved {time.strftime('%B %d, %Y at %H:%M')}):
{context_text}

⚡ **Remember**: You have FRESH web search results AND stored knowledge. Use both to provide comprehensive, up-to-date answers!"""
                else:
                    system_prompt = """You are VittoriaDB Assistant. No relevant information was found in either fresh web search results or the knowledge database for this query. Please suggest alternative search terms or ask the user to be more specific."""
                
                # Stream LLM response with combined context
                yield f"data: {json.dumps({'type': 'llm_start', 'message': 'Generating response with fresh web data...'})}\n\n"
                
                if not rag_system.openai_client:
                    yield f"data: {json.dumps({'type': 'error', 'message': 'OpenAI not configured'})}\n\n"
                    return
                
                # Build conversation history as prompt
                conversation_history = ""
                if request.chat_history:
                    for msg in request.chat_history[-10:]:
                        role = msg.role.value if hasattr(msg.role, 'value') else str(msg.role)
                        conversation_history += f"{role.capitalize()}: {msg.content}\n"
                
                # Build the full prompt with user's current message
                user_prompt = f"{conversation_history}User: {request.message}\nAssistant:"
                
                # Stream the response using Datapizza AI streaming
                # Build messages array for streaming
                stream_messages = [
                    {"role": "system", "content": system_prompt},
                    {"role": "user", "content": user_prompt}
                ]
                
                # Create the stream (await first, then iterate)
                stream = await rag_system.openai_client.client.chat.completions.create(
                    model=model_to_use,
                    messages=stream_messages,
                    temperature=0.7,
                    max_tokens=1500,
                    stream=True
                )
                
                async for chunk in stream:
                    if chunk.choices[0].delta.content:
                        yield f"data: {json.dumps({'type': 'content', 'content': chunk.choices[0].delta.content})}\n\n"
                        await asyncio.sleep(0.001)
                
                yield f"data: {json.dumps({'type': 'done'})}\n\n"
                
            except Exception as e:
                logger.error("RAG streaming error", error=str(e))
                yield f"data: {json.dumps({'type': 'error', 'message': str(e)})}\n\n"
                
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

@app.post("/cancel")
async def cancel_operation():
    """Cancel current operation"""
    try:
        logger.info("🛑 Cancel operation requested")
        
        # For now, we'll just log the cancellation request
        # In a more advanced implementation, you could:
        # 1. Track active operations with unique IDs
        # 2. Use asyncio.Task cancellation
        # 3. Set cancellation flags for long-running operations
        
        return {
            "success": True,
            "message": "Cancel request received",
            "timestamp": time.time()
        }
        
    except Exception as e:
        logger.error("Failed to process cancel request", error=str(e))
        raise HTTPException(status_code=500, detail=f"Cancel failed: {str(e)}")

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
                    
                    # Auto-adjust parameters for knowledge base overview queries
                    search_limit = request.search_limit
                    min_score = 0.3
                    
                    overview_keywords = OVERVIEW_KEYWORDS
                    
                    if any(keyword in request.message.lower() for keyword in overview_keywords):
                        search_limit = min(10, search_limit * 10)  # Increase limit for overview queries
                        min_score = max(0.1, min_score - 0.2)  # Lower threshold
                        logger.info(f"🔍 Knowledge base overview query detected in chat - using limit={search_limit}, min_score={min_score}")
                    
                    search_results = await rag_system.search_knowledge_base(
                        query=request.message,
                        collections=request.search_collections,
                        limit=search_limit,
                        min_score=min_score
                    )
                    
                    if search_results:
                        # Filter results by relevance score
                        high_relevance = [r for r in search_results if r.score >= 0.5]
                        medium_relevance = [r for r in search_results if 0.3 <= r.score < 0.5]
                        
                        # For overview queries, show pagination info
                        overview_keywords = OVERVIEW_KEYWORDS
                        is_overview_query = any(keyword in request.message.lower() for keyword in overview_keywords)
                        
                        if is_overview_query and len(search_results) > 10:
                            yield f"data: {json.dumps({'type': 'search_progress', 'message': f'✅ Found {len(search_results)} relevant documents (showing first 10, {len(search_results) - 10} more available)'})}\n\n"
                        else:
                            yield f"data: {json.dumps({'type': 'search_progress', 'message': f'✅ Found {len(search_results)} relevant documents (High: {len(high_relevance)}, Medium: {len(medium_relevance)})'})}\n\n"
                        
                        # Build enhanced context from search results
                        # For overview queries, show more results in context but limit sources display
                        max_context_results = min(20, len(search_results)) if is_overview_query else min(5, len(search_results))
                        max_sources_display = min(10, len(search_results)) if is_overview_query else len(search_results)
                        
                        context_parts = []
                        for i, result in enumerate(search_results[:max_context_results]):
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
                # Check if this is a knowledge base overview query
                overview_keywords = OVERVIEW_KEYWORDS
                is_overview_query = any(keyword in request.message.lower() for keyword in overview_keywords)
                
                if is_overview_query:
                    # Extract key topics from the search results for better suggestions
                    topics = set()
                    document_types = set()
                    for result in search_results[:20]:  # Analyze top 20 results
                        # Extract topics from metadata
                        if 'language' in result.metadata:
                            topics.add(f"{result.metadata['language']} programming")
                        if 'repository' in result.metadata:
                            topics.add(f"{result.metadata['repository']} codebase")
                        if 'type' in result.metadata:
                            doc_type = result.metadata['type']
                            document_types.add(doc_type)
                            if doc_type == 'github_code':
                                topics.add("code analysis")
                            elif doc_type == 'web_research':
                                topics.add("research findings")
                        
                        # Extract topics from titles
                        title = result.metadata.get('title', result.metadata.get('document_title', ''))
                        if title and title != 'Unknown Document':
                            # Simple topic extraction from titles
                            title_lower = title.lower()
                            if any(word in title_lower for word in ['api', 'endpoint', 'service']):
                                topics.add("API documentation")
                            if any(word in title_lower for word in ['config', 'setup', 'install']):
                                topics.add("configuration")
                            if any(word in title_lower for word in ['test', 'spec', 'example']):
                                topics.add("testing and examples")
                    
                    topics_list = list(topics)[:6]  # Limit to 6 main topics
                    
                    system_prompt = f"""You are VittoriaDB Assistant providing a comprehensive knowledge base overview with intelligent suggestions.

🔍 **KNOWLEDGE BASE OVERVIEW REQUEST DETECTED**
- **Found**: {len(search_results)} documents in your knowledge base
- **Collections**: Documents, GitHub repositories, web research, and chat history
- **Current Date**: {time.strftime('%B %d, %Y')} - Information is current and searchable

🚨 **OVERVIEW INSTRUCTIONS:**
1. **PROVIDE COMPREHENSIVE SUMMARY**: List and categorize all found documents by type and topic
2. **ORGANIZE BY TYPE**: Group by document type (📄 UPLOADED DOCUMENT, 🌐 WEB SEARCH, 📚 KNOWLEDGE BASE, 💻 CODE)
3. **INCLUDE DETAILS**: Show document titles, types, relevance scores, and brief descriptions
4. **IDENTIFY KEY TOPICS**: Highlight main themes and subjects available
5. **SUGGEST FOLLOW-UP QUESTIONS**: Provide specific, actionable questions for deeper exploration

📚 **KNOWLEDGE BASE CONTENTS** (Retrieved {time.strftime('%B %d, %Y at %H:%M')}):
{context_text}

⚡ **REQUIRED RESPONSE FORMAT:**

## 📊 Knowledge Base Summary
- Total documents: {len(search_results)}
- Document types: {', '.join(document_types) if document_types else 'Various'}
- Main topics identified: {', '.join(topics_list) if topics_list else 'General content'}

## 📄 Document Inventory
[Organize documents by type with titles, brief descriptions, and relevance scores]

## 🔍 Key Topics Available
[List 4-6 main topics/themes found in the knowledge base]

## 💡 Suggested Follow-up Questions
**CRITICAL**: End your response with exactly 4-6 specific follow-up questions in this format:
```suggestions
What specific APIs are documented in the codebase?
How do I configure the development environment?
What testing frameworks are being used?
Can you explain the main architecture components?
What are the deployment procedures?
Show me examples of the core functionality?
```

**Remember**: 
- Be thorough and organized in your overview
- Make suggestions specific and actionable
- Always include the suggestions section at the end
- Focus on helping users discover valuable information in their knowledge base"""
                else:
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
            
            # Build messages array with chat history
            messages = [{"role": "system", "content": system_prompt}]
            
            # Add chat history (last 10 messages to maintain context)
            if request.chat_history:
                for msg in request.chat_history[-10:]:
                    messages.append({
                        "role": msg.role.value if hasattr(msg.role, 'value') else str(msg.role),
                        "content": msg.content
                    })
            
            # Add current user message
            messages.append({"role": "user", "content": request.message})
            
            # Build conversation history as prompt
            conversation_prompt = ""
            for msg in messages[1:]:  # Skip system prompt
                role = msg['role']
                conversation_prompt += f"{role.capitalize()}: {msg['content']}\n"
            conversation_prompt += "Assistant:"
            
            # Stream response using Datapizza AI OpenAI client
            response_chunks = []
            # Create the stream (await first, then iterate)
            stream = await rag_system.openai_client.client.chat.completions.create(
                model=request.model,
                messages=messages,
                temperature=0.7,
                max_tokens=1500,
                stream=True
            )
            
            async for chunk in stream:
                if chunk.choices[0].delta.content:
                    content = chunk.choices[0].delta.content
                    response_chunks.append(content)
                    yield f"data: {json.dumps({'type': 'content', 'content': content})}\n\n"
            
            # Extract suggestions from the complete response
            full_response = ''.join(response_chunks)
            suggestions = extract_suggestions_from_response(full_response)
            
            if suggestions:
                yield f"data: {json.dumps({'type': 'suggestions', 'suggestions': suggestions})}\n\n"
            
            # Note: Search results would be processed here in a full implementation
            # For now, we're prioritizing immediate streaming response
            
            # Send limited sources to frontend
            if 'search_results' in locals() and search_results:
                sources_to_send = search_results[:10]  # Limit to 10 sources
                sources_data = []
                for result in sources_to_send:
                    title = result.metadata.get('title', result.metadata.get('document_title', 'Unknown Document'))
                    sources_data.append({
                        'href': f"#source-{hash(result.content) % 10000}",
                        'title': title,
                        'score': result.score,
                        'source': result.source
                    })
                
                # Send sources message
                sources_message = {
                    'type': 'sources',
                    'sources': sources_data,
                    'total_sources': len(search_results),
                    'displayed_sources': len(sources_data),
                    'has_more_sources': len(search_results) > len(sources_data),
                    'is_overview_query': is_overview_query if 'is_overview_query' in locals() else False
                }
                yield f"data: {json.dumps(sources_message)}\n\n"
            
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

async def handle_web_research_websocket(request_data: dict, websocket: WebSocket):
    """Handle web research request via WebSocket"""
    try:
        query = request_data.get('query', '')
        search_engine = request_data.get('search_engine', 'duckduckgo')
        max_results = request_data.get('max_results', 5)
        
        if not query:
            await manager.send_personal_message({
                "type": "web_research_error",
                "error": "Query is required"
            }, websocket)
            return
        
        start_time = time.time()
        research_id = f"web_research_{int(time.time() * 1000)}"
        
        # Send start message
        await manager.send_personal_message({
            "type": "web_research_start",
            "content": f"Starting web research for: {query}",
            "research_id": research_id
        }, websocket)
        
        logger.info("🚀 Starting WebSocket web research", query=query)
        
        # Perform research with Crawl4AI - use 'simple' strategy to avoid scipy errors
        search_results = await advanced_web_researcher.research_query(
            query=query,
            search_engine=search_engine,
            scrape_content=True,
            extraction_strategy='simple'  # Use simple extraction to avoid distance matrix errors
        )
        
        if not search_results:
            processing_time = time.time() - start_time
            await manager.send_personal_message({
                "type": "web_research_error",
                "error": "No results found",
                "processing_time": processing_time
            }, websocket)
            return
        
        # Send progress with found results
        await manager.send_personal_message({
            "type": "web_research_progress",
            "content": f"Found {len(search_results)} results, processing...",
            "progress": 25,
            "results": [{"title": r.get('title', ''), "url": r.get('url', '')} for r in search_results[:5]]
        }, websocket)
        
        # Store results in RAG system
        stored_ids = []
        for i, result in enumerate(search_results):
            try:
                # Send progress update
                progress = 25 + (i / len(search_results)) * 70  # 25-95%
                await manager.send_personal_message({
                    "type": "web_research_progress",
                    "content": f"Processing result {i+1}/{len(search_results)}: {result.get('title', 'Untitled')[:50]}...",
                    "progress": int(progress)
                }, websocket)
                
                # Create enhanced content
                enhanced_content = f"""
Title: {result.get('title', 'No title')}
URL: {result.get('url', 'No URL')}
Content: {result.get('content', result.get('text', 'No content'))}
"""
                
                # Store in RAG system
                document_id = await rag_system.add_document(
                    content=enhanced_content,
                    metadata={
                        'source': 'web_search',
                        'source_collection': 'web_research',
                        'url': result.get('url', ''),
                        'title': result.get('title', ''),
                        'query': query,
                        'search_engine': search_engine,
                        'timestamp': time.time(),
                        'research_id': research_id
                    },
                    collection_name='web_research'
                )
                stored_ids.append(document_id)
                
            except Exception as e:
                logger.error("Failed to store web research result", error=str(e), url=result.get('url', ''))
                continue
        
        processing_time = time.time() - start_time
        
        # Send completion message
        result_response = {
            "success": True,
            "message": f"Web research completed: {len(stored_ids)} results stored",
            "query": query,
            "results_count": len(search_results),
            "stored_count": len(stored_ids),
            "processing_time": processing_time,
            "results": [
                {
                    "title": r.get('title', ''),
                    "url": r.get('url', ''),
                    "snippet": r.get('content', r.get('text', ''))[:200] + '...' if r.get('content') or r.get('text') else ''
                }
                for r in search_results[:10]
            ]
        }
        
        await manager.send_personal_message({
            "type": "web_research_complete",
            "content": f"Research completed: {len(stored_ids)} results stored",
            "progress": 100,
            "results": result_response
        }, websocket)
        
        logger.info("✅ WebSocket web research completed", 
                   query=query, 
                   results_count=len(search_results),
                   stored_count=len(stored_ids),
                   processing_time=processing_time)
        
    except Exception as e:
        logger.error("WebSocket web research failed", query=request_data.get('query', ''), error=str(e))
        await manager.send_personal_message({
            "type": "web_research_error",
            "error": f"Web research failed: {str(e)}"
        }, websocket)

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
            
            logger.info("WebSocket request", type=request_data.get('type', 'chat'), message=request_data.get('message', '')[:100])
            
            try:
                # Handle different message types
                if request_data.get('type') == 'web_research':
                    # Handle web research request
                    await handle_web_research_websocket(request_data, websocket)
                    continue
                
                # Parse chat request
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
                    
                    # Perform web research with Crawl4AI
                    research_results = await advanced_web_researcher.research_query(
                        query=request.message,
                        search_engine='duckduckgo',
                        scrape_content=True,
                        extraction_strategy='simple'
                    )
                    
                    # Store results
                    stored_count = 0
                    for result in research_results:
                        try:
                            # Create enhanced content
                            content_parts = [result.content]
                            if result.structured_data:
                                content_parts.append(f"\n\n**Structured Data:**\n{str(result.structured_data)}")
                            if result.markdown_content:
                                content_parts.append(f"\n\n**Markdown Content:**\n{result.markdown_content[:1000]}...")
                            if result.links:
                                links_text = "\n".join([f"- [{link['text']}]({link['url']})" for link in result.links[:5]])
                                content_parts.append(f"\n\n**Related Links:**\n{links_text}")
                            
                            enhanced_content = "\n".join(content_parts)
                            
                            metadata = {
                                'type': 'web_research_crawl4ai',
                                'title': result.title,
                                'document_title': result.title,
                                'url': result.url,
                                'source': result.source,
                                'source_collection': 'web_search',
                                'query': request.message,
                                'timestamp': result.timestamp,
                                'content_length': len(result.content),
                                'has_structured_data': bool(result.structured_data),
                                'has_markdown': bool(result.markdown_content),
                                'links_count': len(result.links) if result.links else 0,
                                'media_count': len(result.media) if result.media else 0,
                                'extraction_method': 'crawl4ai_cosine'
                            }
                            
                            await rag_system.add_document(
                                content=enhanced_content,
                                metadata=metadata,
                                collection_name='web_research'
                            )
                            stored_count += 1
                        except Exception as e:
                            logger.error(f"Failed to store result {result.url}: {e}")
                            continue
                    
                    research_result = {
                        'success': True,
                        'results_count': len(research_results),
                        'stored_count': stored_count,
                        'message': f'Crawl4AI research completed: {stored_count} enhanced results stored'
                    }
                    
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
                
                # Convert search results (limit to 10)
                sources = [
                    SearchResult(
                        content=result.content,
                        metadata=result.metadata,
                        score=result.score,
                        source=result.source
                    )
                    for result in search_results[:10]  # Limit to 10 sources
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
        
        # Store chunks in VittoriaDB using batch operations for better performance
        total_chunks = len(processed_doc.chunks)
        batch_size = 10  # Process chunks in batches
        chunks_stored = 0
        
        await notification_service.notify_processing_progress(
            document_id=document_id,
            progress=55,
            message=f"Preparing {total_chunks} chunks for batch processing..."
        )
        
        # Prepare documents for batch insertion
        documents = []
        for chunk in processed_doc.chunks:
            # Include document title in chunk metadata
            chunk_metadata = {
                **chunk.metadata,
                'title': processed_doc.title,  # Add document title
                'document_title': processed_doc.title  # Also add as document_title for compatibility
            }
            documents.append({
                'content': chunk.content,
                'metadata': chunk_metadata
            })
        
        # Process chunks in batches
        total_batches = (len(documents) + batch_size - 1) // batch_size
        
        for batch_idx in range(0, len(documents), batch_size):
            batch_docs = documents[batch_idx:batch_idx + batch_size]
            batch_num = (batch_idx // batch_size) + 1
            
            try:
                # Update progress
                progress = 60 + int((batch_num - 1) / total_batches * 35)  # 60-95%
                await notification_service.notify_processing_progress(
                    document_id=document_id,
                    progress=progress,
                    message=f"Storing batch {batch_num}/{total_batches} ({len(batch_docs)} chunks)..."
                )
                
                # Use batch insertion
                await rag_system.add_documents_batch(
                    documents=batch_docs,
                    collection_name='documents'
                )
                
                chunks_stored += len(batch_docs)
                
            except Exception as e:
                logger.error(f"Batch insertion failed for batch {batch_num}, falling back to individual insertion: {e}")
                # Fallback to individual insertion
                for doc in batch_docs:
                    try:
                        await rag_system.add_document(
                            content=doc['content'],
                            metadata=doc['metadata'],
                            collection_name='documents'
                        )
                        chunks_stored += 1
                    except Exception as fallback_error:
                        logger.error(f"Failed to store individual chunk: {fallback_error}")
        
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

# Main web research endpoint (now using Crawl4AI)
@app.post("/research/stream")
async def web_research_stream(request: WebResearchRequest):
    """Perform web research with streaming updates"""
    async def generate_research_stream():
        try:
            start_time = time.time()
            research_id = f"web_research_{int(time.time() * 1000)}"
            
            # Send start event
            yield f"data: {json.dumps({'type': 'start', 'research_id': research_id, 'query': request.query, 'message': f'Starting web research for: {request.query}'})}\n\n"
            
            logger.info("🚀 Starting web research with Crawl4AI", query=request.query)
            
            # Perform research with Crawl4AI
            search_results = await advanced_web_researcher.research_query(
                query=request.query,
                search_engine=request.search_engine,
                scrape_content=True,
                extraction_strategy='simple'  # Use simple extraction to avoid scipy errors
            )
            
            if not search_results:
                processing_time = time.time() - start_time
                yield f"data: {json.dumps({'type': 'error', 'message': 'No results found', 'processing_time': processing_time})}\n\n"
                return
            
            # Send initial results found
            yield f"data: {json.dumps({'type': 'results_found', 'count': len(search_results), 'message': f'Found {len(search_results)} results, processing...'})}\n\n"
            
            # Stream individual results as they're processed
            stored_ids = []
            for i, result in enumerate(search_results):
                try:
                    # Send result details
                    result_data = {
                        'type': 'result_detail',
                        'index': i,
                        'title': result.title or f"Result {i+1}",
                        'url': result.url,
                        'content_preview': result.content[:200] + "..." if len(result.content) > 200 else result.content,
                        'status': 'scraped' if result.content else 'found'
                    }
                    yield f"data: {json.dumps(result_data)}\n\n"
                    
                    if not result.content:
                        continue
                    
                    # Send storing status
                    yield f"data: {json.dumps({'type': 'storing', 'index': i, 'message': f'Storing: {result.title or result.url}'})}\n\n"
                    
                    # Create enhanced content with Crawl4AI data
                    content_parts = [result.content]
                    
                    if result.structured_data:
                        content_parts.append(f"\n\n**Structured Data:**\n{str(result.structured_data)}")
                    
                    if result.markdown_content:
                        content_parts.append(f"\n\n**Markdown Content:**\n{result.markdown_content[:1000]}...")
                    
                    if result.links:
                        links_text = "\n".join([f"- [{link['text']}]({link['url']})" for link in result.links[:5]])
                        content_parts.append(f"\n\n**Related Links:**\n{links_text}")
                    
                    enhanced_content = "\n".join(content_parts)
                    
                    # Enhanced metadata with Crawl4AI features
                    metadata = {
                        'type': 'web_research_crawl4ai',
                        'title': result.title,
                        'document_title': result.title,
                        'url': result.url,
                        'source': result.source,
                        'source_collection': 'web_search',
                        'query': request.query,
                        'timestamp': result.timestamp,
                        'content_length': len(result.content),
                        'has_structured_data': bool(result.structured_data),
                        'has_markdown': bool(result.markdown_content),
                        'links_count': len(result.links) if result.links else 0,
                        'extraction_strategy': 'cosine'
                    }
                    
                    # Store in RAG system
                    doc_id = await rag_system.add_document(
                        content=enhanced_content,
                        metadata=metadata,
                        collection_name="web_research"
                    )
                    
                    if doc_id:
                        stored_ids.append(doc_id)
                        # Send stored confirmation
                        yield f"data: {json.dumps({'type': 'result_stored', 'index': i, 'doc_id': doc_id, 'title': result.title})}\n\n"
                    else:
                        # Send storage error
                        yield f"data: {json.dumps({'type': 'storage_error', 'index': i, 'message': f'Failed to store: {result.title}'})}\n\n"
                        
                except Exception as e:
                    logger.error("Failed to store result", result_url=result.url, error=str(e))
                    yield f"data: {json.dumps({'type': 'storage_error', 'index': i, 'message': f'Error storing {result.title}: {str(e)}'})}\n\n"
            
            processing_time = time.time() - start_time
            
            # Send completion
            completion_data = {
                'type': 'complete',
                'research_id': research_id,
                'query': request.query,
                'message': f'Web research completed: {len(stored_ids)} results stored',
                'urls_found': len(search_results),
                'urls_scraped': len([r for r in search_results if r.content]),
                'results_stored': len(stored_ids),
                'processing_time': processing_time
            }
            yield f"data: {json.dumps(completion_data)}\n\n"
            
        except Exception as e:
            logger.error("Web research failed", query=request.query, error=str(e))
            yield f"data: {json.dumps({'type': 'error', 'message': f'Web research failed: {str(e)}'})}\n\n"
    
    return StreamingResponse(
        generate_research_stream(),
        media_type="text/plain",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "Content-Type": "text/event-stream",
        }
    )

@app.post("/research", response_model=WebResearchResponse)
async def web_research(request: WebResearchRequest, background_tasks: BackgroundTasks):
    """Perform web research using Crawl4AI and store results"""
    try:
        start_time = time.time()
        research_id = f"web_research_{int(time.time() * 1000)}"
        
        # Send start notification
        await notification_service.send_notification({
            "type": "web_research_start",
            "data": {
                "research_id": research_id,
                "query": request.query,
                "message": f"Starting web research for: {request.query}"
            }
        })
        
        logger.info("🚀 Starting web research with Crawl4AI", query=request.query)
        
        # Perform research with Crawl4AI
        search_results = await advanced_web_researcher.research_query(
            query=request.query,
            search_engine=request.search_engine,
            scrape_content=True,
            extraction_strategy='simple'  # Use semantic extraction
        )
        
        if not search_results:
            processing_time = time.time() - start_time
            return WebResearchResponse(
                success=False,
                message="No results found",
                query=request.query,
                results_count=0,
                results=[],
                stored_count=0,
                processing_time=processing_time
            )
        
        # Store results in RAG system
        stored_ids = []
        for result in search_results:
            try:
                # Create enhanced content with Crawl4AI data
                content_parts = [result.content]
                
                if result.structured_data:
                    content_parts.append(f"\n\n**Structured Data:**\n{str(result.structured_data)}")
                
                if result.markdown_content:
                    content_parts.append(f"\n\n**Markdown Content:**\n{result.markdown_content[:1000]}...")
                
                if result.links:
                    links_text = "\n".join([f"- [{link['text']}]({link['url']})" for link in result.links[:5]])
                    content_parts.append(f"\n\n**Related Links:**\n{links_text}")
                
                enhanced_content = "\n".join(content_parts)
                
                # Enhanced metadata with Crawl4AI features
                metadata = {
                    'type': 'web_research_crawl4ai',
                    'title': result.title,
                    'document_title': result.title,
                    'url': result.url,
                    'source': result.source,
                    'source_collection': 'web_search',
                    'query': request.query,
                    'timestamp': result.timestamp,
                    'content_length': len(result.content),
                    'has_structured_data': bool(result.structured_data),
                    'has_markdown': bool(result.markdown_content),
                    'links_count': len(result.links) if result.links else 0,
                    'media_count': len(result.media) if result.media else 0,
                    'extraction_method': 'crawl4ai_cosine'
                }
                
                # Store in web_research collection
                doc_id = await rag_system.add_document(
                    content=enhanced_content,
                    metadata=metadata,
                    collection_name='web_research'
                )
                
                stored_ids.append(doc_id)
                logger.info(f"✅ Stored Crawl4AI result: {result.title[:50]}...")
                
            except Exception as e:
                logger.error(f"Failed to store result {result.url}: {e}")
                continue
        
        # Convert results for response
        response_results = []
        for result in search_results:
            response_results.append({
                'title': result.title,
                'url': result.url,
                'snippet': result.snippet,
                'content_preview': result.content[:200] + "..." if len(result.content) > 200 else result.content,
                'source': result.source,
                'timestamp': result.timestamp,
                'enhanced_features': {
                    'has_structured_data': bool(result.structured_data),
                    'has_markdown': bool(result.markdown_content),
                    'links_found': len(result.links) if result.links else 0,
                    'media_found': len(result.media) if result.media else 0
                }
            })
        
        logger.info("✅ Web research completed with Crawl4AI", 
                   query=request.query,
                   results_count=len(search_results),
                   stored_count=len(stored_ids))
        
        processing_time = time.time() - start_time
        
        # Prepare detailed results for notification
        detailed_results = []
        for i, result in enumerate(search_results):
            stored_id = stored_ids[i] if i < len(stored_ids) else None
            detailed_results.append({
                "title": result.title or f"Result {i+1}",
                "url": result.url,
                "content_preview": result.content[:200] + "..." if len(result.content) > 200 else result.content,
                "status": "stored" if stored_id else ("scraped" if result.content else "error")
            })
        
        # Send completion notification
        await notification_service.send_notification({
            "type": "web_research_complete",
            "data": {
                "research_id": research_id,
                "query": request.query,
                "message": f"Web research completed: {len(stored_ids)} results stored",
                "urls_found": len(search_results),
                "urls_scraped": len(search_results),
                "results_stored": len(stored_ids),
                "processing_time": processing_time,
                "results": detailed_results
            }
        })
        
        return WebResearchResponse(
            success=True,
            message=f"Web research completed with Crawl4AI. Found {len(search_results)} results, stored {len(stored_ids)} successfully.",
            query=request.query,
            results_count=len(search_results),
            results=response_results,
            stored_count=len(stored_ids),
            processing_time=processing_time
        )
        
    except Exception as e:
        logger.error("Web research failed", query=request.query, error=str(e))
        
        # Send error notification
        await notification_service.send_notification({
            "type": "web_research_error",
            "data": {
                "research_id": research_id if 'research_id' in locals() else "unknown",
                "query": request.query,
                "message": f"Web research failed: {str(e)}",
                "error": str(e)
            }
        })
        
        raise HTTPException(status_code=500, detail=f"Web research failed: {str(e)}")

# Legacy web research endpoint (BeautifulSoup)
@app.post("/research/legacy", response_model=WebResearchResponse)
async def legacy_web_research(request: WebResearchRequest, background_tasks: BackgroundTasks):
    """Legacy web research using BeautifulSoup (for fallback/comparison)"""
    try:
        logger.info("Starting legacy web research", query=request.query)
        
        result = await web_researcher.research_and_store(
            query=request.query,
            rag_system=rag_system,
            search_engine=request.search_engine
        )
        
        logger.info("Legacy web research completed", 
                   query=request.query,
                   results_count=result.get('results_count', 0))
        
        return WebResearchResponse(**result)
        
    except Exception as e:
        logger.error("Legacy web research failed", query=request.query, error=str(e))
        raise HTTPException(status_code=500, detail=f"Legacy web research failed: {str(e)}")

# Alternative Crawl4AI endpoint (for testing different strategies)
@app.post("/research/crawl4ai", response_model=WebResearchResponse)
async def advanced_web_research(request: WebResearchRequest, background_tasks: BackgroundTasks):
    """Perform advanced web research using Crawl4AI and store results"""
    try:
        start_time = time.time()
        logger.info("🚀 Starting advanced web research with Crawl4AI", query=request.query)
        
        # Perform research with Crawl4AI
        search_results = await advanced_web_researcher.research_query(
            query=request.query,
            search_engine=request.search_engine,
            scrape_content=True,
            extraction_strategy='simple'  # Use semantic extraction
        )
        
        if not search_results:
            processing_time = time.time() - start_time
            return WebResearchResponse(
                success=False,
                message="No results found",
                query=request.query,
                results_count=0,
                results=[],
                stored_count=0,
                processing_time=processing_time
            )
        
        # Store results in RAG system
        stored_ids = []
        for result in search_results:
            try:
                # Create enhanced content with Crawl4AI data
                content_parts = [result.content]
                
                if result.structured_data:
                    content_parts.append(f"\n\n**Structured Data:**\n{str(result.structured_data)}")
                
                if result.markdown_content:
                    content_parts.append(f"\n\n**Markdown Content:**\n{result.markdown_content[:1000]}...")
                
                if result.links:
                    links_text = "\n".join([f"- [{link['text']}]({link['url']})" for link in result.links[:5]])
                    content_parts.append(f"\n\n**Related Links:**\n{links_text}")
                
                enhanced_content = "\n".join(content_parts)
                
                # Enhanced metadata with Crawl4AI features
                metadata = {
                    'type': 'web_research_crawl4ai',
                    'title': result.title,
                    'document_title': result.title,
                    'url': result.url,
                    'source': result.source,
                    'source_collection': 'web_search',
                    'query': request.query,
                    'timestamp': result.timestamp,
                    'content_length': len(result.content),
                    'has_structured_data': bool(result.structured_data),
                    'has_markdown': bool(result.markdown_content),
                    'links_count': len(result.links) if result.links else 0,
                    'media_count': len(result.media) if result.media else 0,
                    'extraction_method': 'crawl4ai_cosine'
                }
                
                # Store in web_research collection
                doc_id = await rag_system.add_document(
                    content=enhanced_content,
                    metadata=metadata,
                    collection_name='web_research'
                )
                
                stored_ids.append(doc_id)
                logger.info(f"✅ Stored Crawl4AI result: {result.title[:50]}...")
                
            except Exception as e:
                logger.error(f"Failed to store result {result.url}: {e}")
                continue
        
        # Convert results for response
        response_results = []
        for result in search_results:
            response_results.append({
                'title': result.title,
                'url': result.url,
                'snippet': result.snippet,
                'content_preview': result.content[:200] + "..." if len(result.content) > 200 else result.content,
                'source': result.source,
                'timestamp': result.timestamp,
                'enhanced_features': {
                    'has_structured_data': bool(result.structured_data),
                    'has_markdown': bool(result.markdown_content),
                    'links_found': len(result.links) if result.links else 0,
                    'media_found': len(result.media) if result.media else 0
                }
            })
        
        logger.info("✅ Advanced web research completed", 
                   query=request.query,
                   results_count=len(search_results),
                   stored_count=len(stored_ids))
        
        processing_time = time.time() - start_time
        
        return WebResearchResponse(
            success=True,
            message=f"Advanced research completed with Crawl4AI. Found {len(search_results)} results, stored {len(stored_ids)} successfully.",
            query=request.query,
            results_count=len(search_results),
            results=response_results,
            stored_count=len(stored_ids),
            processing_time=processing_time
        )
        
    except Exception as e:
        logger.error("Advanced web research failed", query=request.query, error=str(e))
        raise HTTPException(status_code=500, detail=f"Advanced web research failed: {str(e)}")

# Background GitHub indexing function
async def background_github_indexing(repo_url: str, indexing_id: str):
    """Background GitHub repository indexing with notifications"""
    try:
        # Notify indexing start
        await notification_service.notify_github_indexing_start(repo_url, indexing_id)
        
        # Progress updates
        await notification_service.notify_github_indexing_progress(
            indexing_id, 10, "Connecting to GitHub repository..."
        )
        
        # Perform the actual indexing
        result = await github_indexer.index_and_store(
            repo_url=repo_url,
            rag_system=rag_system,
            progress_callback=lambda progress, message: asyncio.create_task(
                notification_service.notify_github_indexing_progress(indexing_id, progress, message)
            )
        )
        
        # Calculate processing time
        processing_time = time.time() - notification_service.processing_status[indexing_id]['start_time']
        
        # Notify completion
        await notification_service.notify_github_indexing_complete(
            indexing_id=indexing_id,
            files_indexed=result.get('files_indexed', 0),
            repository=result.get('repository', repo_url),
            processing_time=processing_time
        )
        
        # Update stats
        stats = await rag_system.get_stats()
        await notification_service.notify_stats_update(stats)
        
        logger.info("Background GitHub indexing completed", 
                   repo_url=repo_url,
                   indexing_id=indexing_id,
                   files_indexed=result.get('files_indexed', 0))
        
    except Exception as e:
        logger.error("Background GitHub indexing failed", 
                    repo_url=repo_url,
                    indexing_id=indexing_id, 
                    error=str(e))
        await notification_service.notify_github_indexing_error(indexing_id, str(e))

# GitHub indexing endpoint (now returns immediately)
@app.post("/github/index")
async def index_github_repo(request: GitHubIndexRequest, background_tasks: BackgroundTasks):
    """Start GitHub repository indexing in background"""
    try:
        # Generate unique indexing ID
        indexing_id = f"github_{int(time.time())}_{hash(request.repository_url) % 10000}"
        
        logger.info("Starting background GitHub indexing", 
                   repo_url=request.repository_url,
                   indexing_id=indexing_id)
        
        # Start background indexing
        background_tasks.add_task(background_github_indexing, request.repository_url, indexing_id)
        
        # Return immediately with indexing ID
        return {
            "success": True,
            "message": "GitHub indexing started in background",
            "indexing_id": indexing_id,
            "repository_url": request.repository_url
        }
        
    except Exception as e:
        logger.error("Failed to start GitHub indexing", 
                    repo_url=request.repository_url, 
                    error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to start GitHub indexing: {str(e)}")

# Search endpoint
@app.post("/search", response_model=SearchResponse)
async def search_knowledge_base(request: SearchRequest):
    """Search across knowledge base"""
    start_time = time.time()
    
    try:
        logger.info("Knowledge base search", query=request.query)
        
        # Auto-adjust parameters for knowledge base overview queries
        limit = request.limit
        min_score = request.min_score
        
        overview_keywords = OVERVIEW_KEYWORDS
        
        is_overview_query = any(keyword in request.query.lower() for keyword in overview_keywords)
        if is_overview_query:
            limit = min(10, limit * 10)  # Increase limit for overview queries
            min_score = max(0.1, min_score - 0.2)  # Lower threshold
            logger.info(f"Knowledge base overview query detected - using limit={limit}, min_score={min_score}")
        
        search_results = await rag_system.search_knowledge_base(
            query=request.query,
            collections=request.collections,
            limit=limit,
            min_score=min_score
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
        
        # For overview queries, limit displayed results to 10 initially
        displayed_results = results
        has_more = False
        
        if is_overview_query and len(results) > 10:
            displayed_results = results[:10]
            has_more = True
        
        return SearchResponse(
            query=request.query,
            results=displayed_results,
            total_results=len(results),
            processing_time=processing_time,
            is_overview_query=is_overview_query,
            displayed_results=len(displayed_results),
            has_more=has_more
        )
        
    except Exception as e:
        logger.error("Search failed", query=request.query, error=str(e))
        raise HTTPException(status_code=500, detail=f"Search failed: {str(e)}")

@app.post("/search/more", response_model=SearchResponse)
async def search_more_results(request: SearchRequest, offset: int = 10):
    """Get additional search results for pagination"""
    start_time = time.time()
    
    try:
        logger.info("Search more request", query=request.query, offset=offset)
        
        # Auto-adjust parameters for knowledge base overview queries
        limit = request.limit
        min_score = request.min_score
        
        overview_keywords = OVERVIEW_KEYWORDS
        
        is_overview_query = any(keyword in request.query.lower() for keyword in overview_keywords)
        if is_overview_query:
            limit = min(50, limit * 10)  # Increase limit for overview queries
            min_score = max(0.1, min_score - 0.2)  # Lower threshold
        
        # Get all results
        search_results = await rag_system.search_knowledge_base(
            query=request.query,
            collections=request.collections,
            limit=limit,
            min_score=min_score
        )
        
        # Convert search results to Pydantic models
        from models import SearchResult as PydanticSearchResult
        all_results = [
            PydanticSearchResult(
                content=result.content,
                metadata=result.metadata,
                score=result.score,
                source=result.source
            )
            for result in search_results
        ]
        
        # Return results starting from offset
        remaining_results = all_results[offset:]
        processing_time = time.time() - start_time
        
        logger.info("Search more completed", 
                   query=request.query,
                   offset=offset,
                   remaining_count=len(remaining_results),
                   processing_time=processing_time)
        
        return SearchResponse(
            query=request.query,
            results=remaining_results,
            total_results=len(all_results),
            processing_time=processing_time,
            is_overview_query=is_overview_query,
            displayed_results=len(remaining_results),
            has_more=False  # This endpoint returns all remaining results
        )
        
    except Exception as e:
        logger.error("Search more failed", query=request.query, offset=offset, error=str(e))
        raise HTTPException(status_code=500, detail=f"Search more failed: {str(e)}")

# Debug endpoint for AI query optimization
@app.post("/debug/optimize-query")
async def debug_optimize_query(request: dict):
    """Debug endpoint to test AI query optimization"""
    try:
        query = request.get('query', '')
        if not query:
            return {"error": "Query is required"}
        
        # Test the AI optimization with language detection
        optimized, detected_language = await advanced_web_researcher._optimize_search_query_with_ai(query)
        cleaned = advanced_web_researcher._clean_search_query(query)
        
        return {
            "original_query": query,
            "ai_optimized": optimized,
            "detected_language": detected_language,
            "rule_based_cleaned": cleaned,
            "improvement": "AI optimization adds context, synonyms, and detects language for region-specific search"
        }
        
    except Exception as e:
        return {"error": str(e)}

# Document listing endpoints
@app.get("/documents/{collection_name}/original")
async def get_original_documents(collection_name: str):
    """Get original documents (grouped by source file) instead of individual chunks"""
    try:
        logger.info(f"📋 Getting original documents from {collection_name} collection")
        
        documents = await rag_system.get_original_documents(collection_name)
        
        return {
            "success": True,
            "collection": collection_name,
            "documents": documents,
            "count": len(documents)
        }
        
    except Exception as e:
        logger.error("Failed to get original documents", collection=collection_name, error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to get original documents: {str(e)}")

@app.get("/documents/{collection_name}")
async def list_documents(collection_name: str, limit: int = 200):
    """List all document chunks in a collection without using embeddings"""
    try:
        logger.info(f"📋 Listing document chunks from {collection_name} collection")
        
        documents = await rag_system.list_documents(collection_name, limit)
        
        return {
            "success": True,
            "collection": collection_name,
            "documents": documents,
            "count": len(documents)
        }
        
    except Exception as e:
        logger.error("Failed to list documents", collection=collection_name, error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to list documents: {str(e)}")

# Document deletion endpoints
@app.delete("/documents/{collection_name}/{document_id}")
async def delete_document_by_id(collection_name: str, document_id: str):
    """Delete a document by its ID"""
    try:
        logger.info(f"🗑️ Delete request for document {document_id} in {collection_name}")
        
        result = await rag_system.delete_document_by_id(document_id, collection_name)
        
        if result['success']:
            logger.info("Document deletion completed", 
                       collection=collection_name,
                       document_id=document_id)
            return {
                "success": True,
                "message": f"Successfully deleted document {document_id}",
                "document_id": document_id,
                "collection": collection_name
            }
        else:
            logger.error("Document deletion failed", error=result['error'])
            raise HTTPException(status_code=500, detail=result['error'])
            
    except HTTPException:
        raise
    except Exception as e:
        logger.error("Document deletion failed", error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to delete document: {str(e)}")

@app.delete("/documents/{collection_name}")
async def delete_document(
    collection_name: str,
    filename: Optional[str] = None,
    title: Optional[str] = None,
    url: Optional[str] = None
):
    """Delete a document and all its chunks from a collection"""
    try:
        if not any([filename, title, url]):
            raise HTTPException(
                status_code=400, 
                detail="At least one of filename, title, or url must be provided"
            )
        
        logger.info("Document deletion request", 
                   collection=collection_name, 
                   filename=filename, 
                   title=title, 
                   url=url)
        
        # Debug: Log what we're searching for
        search_criteria = []
        if filename: search_criteria.append(f"filename='{filename}'")
        if title: search_criteria.append(f"title='{title}'")
        if url: search_criteria.append(f"url='{url}'")
        logger.info(f"🔍 Searching for documents with criteria: {', '.join(search_criteria)}")
        
        result = await rag_system.delete_document_by_metadata(
            filename=filename,
            title=title,
            url=url,
            collection_name=collection_name
        )
        
        if result['success']:
            logger.info("Document deletion completed", 
                       collection=collection_name,
                       deleted_chunks=result['deleted_chunks'],
                       deleted_documents=result['deleted_documents'])
            return {
                "success": True,
                "message": f"Successfully deleted {result['deleted_chunks']} chunks from {len(result['deleted_documents'])} documents",
                "deleted_chunks": result['deleted_chunks'],
                "deleted_documents": result['deleted_documents'],
                "collection": collection_name
            }
        else:
            logger.error("Document deletion failed", error=result['error'])
            raise HTTPException(status_code=500, detail=result['error'])
            
    except HTTPException:
        raise
    except Exception as e:
        logger.error("Document deletion failed", error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to delete document: {str(e)}")

# Configuration endpoints
@app.get("/config", response_model=ConfigResponse)
async def get_config():
    """Get current configuration"""
    return ConfigResponse(
        openai_configured=bool(os.getenv('OPENAI_API_KEY')),
        github_configured=bool(os.getenv('GITHUB_TOKEN')),
        current_model="gpt-4o",  # Default model
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
                        # Convert to dict for JSON serialization
                        stats_dict = stats.dict() if hasattr(stats, 'dict') else stats.__dict__
                        await websocket.send_text(json.dumps({
                            "type": "stats_update",
                            "data": stats_dict,
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

# Chat Session Management Endpoints
@app.post("/chat/sessions", response_model=ChatSessionResponse)
async def create_chat_session(request: ChatSessionCreateRequest):
    """Create a new chat session"""
    try:
        import uuid
        session_id = str(uuid.uuid4())
        
        # Generate title from first message or use default
        title = request.title or f"Chat Session {time.strftime('%Y-%m-%d %H:%M')}"
        
        session = ChatSession(
            session_id=session_id,
            title=title,
            created_at=time.time(),
            updated_at=time.time(),
            message_count=0,
            last_message_preview=""
        )
        
        # Store session metadata in VittoriaDB
        await rag_system.add_document(
            content=f"Chat Session: {title}",
            metadata={
                'type': 'chat_session',
                'session_id': session_id,
                'title': title,
                'created_at': session.created_at,
                'message_count': 0
            },
            collection_name='chat_history'
        )
        
        logger.info("Chat session created", session_id=session_id, title=title)
        
        return ChatSessionResponse(
            session=session,
            success=True,
            message=f"Chat session '{title}' created successfully"
        )
        
    except Exception as e:
        logger.error("Failed to create chat session", error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to create chat session: {str(e)}")

@app.get("/chat/sessions", response_model=List[ChatSession])
async def list_chat_sessions():
    """List all chat sessions"""
    try:
        # Search for chat sessions in VittoriaDB
        search_results = await rag_system.search_knowledge_base(
            query="chat session",
            collections=['chat_history'],
            limit=100,
            min_score=0.1
        )
        
        sessions = []
        for result in search_results:
            if result.metadata.get('type') == 'chat_session':
                session = ChatSession(
                    session_id=result.metadata.get('session_id', ''),
                    title=result.metadata.get('title', 'Unknown Session'),
                    created_at=result.metadata.get('created_at', time.time()),
                    updated_at=result.metadata.get('updated_at', time.time()),
                    message_count=result.metadata.get('message_count', 0),
                    last_message_preview=result.metadata.get('last_message_preview', '')
                )
                sessions.append(session)
        
        # Sort by updated_at descending
        sessions.sort(key=lambda x: x.updated_at, reverse=True)
        
        return sessions
        
    except Exception as e:
        logger.error("Failed to list chat sessions", error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to list chat sessions: {str(e)}")

@app.post("/chat/save", response_model=SaveChatResponse)
async def save_chat_history(request: SaveChatRequest):
    """Save chat messages to a session"""
    try:
        logger.info("Saving chat history", session_id=request.session_id, messages_count=len(request.messages))
        
        # Save each message as a separate document
        saved_count = 0
        last_message_preview = ""
        
        for i, message in enumerate(request.messages):
            message_id = f"{request.session_id}_msg_{i}_{int(time.time())}"
            
            # Create content for the message
            content = f"""
Session: {request.session_id}
Role: {message.role}
Timestamp: {time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(message.timestamp))}

Message:
{message.content}
"""
            
            # Prepare metadata
            metadata = {
                'type': 'chat_message',
                'session_id': request.session_id,
                'role': message.role,
                'timestamp': message.timestamp,
                'message_index': i,
                'content': message.content,
                'sources_count': len(message.sources) if message.sources else 0
            }
            
            # Add sources information if available
            if message.sources:
                metadata['sources'] = [
                    {
                        'content': source.content[:200] + "..." if len(source.content) > 200 else source.content,
                        'score': source.score,
                        'source': source.source,
                        'metadata': source.metadata
                    }
                    for source in message.sources
                ]
            
            # Store in VittoriaDB
            await rag_system.add_document(
                content=content,
                metadata=metadata,
                collection_name='chat_history'
            )
            
            saved_count += 1
            
            # Update last message preview
            if message.role == 'user':
                last_message_preview = message.content[:100] + "..." if len(message.content) > 100 else message.content
        
        # Update session metadata
        session_metadata_content = f"Chat Session {request.session_id} - {saved_count} messages"
        await rag_system.add_document(
            content=session_metadata_content,
            metadata={
                'type': 'chat_session_update',
                'session_id': request.session_id,
                'message_count': saved_count,
                'last_message_preview': last_message_preview,
                'updated_at': time.time()
            },
            collection_name='chat_history'
        )
        
        logger.info("Chat history saved", session_id=request.session_id, saved_count=saved_count)
        
        return SaveChatResponse(
            success=True,
            message=f"Saved {saved_count} messages to session {request.session_id}",
            session_id=request.session_id,
            messages_saved=saved_count
        )
        
    except Exception as e:
        logger.error("Failed to save chat history", session_id=request.session_id, error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to save chat history: {str(e)}")

@app.get("/chat/sessions/{session_id}/history", response_model=ChatHistoryResponse)
async def get_chat_history(session_id: str, limit: int = 50, offset: int = 0):
    """Get chat history for a specific session"""
    try:
        # Search for messages in this session
        search_results = await rag_system.search_knowledge_base(
            query=f"session {session_id}",
            collections=['chat_history'],
            limit=limit + offset + 10,  # Get extra to account for filtering
            min_score=0.1
        )
        
        # Filter and sort messages
        messages = []
        session_info = None
        
        for result in search_results:
            if result.metadata.get('session_id') == session_id:
                if result.metadata.get('type') == 'chat_message':
                    # Reconstruct ChatMessage
                    message = ChatMessage(
                        role=result.metadata.get('role', 'user'),
                        content=result.metadata.get('content', ''),
                        timestamp=result.metadata.get('timestamp', time.time()),
                        sources=[]  # Sources can be reconstructed if needed
                    )
                    messages.append((result.metadata.get('message_index', 0), message))
                
                elif result.metadata.get('type') == 'chat_session' and session_info is None:
                    # Get session info
                    session_info = ChatSession(
                        session_id=session_id,
                        title=result.metadata.get('title', 'Unknown Session'),
                        created_at=result.metadata.get('created_at', time.time()),
                        updated_at=result.metadata.get('updated_at', time.time()),
                        message_count=result.metadata.get('message_count', 0),
                        last_message_preview=result.metadata.get('last_message_preview', '')
                    )
        
        # Sort messages by index and apply pagination
        messages.sort(key=lambda x: x[0])
        sorted_messages = [msg[1] for msg in messages[offset:offset+limit]]
        
        if session_info is None:
            session_info = ChatSession(
                session_id=session_id,
                title=f"Session {session_id[:8]}",
                created_at=time.time(),
                updated_at=time.time(),
                message_count=len(messages),
                last_message_preview=""
            )
        
        return ChatHistoryResponse(
            session_id=session_id,
            messages=sorted_messages,
            total_messages=len(messages),
            session_info=session_info
        )
        
    except Exception as e:
        logger.error("Failed to get chat history", session_id=session_id, error=str(e))
        raise HTTPException(status_code=500, detail=f"Failed to get chat history: {str(e)}")

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
