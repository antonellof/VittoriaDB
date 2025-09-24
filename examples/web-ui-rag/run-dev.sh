#!/bin/bash

# VittoriaDB RAG Web UI - Development Runner
# This script starts the development environment using Docker Compose

set -e

echo "🚀 Starting VittoriaDB RAG Web UI Development Environment"
echo "=================================================="

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  No .env file found. Creating from env.development..."
    cp env.development .env
    echo "📝 Please edit .env file with your API keys before continuing."
    echo "   Required: OPENAI_API_KEY"
    echo "   Optional: GITHUB_TOKEN"
    read -p "Press Enter to continue after editing .env file..."
fi

# Build and start services
echo "🔨 Building and starting services..."
docker-compose -f docker-compose.dev.yml up --build

echo "🎉 Development environment stopped."
