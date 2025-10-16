# 🚀 Quick Start Guide - VittoriaDB RAG with Datapizza AI

This guide will help you get the entire RAG stack running in minutes using Docker or manual setup.

## 📋 What's Special

The system uses **[Datapizza AI](https://datapizza.tech/en/ai-framework/)** as the unified framework for **both embeddings and LLM streaming**, following official [RAG patterns](https://docs.datapizza.ai/0.0.2/Guides/RAG/rag/) and [streaming patterns](https://docs.datapizza.ai/0.0.2/Guides/Clients/streaming/), while using **VittoriaDB** as the high-performance vector database.

### Key Features

✅ **Datapizza AI Framework** for embeddings + LLM streaming  
✅ **Multiple Providers**: OpenAI (cloud) or Ollama (local)  
✅ **VittoriaDB HNSW** for fast similarity search  
✅ **Production-Ready Patterns** from Datapizza AI docs  
✅ **Docker Compose** for one-command deployment  
✅ **Client & Server-side** embeddings support

## 🎯 System Architecture

```
┌───────────────────────────────────────────────────────────────┐
│                    Full Stack RAG                             │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │   Frontend   │───▶│   Backend    │───▶│  VittoriaDB  │   │
│  │  Next.js     │◀───│   FastAPI    │◀───│   HNSW DB    │   │
│  │  Port: 3000  │    │  Port: 8501  │    │  Port: 8080  │   │
│  └──────────────┘    └──────┬───────┘    └──────────────┘   │
│                              │                                │
│                      ┌───────▼────────┐                       │
│                      │ Datapizza AI   │                       │
│                      │   Framework    │                       │
│                      └───────┬────────┘                       │
│                   ┌──────────┴──────────┐                     │
│                   │                     │                     │
│         ┌─────────▼────────┐   ┌────────▼────────┐           │
│         │   Embeddings     │   │  LLM Streaming  │           │
│         │  OpenAIEmbedder  │   │  OpenAIClient   │           │
│         └─────────┬────────┘   └────────┬────────┘           │
│         ┌─────────┴──────────────────────┴─────┐             │
│         │                                      │             │
│  ┌──────▼──────┐              ┌──────▼──────┐  │            │
│  │   OpenAI    │              │   Ollama    │  │            │
│  │  (Cloud)    │              │  (Local)    │  │            │
│  └─────────────┘              └─────────────┘  │            │
└───────────────────────────────────────────────────────────────┘
```

## ⚙️ Setup Instructions

### Option 1: 🐳 Docker (Recommended - Easiest!)

The fastest way to get everything running:

```bash
cd examples/datapizza-rag

# 1. Copy and configure environment
cp env.docker.example .env
nano .env  # Add your OPENAI_API_KEY

# 2. Start all services (auto-builds)
chmod +x docker-start.sh
./docker-start.sh

# Or use docker-compose directly:
docker-compose up -d
```

**Wait 1-2 minutes for health checks**, then access:
- **Frontend**: http://localhost:3000
- **Backend**: http://localhost:8501/docs
- **VittoriaDB**: http://localhost:8080/docs

**Docker Commands:**
```bash
# View logs
docker-compose logs -f backend

# Check status
docker-compose ps

# Stop services
docker-compose down

# Remove all data
docker-compose down -v
```

That's it! Skip to the [Usage Examples](#📝-usage-examples) section.

---

### Option 2: 💻 Manual Installation

### 0. Install VittoriaDB (Prerequisites)

Before starting, you need to install VittoriaDB. Choose the easiest method:

#### Quick Install (Recommended)
```bash
# One-line installer - downloads pre-built binary for your platform
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash
```

This installs `vittoriadb` to `/usr/local/bin` (or `~/.local/bin` on Linux).

#### Build from Source (If you have Go installed)
```bash
# From the VittoriaDB project root
make build

# This creates ./build/vittoriadb
```

#### Download Pre-built Binary
```bash
# Visit: https://github.com/antonellof/VittoriaDB/releases/latest
# Download for your platform:
# - vittoriadb-v0.5.0-darwin-arm64 (Mac M1/M2)
# - vittoriadb-v0.5.0-darwin-amd64 (Mac Intel)
# - vittoriadb-v0.5.0-linux-amd64 (Linux)
# - vittoriadb-v0.5.0-windows-amd64.exe (Windows)

# Make it executable (Mac/Linux)
chmod +x vittoriadb-*
mv vittoriadb-* /usr/local/bin/vittoriadb
```

#### Verify Installation
```bash
vittoriadb --version
# Should show: VittoriaDB version v0.5.0
```

### 1. Configure Environment Variables

#### Backend Configuration

```bash
cd backend
cp env.example .env
```

Edit `.env` and choose your embedding provider:

**Option A: OpenAI (Recommended for Quality)**
```bash
EMBEDDER_PROVIDER=openai
OPENAI_API_KEY=sk-your-key-here
OPENAI_EMBED_MODEL=text-embedding-3-small
OPENAI_EMBED_DIMENSIONS=1536
```

**Option B: Ollama (Free & Local)**
```bash
EMBEDDER_PROVIDER=ollama
OLLAMA_BASE_URL=http://localhost:11434/v1
OLLAMA_EMBED_MODEL=nomic-embed-text
OLLAMA_EMBED_DIMENSIONS=768

# Install Ollama first:
# brew install ollama  # macOS
# ollama pull nomic-embed-text
# ollama serve
```

#### Frontend Configuration

```bash
cd ../frontend
cp env.local.example .env.local
```

Edit `.env.local`:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8501
```

### 2. Install Dependencies

#### Backend

```bash
cd backend
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
```

Key dependencies installed:
- `datapizza-ai` - Unified AI framework (embeddings + LLM streaming)
- `vittoriadb>=0.2.0` - Vector database
- `fastapi>=0.104.0` - Backend framework
- `crawl4ai>=0.7.4` - Web scraping
- `beautifulsoup4` - HTML parsing

#### Frontend

```bash
cd ../frontend
npm install
```

### 3. Start the Services

Open **3 terminal windows**:

#### Terminal 1: VittoriaDB

```bash
# If you used the installer or downloaded to PATH:
vittoriadb run --data-dir ./data --port 8080

# Or if you built from source:
cd /Users/d695663/Desktop/Dev/CognitoraVector
./build/vittoriadb run --data-dir ./data --port 8080
```

Wait for: `✅ VittoriaDB listening on :8080`

#### Terminal 2: Backend

```bash
cd examples/datapizza-rag/backend
source venv/bin/activate  # If using virtual environment
python main.py
```

Wait for:
```
✅ Datapizza embedder initialized: openai (text-embedding-3-small, 1536D)
✅ Connected to VittoriaDB at http://localhost:8080
INFO: Uvicorn running on http://0.0.0.0:8501
```

#### Terminal 3: Frontend

```bash
cd examples/datapizza-rag/frontend
npm run dev
```

Wait for: `✅ Ready on http://localhost:3000`

### 4. Access the Application

Open your browser: **http://localhost:3000**

## 🔍 Verify Everything Works

### Check VittoriaDB

```bash
curl http://localhost:8080/health
# Expected: {"status": "healthy"}
```

### Check Backend

```bash
curl http://localhost:8501/health
# Expected: {"status": "healthy", "embedder": {...}}
```

### Check Frontend

```bash
curl -I http://localhost:3000
# Expected: HTTP/1.1 200 OK
```

## 📝 Usage Examples

### 1. Upload a Document

1. Click "📄 Documents" in the sidebar
2. Click "Upload Files"
3. Select a PDF, TXT, or DOCX file
4. Wait for embedding generation (using datapizza-ai)
5. Document is now searchable!

### 2. Ask Questions

Type in the chat:
```
"What are the main topics in the uploaded documents?"
```

The system will:
- Generate query embedding via **datapizza-ai**
- Search **VittoriaDB** for relevant chunks
- Generate response with **GPT-4**
- Show sources with relevance scores

### 3. Web Research

1. Click "🌐 Web Research" in sidebar
2. Enter a topic: "Latest AI developments"
3. System will:
   - Search the web with crawl4ai
   - Extract and embed content (datapizza-ai)
   - Store in VittoriaDB
   - Answer questions about findings

### 4. GitHub Indexing

1. Click "📦 GitHub" in sidebar
2. Enter repo: `datapizza-labs/datapizza-ai`
3. System will:
   - Clone repository
   - Extract code files
   - Generate embeddings (datapizza-ai)
   - Enable code search

## 🎨 Embedding Model Comparison

### OpenAI Models

| Model | Dimensions | Speed | Cost | Quality |
|-------|-----------|-------|------|---------|
| text-embedding-3-small | 1536 | ⚡⚡⚡ | $ | ⭐⭐⭐⭐ |
| text-embedding-3-large | 3072 | ⚡⚡ | $$ | ⭐⭐⭐⭐⭐ |
| text-embedding-ada-002 | 1536 | ⚡⚡⚡ | $ | ⭐⭐⭐ |

### Ollama Models (Local)

| Model | Dimensions | Speed (CPU) | Speed (GPU) | Quality |
|-------|-----------|------------|-------------|---------|
| nomic-embed-text | 768 | ⚡ | ⚡⚡⚡ | ⭐⭐⭐ |
| mxbai-embed-large | 1024 | ⚡ | ⚡⚡ | ⭐⭐⭐⭐ |
| all-minilm | 384 | ⚡⚡ | ⚡⚡⚡ | ⭐⭐ |

**Recommendation:**
- **Production**: `text-embedding-3-small` (best balance)
- **High Quality**: `text-embedding-3-large`
- **Free/Local**: `nomic-embed-text` with GPU
- **Fast Local**: `all-minilm`

## 🐛 Troubleshooting

### Backend Won't Start

```bash
# Check if VittoriaDB is running
curl http://localhost:8080/health

# Check Python version (needs 3.8+)
python --version

# Reinstall dependencies
pip install -r requirements.txt
```

### Datapizza Import Error

```bash
pip install datapizza-ai
```

### Ollama Connection Failed

```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# Start Ollama
ollama serve

# Pull embedding model
ollama pull nomic-embed-text

# Verify
ollama list
```

### Port Already in Use

```bash
# Find and kill process using port
lsof -ti:8080 | xargs kill -9  # VittoriaDB
lsof -ti:8501 | xargs kill -9  # Backend
lsof -ti:3000 | xargs kill -9  # Frontend
```

### OpenAI API Key Invalid

```bash
# Test your API key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Make sure it's set in .env
cat backend/.env | grep OPENAI_API_KEY
```

### Dimension Mismatch Error

If you change embedding models, clear VittoriaDB data:

```bash
# Stop VittoriaDB
pkill vittoriadb

# Clear data (WARNING: Deletes all vectors!)
rm -rf data/*

# Restart VittoriaDB
vittoriadb run --data-dir ./data --port 8080
```

## 📊 Performance Tips

### Faster Embeddings

1. **Use OpenAI** (10-50x faster than local CPU)
2. **Batch uploads** (better throughput)
3. **Use GPU** for Ollama (5-10x faster)

### Faster Search

1. **HNSW indexing** (enabled by default in VittoriaDB)
2. **Adjust min_score** (lower = more results, slower)
3. **Limit results** (default: 5, max: 50)

### Memory Usage

- **OpenAI**: ~100MB (no model loading)
- **Ollama CPU**: ~2-4GB per model
- **Ollama GPU**: ~4-8GB per model

## 📚 Additional Resources

- **Datapizza AI Website**: https://datapizza.tech/en/ai-framework/
- **Datapizza AI Docs**: https://docs.datapizza.ai/
- **RAG Guide**: https://docs.datapizza.ai/0.0.2/Guides/RAG/rag/
- **Streaming Guide**: https://docs.datapizza.ai/0.0.2/Guides/Clients/streaming/
- **VittoriaDB GitHub**: https://github.com/antonellof/VittoriaDB
- **Ollama Models**: https://ollama.ai/library

## 🔧 Development

### Run with Auto-Reload

**Backend:**
```bash
uvicorn main:app --reload --port 8501
```

**Frontend:**
```bash
npm run dev
```

### Check Logs

**Backend:**
```bash
tail -f backend/backend.log
```

**Frontend:**
```bash
tail -f frontend/frontend.log
```

### API Documentation

Once running, visit:
- **Backend API Docs**: http://localhost:8501/docs
- **VittoriaDB API**: http://localhost:8080/docs

## 🎯 Next Steps

1. ✅ Start all services
2. ✅ Upload test document
3. ✅ Ask questions
4. ✅ Try web research
5. ✅ Index a GitHub repo
6. Configure embedding provider for your use case
7. Adjust chunk size and search parameters
8. Deploy to production

## 📝 Configuration Files

- `.env` (Docker) or `backend/.env` (Manual) - Main configuration
- `backend/env.example` - Template with all options
- `frontend/.env.local` - Frontend API URL
- `docker-compose.yml` - Docker orchestration
- `env.docker.example` - Docker environment template

## 🤝 Support

If you encounter issues:

1. Check this troubleshooting guide
2. Review Docker logs: `docker-compose logs -f`
3. Check Datapizza AI docs: https://docs.datapizza.ai/
4. Open an issue on GitHub

---

**Built with:**
- ⚡ [Datapizza AI](https://datapizza.tech/en/ai-framework/) - Modern AI Framework
- 🗄️ [VittoriaDB](https://github.com/antonellof/VittoriaDB) - HNSW Vector Database
- ⚛️ [Next.js](https://nextjs.org) - Frontend Framework
- 🚀 [FastAPI](https://fastapi.tiangolo.com) - Backend Framework
- 🐳 [Docker Compose](https://docs.docker.com/compose/) - Container Orchestration

