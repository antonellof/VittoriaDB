#!/bin/bash

# VittoriaDB RAG Web UI Startup Script
# Complete setup and launch script for development

echo "🚀 VittoriaDB RAG Web UI Setup & Launch"
echo "========================================"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from template..."
    cp env.example .env
    echo "⚠️  Please edit .env file with your configuration (especially OPENAI_API_KEY)"
    echo "   You can continue with the setup and add the API key later."
    echo ""
fi

# Function to check if a service is running
check_service() {
    local url=$1
    local name=$2
    if curl -s "$url" > /dev/null 2>&1; then
        echo "✅ $name is running"
        return 0
    else
        echo "❌ $name is not running"
        return 1
    fi
}

# Check Ollama
echo "🔍 Checking Ollama status..."
if check_service "http://localhost:11434/api/tags" "Ollama"; then
    echo "🤖 Checking if nomic-embed-text model is available..."
    if curl -s http://localhost:11434/api/tags | grep -q "nomic-embed-text"; then
        echo "✅ nomic-embed-text model is available"
    else
        echo "📥 Pulling nomic-embed-text model..."
        ollama pull nomic-embed-text
    fi
else
    echo "⚠️  Ollama is not running. Please start it with:"
    echo "   curl -fsSL https://ollama.ai/install.sh | sh"
    echo "   ollama serve"
    echo "   ollama pull nomic-embed-text"
    echo ""
fi

echo ""
echo "🏗️  Setting up Backend..."
echo "========================"

# Setup backend
cd backend

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    echo "📦 Creating Python virtual environment..."
    python -m venv venv
fi

# Activate virtual environment
echo "🔧 Activating virtual environment..."
source venv/bin/activate

# Install Python dependencies
echo "📚 Installing Python dependencies..."
pip install -r requirements.txt

# Copy environment file
if [ ! -f ".env" ]; then
    cp env.example .env
fi

# Start backend in background
echo "🚀 Starting FastAPI backend..."
uvicorn main:app --host 0.0.0.0 --port 8501 --reload &
BACKEND_PID=$!

# Wait for backend to start
echo "⏳ Waiting for backend to start..."
sleep 5

if check_service "http://localhost:8501/health" "FastAPI Backend"; then
    echo "✅ Backend started successfully"
else
    echo "❌ Backend failed to start"
    exit 1
fi

cd ..

echo ""
echo "🎨 Setting up Frontend..."
echo "========================"

# Setup frontend
cd frontend

# Install Node.js dependencies
echo "📦 Installing Node.js dependencies..."
npm install

# Copy environment file
if [ ! -f ".env.local" ]; then
    cp env.local.example .env.local
fi

# Start frontend
echo "🚀 Starting Next.js frontend..."
npm run dev &
FRONTEND_PID=$!

# Wait for frontend to start
echo "⏳ Waiting for frontend to start..."
sleep 10

if check_service "http://localhost:3000" "Next.js Frontend"; then
    echo "✅ Frontend started successfully"
else
    echo "❌ Frontend failed to start"
    kill $BACKEND_PID 2>/dev/null
    exit 1
fi

echo ""
echo "🎉 VittoriaDB RAG Web UI is now running!"
echo "========================================"
echo ""
echo "📱 Frontend:  http://localhost:3000"
echo "🔧 Backend:   http://localhost:8501"
echo "📊 API Docs:  http://localhost:8501/docs"
echo ""
echo "🔑 Next Steps:"
echo "1. Open http://localhost:3000 in your browser"
echo "2. Click Settings in the sidebar to configure your OpenAI API key"
echo "3. Start uploading documents and asking questions!"
echo ""
echo "💡 Features to try:"
echo "• Drag & drop files into the chat area"
echo "• Toggle 'Web Research' for real-time web searches"
echo "• Ask questions about your uploaded documents"
echo "• Index GitHub repositories for code search"
echo ""
echo "🛑 To stop the servers:"
echo "   kill $BACKEND_PID $FRONTEND_PID"
echo "   Or press Ctrl+C to stop this script"

# Keep script running and handle cleanup
cleanup() {
    echo ""
    echo "🛑 Shutting down servers..."
    kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
    echo "✅ Servers stopped"
    exit 0
}

trap cleanup SIGINT SIGTERM

# Wait for user to stop
echo ""
echo "Press Ctrl+C to stop all servers..."
wait
