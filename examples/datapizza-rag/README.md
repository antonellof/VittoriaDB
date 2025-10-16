# VittoriaDB RAG Web UI

A complete ChatGPT-like web interface built with React + Streamlit, powered by VittoriaDB for advanced RAG capabilities.

## 🎯 Features

- **💬 ChatGPT-like Interface**: Clean, modern chat UI with streaming responses
- **📁 File Upload & Processing**: Support for PDF, DOCX, TXT, MD, HTML files
- **🌐 Web Research**: Real-time web search with automatic knowledge storage
- **👨‍💻 GitHub Code Indexing**: Index and search through GitHub repositories
- **🧠 RAG System**: Intelligent document retrieval and context-aware responses
- **⚡ Real-time Streaming**: Live response streaming for better UX
- **🎨 Modern UI**: Built with React, shadcn/ui, and Tailwind CSS

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ React Frontend (Port 3000)                                 │
│ ├─ shadcn/ui components                                     │
│ ├─ Tailwind CSS styling                                     │
│ └─ Real-time chat interface                                 │
└─────────────────┬───────────────────────────────────────────┘
                  │ HTTP/WebSocket
┌─────────────────▼───────────────────────────────────────────┐
│ Streamlit Backend (Port 8501)                              │
│ ├─ File processing & upload                                 │
│ ├─ Web research integration                                 │
│ ├─ GitHub repository indexing                               │
│ └─ LLM integration (OpenAI/Ollama)                          │
└─────────────────┬───────────────────────────────────────────┘
                  │ Python SDK
┌─────────────────▼───────────────────────────────────────────┐
│ VittoriaDB Server (Port 8080)                              │
│ ├─ Vector storage & search                                  │
│ ├─ Ollama embeddings (768D)                                │
│ ├─ Document collections                                     │
│ └─ Semantic similarity search                               │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### 🐳 Docker Setup (Recommended)

The easiest way to run the complete RAG system with all dependencies:

```bash
# 1. Clone and navigate to the web UI directory
cd examples/web-ui-rag

# 2. Set up environment variables (secure approach)
export OPENAI_API_KEY=your_openai_api_key_here
export GITHUB_TOKEN=your_github_token_here  # optional
export OLLAMA_URL=http://ollama:11434        # optional

# 3. Start the development environment
./run-dev.sh
```

**What's included:**
- ✅ **VittoriaDB**: Vector database with HNSW indexing and I/O optimization (built locally)
- ✅ **Backend**: FastAPI with RAG, web research, and file processing
- ✅ **Frontend**: React UI with real-time chat interface
- ✅ **Ollama**: Local LLM inference (optional, for offline usage)
- ✅ **Chromium**: Web scraping with Playwright/Crawl4AI (fully configured)
- ✅ **Docker Compose**: Complete orchestration with health checks
- ✅ **No Redis**: Simplified architecture using FastAPI BackgroundTasks
- ✅ **Unified Configuration**: Advanced configuration management with environment variables

### 📋 Environment Configuration

**🔐 Secure Environment Variable Setup:**

Set your API keys as system environment variables (recommended for security):

```bash
# Required for AI functionality
export OPENAI_API_KEY=your_openai_api_key_here

# Optional but recommended
export GITHUB_TOKEN=your_github_token_here    # For private repos and higher rate limits
export OLLAMA_URL=http://ollama:11434          # For local ML models

# Use the interactive setup script
./setup-env.sh

# Service URLs (default values work for Docker Compose)
VITTORIADB_URL=http://localhost:8080
OLLAMA_URL=http://localhost:11434
NEXT_PUBLIC_API_URL=http://localhost:8501
```

**Docker Compose Files:**
- `docker-compose.dev.yml` - Development with hot reload
- `docker-compose.yml` - Standard setup
- `docker-compose.prod.yml` - Production optimized

### 🎯 Access Points

Once running, access the application at:

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8501
- **VittoriaDB**: http://localhost:8080
- **Ollama**: http://localhost:11434

### 🛠️ Manual Setup (Alternative)

If you prefer to run without Docker:

#### Prerequisites

