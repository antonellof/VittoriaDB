#!/bin/bash

# VittoriaDB RAG Web UI - Docker Monitoring Script
# Real-time monitoring and health checking for all services

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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

log_header() {
    echo -e "${CYAN}$1${NC}"
}

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Determine Docker Compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

# Determine which compose file to use
COMPOSE_FILE="docker-compose.yml"
if [ -f "docker-compose.prod.yml" ] && $DOCKER_COMPOSE -f docker-compose.prod.yml ps | grep -q "Up"; then
    COMPOSE_FILE="docker-compose.prod.yml"
elif [ -f "docker-compose.dev.yml" ] && $DOCKER_COMPOSE -f docker-compose.dev.yml ps | grep -q "Up"; then
    COMPOSE_FILE="docker-compose.dev.yml"
fi

echo -e "${CYAN}🔍 VittoriaDB RAG Web UI - Docker Monitor${NC}"
echo "========================================"
log_info "Using compose file: $COMPOSE_FILE"
echo ""

# Function to check service health
check_service_health() {
    local service_name=$1
    local url=$2
    local description=$3
    
    printf "%-20s " "$description:"
    
    if curl -s "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ Unhealthy${NC}"
        return 1
    fi
}

# Function to get container stats
get_container_stats() {
    local container_name=$1
    
    if docker ps --format "table {{.Names}}" | grep -q "$container_name"; then
        local stats=$(docker stats --no-stream --format "table {{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" "$container_name" 2>/dev/null | tail -n 1)
        echo "$stats"
    else
        echo "Container not running"
    fi
}

# Main monitoring loop
monitor_services() {
    while true; do
        clear
        echo -e "${CYAN}🔍 VittoriaDB RAG Web UI - Live Monitor${NC}"
        echo "========================================"
        echo "$(date '+%Y-%m-%d %H:%M:%S') | Compose file: $COMPOSE_FILE"
        echo ""
        
        # Container Status
        log_header "📦 Container Status"
        echo "----------------------------------------"
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" ps
        echo ""
        
        # Service Health Checks
        log_header "🏥 Service Health"
        echo "----------------------------------------"
        check_service_health "vittoriadb" "http://localhost:8080/health" "VittoriaDB"
        check_service_health "backend" "http://localhost:8501/health" "Backend API"
        check_service_health "frontend" "http://localhost:3000" "Frontend"
        check_service_health "ollama" "http://localhost:11434/api/tags" "Ollama"
        check_service_health "redis" "http://localhost:6379" "Redis" || {
            # Redis doesn't have HTTP endpoint, check with redis-cli
            if docker exec $(docker ps -q -f name=redis) redis-cli ping > /dev/null 2>&1; then
                printf "\r%-20s " "Redis:"
                echo -e "${GREEN}✅ Healthy${NC}"
            fi
        }
        echo ""
        
        # Resource Usage
        log_header "📊 Resource Usage"
        echo "----------------------------------------"
        printf "%-20s %-10s %-20s %-15s\n" "Service" "CPU" "Memory" "Network I/O"
        echo "--------------------------------------------------------------------"
        
        # Get container names from compose file
        local containers=$($DOCKER_COMPOSE -f "$COMPOSE_FILE" ps --services)
        
        for service in $containers; do
            local container_name=$(docker ps --format "{{.Names}}" | grep "$service" | head -1)
            if [ -n "$container_name" ]; then
                local stats=$(get_container_stats "$container_name")
                if [ "$stats" != "Container not running" ]; then
                    printf "%-20s %s\n" "$service" "$stats"
                else
                    printf "%-20s %s\n" "$service" "Not running"
                fi
            fi
        done
        echo ""
        
        # Disk Usage
        log_header "💾 Volume Usage"
        echo "----------------------------------------"
        docker system df -v | grep -E "(VOLUME NAME|vittoriadb|ollama|redis|backend)" || true
        echo ""
        
        # Recent Logs (last 5 lines from each service)
        log_header "📋 Recent Logs"
        echo "----------------------------------------"
        for service in backend frontend vittoriadb; do
            echo -e "${YELLOW}$service:${NC}"
            $DOCKER_COMPOSE -f "$COMPOSE_FILE" logs --tail=2 "$service" 2>/dev/null | sed 's/^/  /' || echo "  No logs available"
            echo ""
        done
        
        echo "Press Ctrl+C to exit, or wait 10 seconds for refresh..."
        sleep 10
    done
}

