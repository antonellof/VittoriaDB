#!/bin/bash

# VittoriaDB RAG Web UI - Enhanced Docker Setup & Launch
# Complete containerized deployment script with environment support

set -euo pipefail  # Exit on error, undefined vars, pipe failures

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${BLUE}🐳 VittoriaDB RAG Web UI - Enhanced Docker Setup${NC}"
echo "=================================================="

# Parse command line arguments
ENVIRONMENT="production"
COMPOSE_FILE=""
PULL_MODELS=true
DETACHED=true
BUILD=true

show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -e, --env ENV        Environment: development, production (default: production)"
    echo "  -f, --foreground     Run in foreground (not detached)"
    echo "  --no-build          Skip building images"
    echo "  --no-models         Skip pulling Ollama models"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                           # Production mode"
    echo "  $0 -e development            # Development mode with hot reload"
    echo "  $0 -e production -f          # Production mode in foreground"
    echo "  $0 --no-build --no-models    # Quick restart without rebuilding"
}

while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -f|--foreground)
            DETACHED=false
            shift
            ;;
        --no-build)
            BUILD=false
            shift
            ;;
        --no-models)
            PULL_MODELS=false
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validate environment
case $ENVIRONMENT in
    development|dev)
        ENVIRONMENT="development"
        COMPOSE_FILE="docker-compose.dev.yml"
        ;;
    production|prod)
        ENVIRONMENT="production"
        COMPOSE_FILE="docker-compose.prod.yml"
        ;;
    *)
        log_error "Invalid environment: $ENVIRONMENT. Use 'development' or 'production'"
        exit 1
        ;;
esac

log_info "Environment: $ENVIRONMENT"
log_info "Compose file: $COMPOSE_FILE"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed. Please install Docker first:"
    echo "   https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    log_error "Docker Compose is not available. Please install Docker Compose:"
    echo "   https://docs.docker.com/compose/install/"
    exit 1
fi

# Determine Docker Compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

log_success "Docker and Docker Compose are available"

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    log_error "Docker daemon is not running. Please start Docker first."
    exit 1
fi

# Environment file setup
ENV_FILE=".env"
ENV_TEMPLATE="env.${ENVIRONMENT}"

if [ ! -f "$ENV_FILE" ]; then
    log_info "Creating .env file from $ENV_TEMPLATE template..."
    if [ -f "$ENV_TEMPLATE" ]; then
        cp "$ENV_TEMPLATE" "$ENV_FILE"
        log_success "Created .env file from $ENV_TEMPLATE"
    else
        cp env.example "$ENV_FILE"
        log_success "Created .env file from env.example"
    fi
    
    log_warning "IMPORTANT: Please edit .env file with your configuration:"
    echo "   - Add your OpenAI API key (OPENAI_API_KEY)"
    echo "   - Optionally add GitHub token (GITHUB_TOKEN)"
    echo ""
    
    if [ "$ENVIRONMENT" = "development" ]; then
        log_info "Development mode: You can continue and update the .env file later"
    else
        read -p "Press Enter after updating .env file, or press Ctrl+C to exit and configure later..."
    fi
fi

# Create necessary directories
log_info "Creating necessary directories..."
mkdir -p data uploads logs

# Cleanup function for graceful shutdown
cleanup() {
    log_info "Shutting down containers..."
    $DOCKER_COMPOSE -f "$COMPOSE_FILE" down
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Check if compose file exists
if [ ! -f "$COMPOSE_FILE" ]; then
    log_error "Compose file $COMPOSE_FILE not found!"
    exit 1
fi

log_info "Using compose file: $COMPOSE_FILE"

# Stop any existing containers
log_info "Stopping any existing containers..."
$DOCKER_COMPOSE -f "$COMPOSE_FILE" down --remove-orphans

echo ""
log_info "🏗️  Building and starting containers..."
echo "======================================"

# Build arguments
BUILD_ARGS=""
if [ "$BUILD" = true ]; then
    BUILD_ARGS="--build"
    log_info "Building images..."
else
    log_info "Skipping image build..."
fi

# Detached arguments
DETACH_ARGS=""
if [ "$DETACHED" = true ]; then
    DETACH_ARGS="-d"
    log_info "Running in detached mode..."
else
    log_info "Running in foreground mode..."
fi

# Start containers
$DOCKER_COMPOSE -f "$COMPOSE_FILE" up $BUILD_ARGS $DETACH_ARGS

# Only wait for services if running in detached mode
if [ "$DETACHED" = true ]; then
    echo ""
    log_info "⏳ Waiting for services to start..."
    
    # Function to wait for service health
    wait_for_service() {
        local service_name=$1
        local url=$2
        local max_attempts=$3
        local sleep_time=${4:-3}
        
        log_info "🔍 Checking $service_name health..."
        for i in $(seq 1 $max_attempts); do
            if curl -s "$url" > /dev/null 2>&1; then
                log_success "$service_name is healthy"
                return 0
            fi
            
            if [ $i -eq $max_attempts ]; then
                log_error "$service_name health check failed after $max_attempts attempts"
                log_info "📋 $service_name logs:"
                $DOCKER_COMPOSE -f "$COMPOSE_FILE" logs --tail=20 "$service_name" || true
                return 1
            fi
            
            sleep $sleep_time
        done
    }
    
    # Wait for VittoriaDB
    wait_for_service "vittoriadb" "http://localhost:8080/health" 20 3
    
    # Wait for backend
    wait_for_service "backend" "http://localhost:8501/health" 30 3
    
    # Wait for frontend
    wait_for_service "frontend" "http://localhost:3000" 20 3
    
    # Pull Ollama models if requested
    if [ "$PULL_MODELS" = true ]; then
        log_info "📥 Pulling Ollama models..."
        
        # Wait for Ollama to be ready first
        if wait_for_service "ollama" "http://localhost:11434/api/tags" 20 3; then
            log_info "Pulling nomic-embed-text model..."
            $DOCKER_COMPOSE -f "$COMPOSE_FILE" exec -T ollama ollama pull nomic-embed-text || {
                log_warning "Failed to pull nomic-embed-text model. You can pull it manually later."
            }
        else
            log_warning "Ollama not ready, skipping model pull"
        fi
    fi
    
    echo ""
    log_success "🎉 VittoriaDB RAG Web UI is now running!"
    echo "============================================="
    echo ""
    echo "📱 Frontend:        http://localhost:3000"
    echo "🔧 Backend API:     http://localhost:8501"
    echo "📊 API Docs:        http://localhost:8501/docs"
    echo "🗄️  VittoriaDB:      http://localhost:8080"
    echo "🤖 Ollama:          http://localhost:11434"
    echo "📊 Redis:           http://localhost:6379"
    
    if [ "$ENVIRONMENT" = "development" ]; then
        echo "🔍 Redis Commander: http://localhost:8081"
    fi
    
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
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f              # View all logs"
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f backend      # View backend logs"
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE logs -f frontend     # View frontend logs"
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE down                 # Stop all containers"
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE down -v              # Stop and remove volumes"
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE restart              # Restart all containers"
    echo ""
    echo "🛑 To stop the application:"
    echo "   $DOCKER_COMPOSE -f $COMPOSE_FILE down"
    
    if [ "$ENVIRONMENT" = "development" ]; then
        echo ""
        log_info "Development mode features:"
        echo "• Hot reload enabled for both frontend and backend"
        echo "• Debug logging enabled"
        echo "• Source code mounted for live editing"
        echo "• Redis Commander available for debugging"
    fi
else
    log_info "Running in foreground mode. Press Ctrl+C to stop."
fi
