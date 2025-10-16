# VittoriaDB RAG Assistant with Datapizza AI

A complete, production-ready RAG (Retrieval-Augmented Generation) system powered by **[Datapizza AI](https://datapizza.tech/en/ai-framework/)** for embeddings and LLM interactions, and **VittoriaDB** as the vector database.

## ✨ What's Inside

- **🍕 Datapizza AI Pipelines**: Production-ready RAG with `IngestionPipeline` & `DagPipeline` architecture
- **🧠 Datapizza AI Integration**: Modern AI framework for embeddings and LLM streaming (OpenAI & Ollama)
- **⚡ VittoriaDB**: High-performance HNSW vector database for semantic search
- **💬 Chat Interface**: Beautiful Next.js UI with real-time streaming responses
- **📁 Document Processing**: Upload and index PDFs, DOCX, TXT, MD, HTML files
- **🌐 Web Research**: Live web search with automatic knowledge storage using Crawl4AI
- **💻 GitHub Indexing**: Index and search through code repositories
- **🐳 Docker Support**: One-command deployment with Docker Compose

> **🆕 NEW**: Now using Datapizza AI's pipeline architecture for modular, production-ready RAG!  
> See [`DATAPIZZA_PIPELINE_INTEGRATION.md`](./DATAPIZZA_PIPELINE_INTEGRATION.md) for details.

## 🏗️ Architecture

```
┌───────────────────────────────────────────────────────────┐
│                    Full Stack RAG                         │
├───────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────┐  │
│  │   Frontend   │───▶│   Backend    │───▶│VittoriaDB │  │
│  │   Next.js    │    │   FastAPI    │    │  HNSW DB  │  │
│  │  Port: 3000  │◀───│  Port: 8501  │◀───│Port: 8080 │  │
│  └──────────────┘    └──────┬───────┘    └───────────┘  │
│                              │                           │
│                      ┌───────▼────────┐                  │
│                      │ Datapizza AI   │                  │
│                      │ Framework      │                  │
│                      └───────┬────────┘                  │
│                   ┌──────────┴──────────┐                │
│                   │                     │                │
│         ┌─────────▼────────┐   ┌────────▼────────┐      │
│         │   Embeddings     │   │  LLM Streaming  │      │
│         │  OpenAIEmbedder  │   │  OpenAIClient   │      │
│         └─────────┬────────┘   └────────┬────────┘      │
│         ┌─────────┴──────────────────────┴─────┐        │
│         │                                      │        │
│  ┌──────▼──────┐              ┌──────▼──────┐ │        │
│  │   OpenAI    │              │   Ollama    │ │        │
│  │  (Cloud)    │              │  (Local)    │ │        │
│  └─────────────┘              └─────────────┘ │        │
└───────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### 🐳 Docker (Recommended - One Command!)

The fastest and easiest way to run the complete stack:

```bash
cd examples/datapizza-rag

# 1. Copy environment file and add your OpenAI API key
cp env.docker.example .env
nano .env  # Set OPENAI_API_KEY=sk-your-key-here

# 2. Start everything with one command (builds all services)
chmod +x docker-start.sh
./docker-start.sh

# Or use docker-compose directly:
docker-compose up -d
```

**Wait for all services to be healthy** (~1-2 minutes), then access:
- ✨ **Frontend UI**: http://localhost:3000
- 🔧 **Backend API Docs**: http://localhost:8501/docs
- 🗄️ **VittoriaDB API**: http://localhost:8080/docs

**Useful Docker Commands:**
```bash
# View logs in real-time
docker-compose logs -f

# View specific service logs
docker-compose logs -f backend

# Check service health
docker-compose ps

# Stop services (data preserved)
docker-compose down

# Restart a single service
docker-compose restart backend

# Remove everything including data
docker-compose down -v
```

---

### 💻 Manual Installation (Advanced)

### Prerequisites

- **Go 1.21+** (to build VittoriaDB)
- **Python 3.11+**
- **Node.js 18+**
- **OpenAI API key** OR **Ollama** installed locally

### 1. Install/Build VittoriaDB

Choose one of these methods:

#### Option A: Quick Install (Recommended)
```bash
# One-line installer (downloads pre-built binary)
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash
```

#### Option B: Build from Source
```bash
# From project root
make build

# This creates ./build/vittoriadb
```

#### Option C: Download Release
```bash
# Download for your platform
# Visit: https://github.com/antonellof/VittoriaDB/releases
# Extract and place binary in your PATH or project root
```

### 2. Start VittoriaDB

```bash
# If you used the installer or downloaded release:
vittoriadb run --data-dir ./data --port 8080

# Or if you built from source:
./build/vittoriadb run --data-dir ./data --port 8080
```

Wait for: `✅ VittoriaDB listening on :8080`

### 3. Configure & Start Backend

```bash
cd examples/datapizza-rag/backend

# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Configure environment (choose one):

# Option A: OpenAI (Cloud)
cp env.example .env
# Edit .env and set:
# OPENAI_API_KEY=your_key_here
# EMBEDDER_PROVIDER=openai

# Option B: Ollama (Local)
# Edit .env and set:
# EMBEDDER_PROVIDER=ollama
# OLLAMA_BASE_URL=http://localhost:11434/v1
# OLLAMA_EMBED_MODEL=nomic-embed-text

# Start backend
python main.py
```

Backend will be available at `http://localhost:8501`

### 4. Start Frontend

```bash
cd examples/datapizza-rag/frontend

# Install dependencies (first time only)
npm install

# Start frontend
npm run dev
```

Frontend will be available at `http://localhost:3000`

## 📖 Usage

### Chat with Your Documents

1. **Upload Files**: Click "Add Data" → Upload your documents
2. **Ask Questions**: Type your questions in the chat
3. **Get Answers**: AI responds with context from your documents

### Web Research

1. Enable **"Web Search"** toggle
2. Ask about any topic
3. Results are automatically stored in your knowledge base

### Index GitHub Repositories

1. Click "Add Data" → "GitHub"
2. Enter repository URL (e.g., `https://github.com/username/repo`)
3. Code is indexed and searchable

## 🔧 Configuration

### Embedding Providers

**OpenAI (Recommended for production)**
```env
EMBEDDER_PROVIDER=openai
OPENAI_API_KEY=sk-...
OPENAI_EMBED_MODEL=text-embedding-ada-002
OPENAI_EMBED_DIMENSIONS=1536
```

**Ollama (Free, runs locally)**
```env
EMBEDDER_PROVIDER=ollama
OLLAMA_BASE_URL=http://localhost:11434/v1
OLLAMA_EMBED_MODEL=nomic-embed-text
OLLAMA_EMBED_DIMENSIONS=768
```

### LLM Configuration

```env
OPENAI_API_KEY=sk-...  # For GPT-4/GPT-3.5
LLM_MODEL=gpt-4o-mini  # Or gpt-4, gpt-3.5-turbo
```

## 🎯 Features in Detail

### Document Processing
- **Formats**: PDF, DOCX, DOC, TXT, MD, HTML
- **Automatic chunking** with configurable size and overlap
- **Metadata extraction** (filename, type, timestamp)
- **Background processing** for large files

### Web Research
- **Real-time search** using Crawl4AI
- **Automatic storage** in vector database
- **Smart content extraction** with structured data
- **Link and media tracking**

### GitHub Integration
- **Repository indexing** by URL or local path
- **Code-aware chunking** for functions and classes
- **Language detection** and filtering
- **Metadata tracking** (repo, file, language)

### RAG System
- **Semantic search** across multiple collections
- **Context-aware responses** with source citations
- **Streaming chat** for real-time feedback
- **Conversation history** with automatic saving

## 🔒 Collections

The system uses 4 main collections:

1. **documents** - Your uploaded files
2. **web_research** - Web search results
3. **github_code** - Indexed code repositories
4. **chat_history** - Conversation history

## 🐳 Docker Architecture

The Docker Compose setup includes:

**Services:**
1. **VittoriaDB** - Built from source (Golang)
2. **Backend** - FastAPI with Datapizza AI
3. **Frontend** - Next.js production build

**Volumes:**
- `vittoriadb-data` - Persistent vector storage
- `backend-data` - Uploaded files and logs

**Network:**
- All services on `datapizza-network` bridge

**Health Checks:**
- All services have health checks for reliable startup
- Automatic dependency ordering (VittoriaDB → Backend → Frontend)

## 📚 Learn More

- **Datapizza AI Framework**: [https://datapizza.tech/en/ai-framework/](https://datapizza.tech/en/ai-framework/)
- **Datapizza AI Documentation**: [https://docs.datapizza.ai/](https://docs.datapizza.ai/)
- **Datapizza RAG Guide**: [https://docs.datapizza.ai/0.0.2/Guides/RAG/rag/](https://docs.datapizza.ai/0.0.2/Guides/RAG/rag/)
- **Datapizza Streaming Guide**: [https://docs.datapizza.ai/0.0.2/Guides/Clients/streaming/](https://docs.datapizza.ai/0.0.2/Guides/Clients/streaming/)
- **VittoriaDB**: [https://github.com/antonellof/VittoriaDB](https://github.com/antonellof/VittoriaDB)

## 🎯 What Makes This Special

This example demonstrates **production-ready RAG patterns** using:
- ✅ **Datapizza AI** for unified embeddings & LLM streaming
- ✅ **VittoriaDB** HNSW for fast similarity search
- ✅ **Docker Compose** for one-command deployment
- ✅ **Client & Server-side embeddings** support
- ✅ **Streaming responses** with SSE
- ✅ **Multiple collections** for organized knowledge

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is part of VittoriaDB and follows the same license.

---

**Built with ⚡ [Datapizza AI](https://datapizza.tech/en/ai-framework/) + 🗄️ [VittoriaDB](https://github.com/antonellof/VittoriaDB)**
