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

### Prerequisites

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

### Backend Setup

```bash
cd backend
pip install -r requirements.txt
streamlit run app.py
```

### Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

### Access the Application

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8501
- **VittoriaDB**: http://localhost:8080

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
# Backend (.env)
OPENAI_API_KEY=your_openai_key_here
VITTORIADB_URL=http://localhost:8080
OLLAMA_URL=http://localhost:11434

# Frontend (.env.local)
NEXT_PUBLIC_API_URL=http://localhost:8501
```

## 📊 Performance

- **File Processing**: ~1-2 seconds per document
- **Web Research**: ~3-5 seconds per query
- **Vector Search**: <100ms response time
- **Chat Response**: ~1-3 seconds with streaming

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

MIT License - see [LICENSE](../../LICENSE) file for details.
