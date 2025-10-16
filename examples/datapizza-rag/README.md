# VittoriaDB RAG Assistant with Datapizza AI

A complete, production-ready RAG (Retrieval-Augmented Generation) system powered by **[Datapizza AI](https://datapizza.ai)** embeddings and **VittoriaDB** vector database.

## ✨ What's Inside

- **🧠 Datapizza AI Integration**: Modern embeddings API supporting OpenAI and local Ollama
- **⚡ VittoriaDB**: High-performance vector database for semantic search
- **💬 Chat Interface**: Beautiful Next.js UI with real-time streaming
- **📁 Document Processing**: Upload PDFs, DOCX, TXT, MD, HTML files
- **🌐 Web Research**: Live web search with automatic knowledge storage
- **💻 GitHub Indexing**: Index and search through code repositories

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────┐
│                     Full Stack                           │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────┐ │
│  │   Frontend   │───▶│   Backend    │───▶│VittoriaDB │ │
│  │   Next.js    │    │   FastAPI    │    │ Vector DB │ │
│  │  Port: 3000  │    │  Port: 8501  │    │Port: 8080 │ │
│  └──────────────┘    └──────┬───────┘    └───────────┘ │
│                              │                          │
│                      ┌───────▼────────┐                 │
│                      │ Datapizza AI   │                 │
│                      │   Embeddings   │                 │
│                      └───────┬────────┘                 │
│                   ┌──────────┴──────────┐               │
│            ┌──────▼──────┐      ┌───────▼──────┐       │
│            │   OpenAI    │      │   Ollama     │       │
│            │  (Cloud)    │      │  (Local)     │       │
│            └─────────────┘      └──────────────┘       │
└──────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Python 3.11+
- Node.js 18+
- OpenAI API key OR Ollama installed locally

### 1. Start VittoriaDB

```bash
# From project root
./build/vittoriadb run --data-dir ./data --port 8080
```

### 2. Configure & Start Backend

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

### 3. Start Frontend

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

## 🐳 Docker Support

```bash
# Build and run with Docker Compose
cd examples/datapizza-rag
docker-compose up -d
```

Services:
- Frontend: `http://localhost:3000`
- Backend: `http://localhost:8501`
- VittoriaDB: `http://localhost:8080`

## 📚 Learn More

- **Datapizza AI**: [https://datapizza.ai](https://datapizza.ai)
- **Datapizza RAG Guide**: [https://docs.datapizza.ai/Guides/RAG/rag/](https://docs.datapizza.ai/Guides/RAG/rag/)
- **VittoriaDB**: High-performance vector database for AI applications

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is part of VittoriaDB and follows the same license.

---

**Built with ❤️ using Datapizza AI and VittoriaDB**
