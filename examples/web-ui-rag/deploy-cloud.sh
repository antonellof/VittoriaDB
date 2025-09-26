#!/bin/bash

# VittoriaDB RAG Web UI - Cloud Deployment Script
# Deploys the application using pre-built images from GitHub Container Registry

set -e

echo "🚀 VittoriaDB RAG Web UI - Cloud Deployment"
echo "============================================"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check environment variables
echo "🔍 Checking required environment variables..."

if [ -z "$OPENAI_API_KEY" ]; then
    echo "❌ OPENAI_API_KEY environment variable is not set"
    echo "   Please set it with: export OPENAI_API_KEY=your_api_key_here"
    MISSING_VARS=true
fi

if [ -z "$OLLAMA_URL" ]; then
    echo "⚠️  OLLAMA_URL not set, using default: http://ollama:11434"
    export OLLAMA_URL="http://ollama:11434"
fi

if [ -z "$GITHUB_TOKEN" ]; then
    echo "⚠️  GITHUB_TOKEN not set (optional for private repos)"
fi

if [ "$MISSING_VARS" = true ]; then
    echo ""
    echo "💡 Set environment variables and try again:"
    echo "   export OPENAI_API_KEY=your_openai_api_key"
    echo "   export GITHUB_TOKEN=your_github_token  # optional"
    echo "   export OLLAMA_URL=http://ollama:11434   # optional"
    exit 1
fi

echo "✅ Environment variables configured"

# Pull the latest VittoriaDB image
echo "📥 Pulling VittoriaDB v0.5.0 from GitHub Container Registry..."
docker pull ghcr.io/antonellof/vittoriadb:v0.5.0

# Check if we should use the cloud configuration
COMPOSE_FILE="docker-compose.cloud.yml"
if [ "$1" = "--local" ]; then
    COMPOSE_FILE="docker-compose.yml"
    echo "🏠 Using local build configuration"
else
    echo "☁️  Using cloud deployment configuration with ghcr.io image"
fi

# Start services
echo "🔨 Starting services with $COMPOSE_FILE..."
docker-compose -f $COMPOSE_FILE up -d

echo ""
echo "🎉 VittoriaDB RAG Web UI is starting up!"
echo ""
echo "📊 Service URLs:"
echo "   • Web UI:     http://localhost:3000"
echo "   • Backend:    http://localhost:8501"
echo "   • VittoriaDB: http://localhost:8080"
echo "   • Ollama:     http://localhost:11434"
echo ""
echo "📋 Useful commands:"
echo "   • View logs:    docker-compose -f $COMPOSE_FILE logs -f"
echo "   • Stop:         docker-compose -f $COMPOSE_FILE down"
echo "   • Restart:      docker-compose -f $COMPOSE_FILE restart"
echo "   • Status:       docker-compose -f $COMPOSE_FILE ps"
echo ""
echo "⏳ Services are starting up... This may take a few minutes."
echo "   Check status with: docker-compose -f $COMPOSE_FILE ps"