# Function to show service URLs
show_urls() {
    echo -e "${CYAN}🌐 Service URLs${NC}"
    echo "==============="
    echo "📱 Frontend:        http://localhost:3000"
    echo "🔧 Backend API:     http://localhost:8501"
    echo "📊 API Docs:        http://localhost:8501/docs"
    echo "🗄️  VittoriaDB:      http://localhost:8080"
    echo "🤖 Ollama:          http://localhost:11434"
    echo "📊 Redis:           http://localhost:6379"
    
    if [ "$COMPOSE_FILE" = "docker-compose.dev.yml" ]; then
        echo "🔍 Redis Commander: http://localhost:8081"
    fi
    echo ""
}

# Function to show quick stats
show_quick_stats() {
    echo -e "${CYAN}📊 Quick Stats${NC}"
    echo "==============="
    
    # Container count
    local running_containers=$($DOCKER_COMPOSE -f "$COMPOSE_FILE" ps | grep -c "Up" || echo "0")
    local total_containers=$($DOCKER_COMPOSE -f "$COMPOSE_FILE" ps | wc -l | tr -d ' ')
    total_containers=$((total_containers - 1)) # Subtract header line
    
    echo "Containers: $running_containers/$total_containers running"
    
    # Health status
    local healthy_services=0
    local total_services=5
    
    curl -s http://localhost:8080/health > /dev/null 2>&1 && ((healthy_services++))
    curl -s http://localhost:8501/health > /dev/null 2>&1 && ((healthy_services++))
    curl -s http://localhost:3000 > /dev/null 2>&1 && ((healthy_services++))
    curl -s http://localhost:11434/api/tags > /dev/null 2>&1 && ((healthy_services++))
    docker exec $(docker ps -q -f name=redis) redis-cli ping > /dev/null 2>&1 && ((healthy_services++))
    
    echo "Health: $healthy_services/$total_services services healthy"
    
    # System resources
    local total_cpu=$(docker stats --no-stream --format "{{.CPUPerc}}" | grep -o '[0-9.]*' | awk '{sum += $1} END {printf "%.1f", sum}')
    echo "Total CPU: ${total_cpu}%"
    
    echo ""
}

# Parse command line arguments
case "${1:-monitor}" in
    monitor|m)
        monitor_services
        ;;
    status|s)
        show_urls
        show_quick_stats
        log_header "📦 Container Status"
        echo "----------------------------------------"
        $DOCKER_COMPOSE -f "$COMPOSE_FILE" ps
        ;;
    urls|u)
        show_urls
        ;;
    logs|l)
        service="${2:-}"
        if [ -n "$service" ]; then
            log_info "Showing logs for $service (press Ctrl+C to exit)"
            $DOCKER_COMPOSE -f "$COMPOSE_FILE" logs -f "$service"
        else
            log_info "Showing all logs (press Ctrl+C to exit)"
            $DOCKER_COMPOSE -f "$COMPOSE_FILE" logs -f
        fi
        ;;
    help|h)
        echo "Usage: $0 [COMMAND] [OPTIONS]"
        echo ""
        echo "Commands:"
        echo "  monitor, m          Start live monitoring (default)"
        echo "  status, s           Show current status"
        echo "  urls, u             Show service URLs"
        echo "  logs, l [service]   Show logs (all or specific service)"
        echo "  help, h             Show this help"
        echo ""
        echo "Examples:"
        echo "  $0                  # Start live monitoring"
        echo "  $0 status           # Show current status"
        echo "  $0 logs backend     # Show backend logs"
        ;;
    *)
        log_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
