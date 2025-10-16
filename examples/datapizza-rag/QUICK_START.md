# 🚀 Quick Start Guide - VittoriaDB RAG with Datapizza AI

This guide will help you get the entire application stack running in minutes.

## 📋 What We've Updated

The backend now uses **[datapizza-ai](https://github.com/datapizza-labs/datapizza-ai)** for embeddings, following the official [RAG Guide](https://docs.datapizza.ai/0.0.2/Guides/RAG/rag/) while using **VittoriaDB** as the vector database.

### Key Features

✅ **Unified Embeddings API** via datapizza-ai  
✅ **Multiple Providers**: OpenAI (cloud) or Ollama (local)  
✅ **VittoriaDB Integration** for high-performance vector storage  
✅ **Production-Ready RAG** patterns from datapizza-ai  
✅ **Client/Server Embeddings** support

## 🎯 System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Full Stack                               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐ │
│  │   Frontend   │───▶│   Backend    │───▶│  VittoriaDB  │ │
│  │  Next.js     │    │   FastAPI    │    │   Vector DB  │ │
│  │  Port: 3000  │    │  Port: 8501  │    │  Port: 8080  │ │
│  └──────────────┘    └──────────────┘    └──────────────┘ │
│                              │                              │
│                              ▼                              │
│                      ┌───────────────┐                      │
│                      │ Datapizza AI  │                      │
│                      │  Embeddings   │                      │
│                      └───────┬───────┘                      │
│                              │                              │
│                   ┌──────────┴──────────┐                   │
│                   │                     │                   │
│            ┌──────▼──────┐      ┌──────▼──────┐            │
│            │   OpenAI    │      │   Ollama    │            │
│            │  (Cloud)    │      │   (Local)   │            │
│            └─────────────┘      └─────────────┘            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## ⚙️ Setup Instructions

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
pip install -r requirements.txt
```

Key dependencies installed:
- `datapizza-ai-core>=0.0.1`
- `datapizza-ai-embedders>=0.0.1`
- `datapizza-ai-clients>=0.0.1`
- `vittoriadb>=0.2.0`
- `fastapi>=0.104.0`
- `openai>=1.3.0`

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
cd examples/web-ui-rag/frontend
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
pip install datapizza-ai-core datapizza-ai-embedders datapizza-ai-clients
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

- **Datapizza AI Docs**: https://docs.datapizza.ai/
- **RAG Guide**: https://docs.datapizza.ai/0.0.2/Guides/RAG/rag/
- **VittoriaDB Docs**: https://vittoriadb.com
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

- `backend/.env` - Backend configuration
- `backend/env.example` - Template with all options
- `frontend/.env.local` - Frontend configuration
- `backend/DATAPIZZA_INTEGRATION.md` - Detailed integration guide

## 🤝 Support

If you encounter issues:

1. Check this troubleshooting guide
2. Review `backend/DATAPIZZA_INTEGRATION.md`
3. Check datapizza-ai docs: https://docs.datapizza.ai/
4. Open an issue on GitHub

---

**Built with:**
- [Datapizza AI](https://github.com/datapizza-labs/datapizza-ai) - AI Framework
- [VittoriaDB](https://vittoriadb.com) - Vector Database
- [Next.js](https://nextjs.org) - Frontend Framework
- [FastAPI](https://fastapi.tiangolo.com) - Backend Framework

