#!/bin/bash

# VittoriaDB RAG Web UI - Production Runner
# This script starts the production environment using Docker Compose

set -e

echo "🚀 Starting VittoriaDB RAG Web UI Production Environment"
echo "====================================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  No .env file found. Creating from env.production..."
    cp env.production .env
    echo "📝 Please edit .env file with your API keys before continuing."
    echo "   Required: OPENAI_API_KEY"
    echo "   Optional: GITHUB_TOKEN"
    read -p "Press Enter to continue after editing .env file..."
fi

# Build and start services
echo "🔨 Building and starting services..."
docker-compose -f docker-compose.prod.yml up --build -d

echo "🎉 Production environment started!"
echo ""
echo "📊 Services:"
echo "   Frontend: http://localhost:3000"
echo "   Backend API: http://localhost:8501"
echo "   VittoriaDB: http://localhost:8080"
echo "   Ollama: http://localhost:11434"
echo ""
echo "📝 To view logs: docker-compose -f docker-compose.prod.yml logs -f"
echo "🛑 To stop: docker-compose -f docker-compose.prod.yml down"
