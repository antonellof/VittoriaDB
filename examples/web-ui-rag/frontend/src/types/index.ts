// Type definitions for VittoriaDB RAG Frontend

export interface SearchResult {
  content: string;
  metadata: Record<string, any>;
  score: number;
  source: string;
}

export interface ChatMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: number;
  sources?: SearchResult[];
}

export interface ChatRequest {
  message: string;
  chat_history?: ChatMessage[];
  search_collections?: string[];
  model?: string;
  search_limit?: number;
}

export interface ChatResponse {
  message: string;
  sources: SearchResult[];
  processing_time: number;
  search_results_count: number;
}

export interface StreamingChatResponse {
  type: 'content' | 'sources' | 'suggestions' | 'done' | 'error' | 'typing' | 'status';
  content?: string;
  sources?: SearchResult[];
  suggestions?: string[];
  error?: string;
  processing_time?: number;
  total_sources?: number;
  displayed_sources?: number;
  has_more_sources?: boolean;
  is_overview_query?: boolean;
}

export interface FileUploadResponse {
  success: boolean;
  message: string;
  filename: string;
  document_id: string;
  chunks_created: number;
  processing_time: number;
  metadata: Record<string, any>;
}

export interface WebResearchRequest {
  query: string;
  search_engine?: string;
  max_results?: number;
}

export interface WebResearchResponse {
  success: boolean;
  message: string;
  query: string;
  results_count: number;
  stored_count: number;
  processing_time: number;
  results: Array<{
    title: string;
    url: string;
    snippet: string;
  }>;
}

export interface GitHubIndexRequest {
  repository_url: string;
  max_files?: number;
}

export interface GitHubIndexResponse {
  success: boolean;
  message: string;
  repository: string;
  repository_url: string;
  files_indexed: number;
  files_stored: number;
  languages: string[];
  repository_stars: number;
  processing_time: number;
}

export interface CollectionStats {
  name: string;
  vector_count: number;
  dimensions: number;
  metric: string;
  index_type: string;
  description: string;
}

export interface SystemStats {
  collections: Record<string, CollectionStats>;
  total_vectors: number;
  vittoriadb_status: string;
  uptime: number;
}

export interface HealthResponse {
  status: string;
  vittoriadb_connected: boolean;
  openai_configured: boolean;
  timestamp: number;
}

export interface SearchRequest {
  query: string;
  collections?: string[];
  limit?: number;
  min_score?: number;
}

export interface SearchResponse {
  query: string;
  results: SearchResult[];
  total_results: number;
  processing_time: number;
}

export interface ConfigResponse {
  openai_configured: boolean;
  github_configured: boolean;
  current_model: string;
  search_limit: number;
  vittoriadb_url: string;
}

// UI State Types
export interface UIState {
  sidebarOpen: boolean;
  darkMode: boolean;
  currentModel: string;
  searchLimit: number;
}

export interface ChatState {
  messages: ChatMessage[];
  isLoading: boolean;
  isConnected: boolean;
  error: string | null;
}

export interface UploadState {
  isUploading: boolean;
  uploadProgress: number;
  uploadedFiles: string[];
  error: string | null;
}

export interface ResearchState {
  isResearching: boolean;
  lastQuery: string;
  results: WebResearchResponse | null;
  error: string | null;
}

export interface GitHubState {
  isIndexing: boolean;
  lastRepo: string;
  results: GitHubIndexResponse | null;
  error: string | null;
}

export interface StatsState {
  stats: SystemStats | null;
  health: HealthResponse | null;
  isLoading: boolean;
  error: string | null;
}

// Component Props Types
export interface MessageBubbleProps {
  message: ChatMessage;
  isStreaming?: boolean;
}

export interface FileUploadProps {
  onUpload: (files: File[]) => void;
  isUploading: boolean;
  uploadProgress: number;
}

export interface SidebarProps {
  isOpen: boolean;
  onToggle: () => void;
  stats: SystemStats | null;
}

export interface ChatInputProps {
  onSend: (message: string) => void;
  isLoading: boolean;
  placeholder?: string;
}

export interface SourceCitationProps {
  sources: SearchResult[];
  maxSources?: number;
}

export interface WebResearchPanelProps {
  onResearch: (query: string) => void;
  isResearching: boolean;
  results: WebResearchResponse | null;
}

export interface GitHubIndexPanelProps {
  onIndex: (repoUrl: string) => void;
  isIndexing: boolean;
  results: GitHubIndexResponse | null;
}

// API Error Types
export interface APIError {
  error: string;
  detail?: string;
  timestamp: number;
}

// WebSocket Message Types
export interface WebSocketMessage {
  type: string;
  data: any;
  timestamp: number;
}
