#!/bin/bash

# VittoriaDB RAG Web UI - Development Runner
# Simple script to start the development environment

set -e

echo "🚀 VittoriaDB RAG Web UI - Development Environment"
echo "================================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check required environment variables
echo "🔍 Checking environment variables..."
if [ -z "$OPENAI_API_KEY" ]; then
    echo "❌ OPENAI_API_KEY environment variable is required"
    echo "   Set it with: export OPENAI_API_KEY=your_api_key_here"
    exit 1
fi

if [ -z "$OLLAMA_URL" ]; then
    export OLLAMA_URL="http://ollama:11434"
    echo "🔧 OLLAMA_URL set to default: $OLLAMA_URL"
fi

if [ -z "$GITHUB_TOKEN" ]; then
    echo "⚠️  GITHUB_TOKEN not set (optional for private repos)"
else
    echo "✅ GITHUB_TOKEN is set"
fi

echo "✅ Environment variables configured"

# Build and start services
echo "🔨 Building and starting services..."
docker-compose -f docker-compose.dev.yml up --build

echo "🎉 Development environment stopped."
