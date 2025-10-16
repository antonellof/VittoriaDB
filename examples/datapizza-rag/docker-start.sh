#!/bin/bash
# VittoriaDB RAG with Datapizza AI - Docker Compose Launcher
# One command to start the complete RAG stack

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║   VittoriaDB RAG Assistant with Datapizza AI               ║"
echo "║   Docker Compose Launcher                                  ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  No .env file found. Creating from .env.example...${NC}"
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${YELLOW}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "⚠️  IMPORTANT: Configure your .env file before continuing!"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        echo "Required configuration:"
        echo "  1. Set OPENAI_API_KEY=your-key-here"
        echo "     Get your key from: https://platform.openai.com/api-keys"
        echo ""
        echo "Optional configuration:"
        echo "  - Change LLM_MODEL (default: gpt-4o-mini)"
        echo "  - Use Ollama instead of OpenAI (see .env for instructions)"
        echo ""
        echo "Edit the file now:"
        echo "  nano .env   (or use your preferred editor)"
        echo -e "${NC}"
        exit 1
    else
        echo -e "${RED}❌ Error: .env.example not found!${NC}"
        exit 1
    fi
fi

# Check if OPENAI_API_KEY is set
source .env
if [ "$EMBEDDER_PROVIDER" = "openai" ] && [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${RED}❌ Error: OPENAI_API_KEY not set in .env file${NC}"
    echo -e "${YELLOW}Please edit .env and set your OpenAI API key${NC}"
    exit 1
fi

# Check if docker and docker-compose are installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
    echo "   Visit: https://docs.docker.com/get-docker/"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed.${NC}"
    echo "   Visit: https://docs.docker.com/compose/install/"
    exit 1
fi

# Use 'docker compose' if available, otherwise 'docker-compose'
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo -e "${BLUE}📦 Starting services with Docker Compose...${NC}"
echo ""

# Build all services
echo -e "${BLUE}1️⃣  Building all services (VittoriaDB, backend, frontend)...${NC}"
echo -e "${YELLOW}   This may take a few minutes on first run...${NC}"
$DOCKER_COMPOSE build --pull

echo ""
echo -e "${BLUE}2️⃣  Starting all services...${NC}"
$DOCKER_COMPOSE up -d

echo ""
echo -e "${BLUE}3️⃣  Waiting for services to be healthy...${NC}"

# Wait for services
MAX_WAIT=120
ELAPSED=0

while [ $ELAPSED -lt $MAX_WAIT ]; do
    if docker compose ps | grep -q "healthy"; then
        VITTORIA_HEALTHY=$(docker compose ps vittoriadb 2>/dev/null | grep -q "healthy" && echo "true" || echo "false")
        BACKEND_HEALTHY=$(docker compose ps backend 2>/dev/null | grep -q "healthy" && echo "true" || echo "false")
        FRONTEND_HEALTHY=$(docker compose ps frontend 2>/dev/null | grep -q "healthy" && echo "true" || echo "false")
        
        if [ "$VITTORIA_HEALTHY" = "true" ] && [ "$BACKEND_HEALTHY" = "true" ] && [ "$FRONTEND_HEALTHY" = "true" ]; then
            break
        fi
    fi
    
    echo -n "."
    sleep 2
    ELAPSED=$((ELAPSED + 2))
done

echo ""
echo ""

# Check final status
if [ $ELAPSED -ge $MAX_WAIT ]; then
    echo -e "${YELLOW}⚠️  Services are taking longer than expected to start${NC}"
    echo -e "${YELLOW}   Check logs with: $DOCKER_COMPOSE logs${NC}"
else
    echo -e "${GREEN}✅ All services are healthy!${NC}"
fi

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                 🎉 System Ready!                           ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}📍 Access Points:${NC}"
echo -e "   🖥️  Frontend:    ${GREEN}http://localhost:3000${NC}"
echo -e "   🔌 Backend API:  ${GREEN}http://localhost:8501${NC}"
echo -e "   🗄️  VittoriaDB:   ${GREEN}http://localhost:8080${NC}"
echo ""
echo -e "${BLUE}📊 Useful Commands:${NC}"
echo -e "   View logs:       ${YELLOW}$DOCKER_COMPOSE logs -f${NC}"
echo -e "   Stop services:   ${YELLOW}$DOCKER_COMPOSE down${NC}"
echo -e "   Restart:         ${YELLOW}$DOCKER_COMPOSE restart${NC}"
echo -e "   View status:     ${YELLOW}$DOCKER_COMPOSE ps${NC}"
echo ""
echo -e "${BLUE}🔧 Tech Stack:${NC}"
echo -e "   - Datapizza AI (Embeddings)"
echo -e "   - VittoriaDB (Vector Database)"
echo -e "   - FastAPI (Backend)"
echo -e "   - Next.js (Frontend)"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"