```bash
# Install VittoriaDB Python SDK
pip install vittoriadb

# Install Ollama for local embeddings
curl -fsSL https://ollama.ai/install.sh | sh
ollama serve
ollama pull nomic-embed-text

# Install Node.js and npm
# https://nodejs.org/
```

#### Backend Setup

```bash
cd backend
pip install -r requirements.txt
streamlit run app.py
```

#### Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

## 📖 Usage

1. **Start VittoriaDB**: The backend automatically starts VittoriaDB server
2. **Upload Files**: Drag & drop documents for automatic processing and indexing
3. **Web Research**: Ask questions that trigger web searches with automatic storage
4. **GitHub Indexing**: Provide GitHub repo URLs for code indexing
5. **Chat**: Ask questions and get context-aware responses from your knowledge base

## 🛠️ Development

### Backend Structure
```
backend/
├── app.py              # Main Streamlit application
├── rag_system.py       # RAG logic and VittoriaDB integration
├── file_processor.py   # Document processing utilities
├── web_research.py     # Web search and scraping
├── github_indexer.py   # GitHub repository indexing
└── requirements.txt    # Python dependencies
```

### Frontend Structure
```
frontend/
├── src/
│   ├── components/     # React components
│   ├── lib/           # Utilities and API clients
│   ├── hooks/         # Custom React hooks
│   └── styles/        # Tailwind CSS styles
├── package.json       # Node.js dependencies
└── tailwind.config.js # Tailwind configuration
```

## 🎨 UI Components

- **Chat Interface**: Message bubbles, typing indicators, streaming text
- **File Upload**: Drag & drop zone with progress indicators
- **Sidebar**: Knowledge base management, collection browser
- **Settings**: Model selection, embedding configuration
- **Research Panel**: Web search results and source citations

## 🔧 Configuration

### Environment Variables

```bash
# Set environment variables (secure approach)
export OPENAI_API_KEY=your_openai_key_here
export GITHUB_TOKEN=your_github_token_here  # optional
export VITTORIADB_URL=http://localhost:8080
export OLLAMA_URL=http://localhost:11434

# Frontend environment (if running separately)
export NEXT_PUBLIC_API_URL=http://localhost:8501
```

## 🐳 Docker Architecture

The Docker setup includes several optimizations and fixes:

### ✅ **What's Fixed:**
- **Chromium Browser**: Fully configured with Playwright for web scraping
- **VittoriaDB Build**: Local build instead of non-existent external image
- **Redis Removed**: Simplified architecture using FastAPI BackgroundTasks
- **Vectorizer Config**: Proper collection initialization with embeddings
- **Build Optimization**: Reduced Docker context from 3.74GB → 438KB
- **Go Version**: Updated to 1.24 to match project requirements

### 🏗️ **Service Architecture:**
```
┌─────────────────────────────────────────────────────────────┐
│ Frontend (React + Next.js)                    Port 3000    │
└─────────────────┬───────────────────────────────────────────┘
                  │ HTTP/WebSocket
┌─────────────────▼───────────────────────────────────────────┐
│ Backend (FastAPI + Streamlit)                 Port 8501    │
│ ├─ RAG System with VittoriaDB integration                  │
│ ├─ Web Research with Chromium/Crawl4AI                     │
│ ├─ File Processing (PDF, DOCX, TXT, MD, HTML)              │
│ └─ GitHub Repository Indexing                              │
└─────────────────┬───────────────────────────────────────────┘
                  │ Python SDK
┌─────────────────▼───────────────────────────────────────────┐
│ VittoriaDB (Local Build)                      Port 8080    │
│ ├─ Vector Storage with HNSW Indexing                       │
│ ├─ OpenAI/Ollama Embeddings Integration                    │
│ └─ ACID-compliant Persistence                              │
└─────────────────────────────────────────────────────────────┘
```

## 📊 Performance

- **File Processing**: ~1-2 seconds per document
- **Web Research**: ~20-30 seconds per query (with Chromium rendering)
- **Vector Search**: <100ms response time
- **Chat Response**: ~1-3 seconds with streaming
- **Docker Build**: ~2-3 minutes (with caching)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

MIT License - see [LICENSE](../../LICENSE) file for details.
