#!/bin/bash

# VittoriaDB RAG Web UI - Docker Setup & Launch
# Complete containerized deployment script

echo "🐳 VittoriaDB RAG Web UI - Docker Setup"
echo "======================================="

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first:"
    echo "   https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "❌ Docker Compose is not available. Please install Docker Compose:"
    echo "   https://docs.docker.com/compose/install/"
    exit 1
fi

# Determine Docker Compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

echo "✅ Docker and Docker Compose are available"

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "📝 Creating .env file from template..."
    cp env.example .env
    echo ""
    echo "⚠️  IMPORTANT: Please edit .env file with your configuration:"
    echo "   - Add your OpenAI API key (OPENAI_API_KEY)"
    echo "   - Optionally add GitHub token (GITHUB_TOKEN)"
    echo ""
    read -p "Press Enter after updating .env file, or press Ctrl+C to exit and configure later..."
fi

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p data uploads

# Check if Ollama should run in Docker or use host
echo ""
echo "🤖 Ollama Configuration:"
echo "1. Use existing Ollama on host (recommended if already installed)"
echo "2. Run Ollama in Docker container (slower, but isolated)"
echo ""
read -p "Choose option (1 or 2, default: 1): " ollama_choice
ollama_choice=${ollama_choice:-1}

if [ "$ollama_choice" = "1" ]; then
    echo "🔍 Checking host Ollama..."
    if curl -s http://localhost:11434/api/tags > /dev/null; then
        echo "✅ Ollama is running on host"
        
        # Check for nomic-embed-text model
        if curl -s http://localhost:11434/api/tags | grep -q "nomic-embed-text"; then
            echo "✅ nomic-embed-text model is available"
        else
            echo "📥 Pulling nomic-embed-text model..."
            ollama pull nomic-embed-text
        fi
    else
        echo "⚠️  Ollama is not running on host. Please start it:"
        echo "   ollama serve"
        echo "   ollama pull nomic-embed-text"
        echo ""
        read -p "Press Enter after starting Ollama, or press Ctrl+C to exit..."
    fi
    
    # Use host Ollama
    export OLLAMA_URL="http://host.docker.internal:11434"
else
    echo "🐳 Will run Ollama in Docker container..."
    
    # Enable Ollama service in docker-compose
    sed -i.bak 's/# ollama:/ollama:/' docker-compose.yml
    sed -i.bak 's/#   image: ollama/  image: ollama/' docker-compose.yml
    sed -i.bak 's/#   ports:/  ports:/' docker-compose.yml
    sed -i.bak 's/#     - "11434:11434"/    - "11434:11434"/' docker-compose.yml
    sed -i.bak 's/#   volumes:/  volumes:/' docker-compose.yml
    sed -i.bak 's/#     - ollama_data:/    - ollama_data:/' docker-compose.yml
    sed -i.bak 's/#   networks:/  networks:/' docker-compose.yml
    sed -i.bak 's/#     - vittoriadb-network/    - vittoriadb-network/' docker-compose.yml
    sed -i.bak 's/#   restart:/  restart:/' docker-compose.yml
    sed -i.bak 's/#   environment:/  environment:/' docker-compose.yml
    sed -i.bak 's/#     - OLLAMA_HOST=/    - OLLAMA_HOST=/' docker-compose.yml
fi

echo ""
echo "🏗️  Building and starting containers..."
echo "======================================"

# Build and start containers
$DOCKER_COMPOSE up --build -d

echo ""
echo "⏳ Waiting for services to start..."

# Wait for backend health check
echo "🔍 Checking backend health..."
for i in {1..30}; do
    if curl -s http://localhost:8501/health > /dev/null; then
        echo "✅ Backend is healthy"
        break
    fi
    
    if [ $i -eq 30 ]; then
        echo "❌ Backend health check failed"
        echo "📋 Backend logs:"
        $DOCKER_COMPOSE logs backend
        exit 1
    fi
    
    sleep 2
done

# Wait for frontend
echo "🔍 Checking frontend..."
for i in {1..20}; do
    if curl -s http://localhost:3000 > /dev/null; then
        echo "✅ Frontend is ready"
        break
    fi
    
    if [ $i -eq 20 ]; then
        echo "❌ Frontend check failed"
        echo "📋 Frontend logs:"
        $DOCKER_COMPOSE logs frontend
        exit 1
    fi
    
    sleep 3
done

# If using Docker Ollama, pull the model
if [ "$ollama_choice" = "2" ]; then
    echo "📥 Pulling nomic-embed-text model in Ollama container..."
    $DOCKER_COMPOSE exec ollama ollama pull nomic-embed-text
fi

echo ""
echo "🎉 VittoriaDB RAG Web UI is now running in Docker!"
echo "================================================="
echo ""
echo "📱 Frontend:     http://localhost:3000"
echo "🔧 Backend API:  http://localhost:8501"
echo "📊 API Docs:     http://localhost:8501/docs"
echo "🤖 Ollama:       http://localhost:11434 (if using Docker Ollama)"
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
echo "📋 Useful Docker commands:"
echo "   $DOCKER_COMPOSE logs -f                    # View all logs"
echo "   $DOCKER_COMPOSE logs -f backend            # View backend logs"
echo "   $DOCKER_COMPOSE logs -f frontend           # View frontend logs"
echo "   $DOCKER_COMPOSE down                       # Stop all containers"
echo "   $DOCKER_COMPOSE down -v                    # Stop and remove volumes"
echo "   $DOCKER_COMPOSE restart                    # Restart all containers"
echo ""
echo "🛑 To stop the application:"
echo "   $DOCKER_COMPOSE down"
