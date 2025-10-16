#!/bin/bash

# VittoriaDB RAG Backend Startup Script
echo "🚀 Starting VittoriaDB RAG Backend..."

# Check if we're in the right directory
if [ ! -f "main.py" ]; then
    echo "❌ Error: main.py not found. Please run this script from the backend directory."
    exit 1
fi

# Activate virtual environment if it exists
if [ -d "venv" ]; then
    echo "📦 Activating virtual environment..."
    source venv/bin/activate
else
    echo "⚠️  Warning: No virtual environment found. Using system Python."
fi

# Set OpenAI API key (uncomment and add your key)
# export OPENAI_API_KEY="your_openai_api_key_here"

# Kill any existing backend processes
echo "🔄 Stopping any existing backend processes..."
pkill -f "uvicorn.*main:app" 2>/dev/null || true
pkill -f "python.*main" 2>/dev/null || true
sleep 2

# Check if port 8501 is available
if lsof -Pi :8501 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo "⚠️  Port 8501 is still in use. Trying to free it..."
    lsof -ti:8501 | xargs kill -9 2>/dev/null || true
    sleep 3
fi

# Start the backend
echo "🌟 Starting FastAPI backend on http://localhost:8501"
echo "📊 Health check: http://localhost:8501/health"
echo "📚 API docs: http://localhost:8501/docs"
echo ""
echo "Press Ctrl+C to stop the server"
echo "================================"

# Start uvicorn with proper logging
uvicorn main:app \
    --host 0.0.0.0 \
    --port 8501 \
    --reload \
    --log-level info \
    --access-log
