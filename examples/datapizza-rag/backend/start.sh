#!/bin/bash

# VittoriaDB RAG Backend Startup Script

echo "🚀 Starting VittoriaDB RAG Backend"

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo "📦 Creating virtual environment..."
    python -m venv venv
fi

# Activate virtual environment
echo "🔧 Activating virtual environment..."
source venv/bin/activate

# Install dependencies
echo "📚 Installing dependencies..."
pip install -r requirements.txt

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "⚠️  No .env file found. Creating from example..."
    cp env.example .env
    echo "📝 Please edit .env file with your configuration"
fi

# Check if Ollama is running
echo "🔍 Checking Ollama status..."
if curl -s http://localhost:11434/api/tags > /dev/null; then
    echo "✅ Ollama is running"
else
    echo "⚠️  Ollama is not running. Starting Ollama..."
    echo "Please run: ollama serve"
    echo "Then run: ollama pull nomic-embed-text"
fi

# Start the FastAPI server
echo "🌟 Starting FastAPI server..."
uvicorn main:app --host 0.0.0.0 --port 8501 --reload
