"""
Pydantic Models for FastAPI Backend
Data models for API requests and responses
"""

from typing import List, Dict, Any, Optional, Union
from pydantic import BaseModel, Field
from enum import Enum
import time

class MessageRole(str, Enum):
    """Chat message roles"""
    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"

class SearchResult(BaseModel):
    """Search result from vector database"""
    content: str
    metadata: Dict[str, Any]
    score: float
    source: str

class ChatMessage(BaseModel):
    """Chat message structure"""
    role: MessageRole
    content: str
    timestamp: float = Field(default_factory=time.time)
    sources: Optional[List[SearchResult]] = None

class ChatRequest(BaseModel):
    """Chat request from frontend"""
    message: str
    chat_history: Optional[List[ChatMessage]] = []
    search_collections: Optional[List[str]] = ["documents", "web_research", "github_code"]
    model: str = "gpt-3.5-turbo"
    search_limit: int = 10  # Increased default for better overview queries
    web_search: bool = False

class ChatResponse(BaseModel):
    """Chat response to frontend"""
    message: str
    sources: List[SearchResult]
    processing_time: float
    search_results_count: int

class StreamingChatResponse(BaseModel):
    """Streaming chat response chunk"""
    type: str  # "content", "sources", "done", "error"
    content: Optional[str] = None
    sources: Optional[List[SearchResult]] = None
    error: Optional[str] = None
    processing_time: Optional[float] = None

class FileUploadResponse(BaseModel):
    """File upload response"""
    success: bool
    message: str
    filename: str
    document_id: str
    chunks_created: int
    processing_time: float
    metadata: Dict[str, Any]

class WebResearchRequest(BaseModel):
    """Web research request"""
    query: str
    search_engine: str = "duckduckgo"
    max_results: int = 5

class WebResearchResponse(BaseModel):
    """Web research response"""
    success: bool
    message: str
    query: str
    results_count: int
    stored_count: int
    processing_time: float
    results: List[Dict[str, str]]

class GitHubIndexRequest(BaseModel):
    """GitHub repository indexing request"""
    repository_url: str
    max_files: int = 500

class GitHubIndexResponse(BaseModel):
    """GitHub indexing response"""
    success: bool
    message: str
    repository: str
    repository_url: str
    files_indexed: int
    files_stored: int
    languages: List[str]
    repository_stars: int
    processing_time: float

class CollectionStats(BaseModel):
    """Collection statistics"""
    name: str
    vector_count: int
    dimensions: int
    metric: str
    index_type: str
    description: str

class SystemStats(BaseModel):
    """System statistics response"""
    collections: Dict[str, CollectionStats]
    total_vectors: int
    vittoriadb_status: str
    uptime: float

class ErrorResponse(BaseModel):
    """Error response"""
    error: str
    detail: Optional[str] = None
    timestamp: float = Field(default_factory=time.time)

class HealthResponse(BaseModel):
    """Health check response"""
    status: str
    vittoriadb_connected: bool
    openai_configured: bool
    timestamp: float = Field(default_factory=time.time)

class DocumentChunk(BaseModel):
    """Document chunk information"""
    id: str
    content: str
    metadata: Dict[str, Any]
    chunk_index: int
    total_chunks: int

class ProcessedDocument(BaseModel):
    """Processed document information"""
    id: str
    title: str
    content: str
    chunks: List[DocumentChunk]
    metadata: Dict[str, Any]
    file_type: str

class SearchRequest(BaseModel):
    """Search request"""
    query: str
    collections: Optional[List[str]] = ["documents", "web_research", "github_code"]
    limit: int = 5
    min_score: float = 0.3

class SearchResponse(BaseModel):
    """Search response"""
    query: str
    results: List[SearchResult]
    total_results: int
    processing_time: float
    is_overview_query: bool = False
    displayed_results: int = 0
    has_more: bool = False

class ConfigUpdate(BaseModel):
    """Configuration update request"""
    openai_api_key: Optional[str] = None
    github_token: Optional[str] = None
    model: Optional[str] = None
    search_limit: Optional[int] = None

class ConfigResponse(BaseModel):
    """Configuration response"""
    openai_configured: bool
    github_configured: bool
    current_model: str
    search_limit: int
    vittoriadb_url: str

class ChatSession(BaseModel):
    """Chat session information"""
    session_id: str
    title: str
    created_at: float = Field(default_factory=time.time)
    updated_at: float = Field(default_factory=time.time)
    message_count: int = 0
    last_message_preview: str = ""

class ChatSessionCreateRequest(BaseModel):
    """Request to create a new chat session"""
    title: Optional[str] = None

class ChatSessionResponse(BaseModel):
    """Chat session response"""
    session: ChatSession
    success: bool
    message: str

class ChatHistoryRequest(BaseModel):
    """Request to get chat history for a session"""
    session_id: str
    limit: int = 50
    offset: int = 0

class ChatHistoryResponse(BaseModel):
    """Chat history response"""
    session_id: str
    messages: List[ChatMessage]
    total_messages: int
    session_info: ChatSession

class SaveChatRequest(BaseModel):
    """Request to save chat messages"""
    session_id: str
    messages: List[ChatMessage]

class SaveChatResponse(BaseModel):
    """Save chat response"""
    success: bool
    message: str
    session_id: str
    messages_saved: int
