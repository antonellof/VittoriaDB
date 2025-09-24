#!/bin/bash

# VittoriaDB RAG Web UI - Docker Utilities
# Utility functions for managing the Docker environment

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
get_compose_file() {
    local env="${1:-production}"
    case $env in
        development|dev)
            echo "docker-compose.dev.yml"
            ;;
        production|prod)
            echo "docker-compose.prod.yml"
            ;;
        *)
            echo "docker-compose.yml"
            ;;
    esac
}

# Function to clean up Docker resources
cleanup_docker() {
    local force="${1:-false}"
    
    echo -e "${CYAN}🧹 Docker Cleanup${NC}"
    echo "=================="
    
    if [ "$force" = "true" ]; then
        log_warning "Force cleanup mode - this will remove ALL Docker resources"
        read -p "Are you sure? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Cleanup cancelled"
            return 0
        fi
    fi
    
    # Stop all containers
    log_info "Stopping all VittoriaDB containers..."
    for compose_file in docker-compose.yml docker-compose.dev.yml docker-compose.prod.yml; do
        if [ -f "$compose_file" ]; then
            $DOCKER_COMPOSE -f "$compose_file" down --remove-orphans 2>/dev/null || true
        fi
    done
    
    # Remove containers
    log_info "Removing VittoriaDB containers..."
    docker ps -a --format "{{.Names}}" | grep -E "(vittoriadb|rag)" | xargs -r docker rm -f 2>/dev/null || true
    
    if [ "$force" = "true" ]; then
        # Remove volumes
        log_info "Removing VittoriaDB volumes..."
        docker volume ls --format "{{.Name}}" | grep -E "(vittoriadb|ollama|redis|backend)" | xargs -r docker volume rm 2>/dev/null || true
        
        # Remove networks
        log_info "Removing VittoriaDB networks..."
        docker network ls --format "{{.Name}}" | grep -E "(vittoriadb)" | xargs -r docker network rm 2>/dev/null || true
        
        # Remove images
        log_info "Removing VittoriaDB images..."
        docker images --format "{{.Repository}}:{{.Tag}}" | grep -E "(vittoriadb|rag)" | xargs -r docker rmi -f 2>/dev/null || true
        
        # System cleanup
        log_info "Running Docker system cleanup..."
        docker system prune -f --volumes
    else
        # Gentle cleanup
        log_info "Running gentle Docker cleanup..."
        docker system prune -f
    fi
    
    log_success "Cleanup completed"
}

# Function to backup data
backup_data() {
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    
    echo -e "${CYAN}💾 Data Backup${NC}"
    echo "==============="
    
    log_info "Creating backup directory: $backup_dir"
    mkdir -p "$backup_dir"
    
    # Backup volumes
    log_info "Backing up Docker volumes..."
    
    local volumes=$(docker volume ls --format "{{.Name}}" | grep -E "(vittoriadb|ollama|redis|backend)")
    
    for volume in $volumes; do
        log_info "Backing up volume: $volume"
        docker run --rm -v "$volume:/data" -v "$PWD/$backup_dir:/backup" alpine tar czf "/backup/${volume}.tar.gz" -C /data . 2>/dev/null || {
            log_warning "Failed to backup volume: $volume"
        }
    done
    
    # Backup configuration files
    log_info "Backing up configuration files..."
    cp -r *.yml *.env* *.sh "$backup_dir/" 2>/dev/null || true
    
    # Backup uploads directory
    if [ -d "uploads" ]; then
        log_info "Backing up uploads directory..."
        tar czf "$backup_dir/uploads.tar.gz" uploads/
    fi
    
    log_success "Backup completed: $backup_dir"
    
    # Show backup size
    local backup_size=$(du -sh "$backup_dir" | cut -f1)
    log_info "Backup size: $backup_size"
}

# Function to restore data
restore_data() {
    local backup_dir="$1"
    
    if [ ! -d "$backup_dir" ]; then
        log_error "Backup directory not found: $backup_dir"
        return 1
    fi
    
    echo -e "${CYAN}📥 Data Restore${NC}"
    echo "================"
    
    log_warning "This will restore data from: $backup_dir"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Restore cancelled"
        return 0
    fi
    
    # Stop services
    log_info "Stopping services..."
    for compose_file in docker-compose.yml docker-compose.dev.yml docker-compose.prod.yml; do
        if [ -f "$compose_file" ]; then
            $DOCKER_COMPOSE -f "$compose_file" down 2>/dev/null || true
        fi
    done
    
    # Restore volumes
    log_info "Restoring Docker volumes..."
    
    for backup_file in "$backup_dir"/*.tar.gz; do
        if [ -f "$backup_file" ]; then
            local volume_name=$(basename "$backup_file" .tar.gz)
            
            # Skip non-volume backups
            if [[ ! "$volume_name" =~ (vittoriadb|ollama|redis|backend) ]]; then
                continue
            fi
            
            log_info "Restoring volume: $volume_name"
            
            # Create volume if it doesn't exist
            docker volume create "$volume_name" >/dev/null 2>&1 || true
            
            # Restore data
            docker run --rm -v "$volume_name:/data" -v "$backup_dir:/backup" alpine sh -c "cd /data && tar xzf /backup/${volume_name}.tar.gz" || {
                log_warning "Failed to restore volume: $volume_name"
            }
        fi
    done
    
    # Restore uploads directory
    if [ -f "$backup_dir/uploads.tar.gz" ]; then
        log_info "Restoring uploads directory..."
        tar xzf "$backup_dir/uploads.tar.gz" || {
            log_warning "Failed to restore uploads directory"
        }
    fi
    
    log_success "Restore completed"
}

# Function to update images
update_images() {
    local env="${1:-production}"
    local compose_file=$(get_compose_file "$env")
    
    echo -e "${CYAN}🔄 Update Images${NC}"
    echo "================"
    
    log_info "Updating images for environment: $env"
    log_info "Using compose file: $compose_file"
    
    # Pull latest images
    log_info "Pulling latest images..."
    $DOCKER_COMPOSE -f "$compose_file" pull
    
    # Rebuild custom images
    log_info "Rebuilding custom images..."
    $DOCKER_COMPOSE -f "$compose_file" build --no-cache
    
    log_success "Images updated"
    
    log_info "Restart services to use updated images:"
    echo "  ./docker-start.sh -e $env"
}

# Function to reset environment
reset_environment() {
    local env="${1:-production}"
    local compose_file=$(get_compose_file "$env")
    
    echo -e "${CYAN}🔄 Reset Environment${NC}"
    echo "==================="
    
    log_warning "This will reset the $env environment"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Reset cancelled"
        return 0
    fi
    
    # Stop services
    log_info "Stopping services..."
    $DOCKER_COMPOSE -f "$compose_file" down -v --remove-orphans
    
    # Remove environment-specific volumes
    log_info "Removing volumes..."
    local env_suffix=""
    if [ "$env" = "development" ]; then
        env_suffix="_dev"
    elif [ "$env" = "production" ]; then
        env_suffix="_prod"
    fi
    
    docker volume ls --format "{{.Name}}" | grep -E "(vittoriadb|ollama|redis|backend).*${env_suffix}" | xargs -r docker volume rm 2>/dev/null || true
    
    # Rebuild and restart
    log_info "Rebuilding and restarting..."
    $DOCKER_COMPOSE -f "$compose_file" up --build -d
    
    log_success "Environment reset completed"
}

# Function to show help
show_help() {
    echo "VittoriaDB RAG Web UI - Docker Utilities"
    echo "========================================"
    echo ""
    echo "Usage: $0 COMMAND [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  cleanup [--force]           Clean up Docker resources"
    echo "  backup                      Backup data and configuration"
    echo "  restore BACKUP_DIR          Restore from backup"
    echo "  update [ENV]                Update images (dev/prod)"
    echo "  reset [ENV]                 Reset environment (dev/prod)"
    echo "  help                        Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 cleanup                  # Gentle cleanup"
    echo "  $0 cleanup --force          # Force cleanup (removes everything)"
    echo "  $0 backup                   # Create backup"
    echo "  $0 restore backups/20231201_120000  # Restore from backup"
    echo "  $0 update development       # Update development images"
    echo "  $0 reset production         # Reset production environment"
}

# Main command handling
case "${1:-help}" in
    cleanup)
        cleanup_docker "${2:-false}"
        ;;
    backup)
        backup_data
        ;;
    restore)
        if [ -z "${2:-}" ]; then
            log_error "Please specify backup directory"
            echo "Usage: $0 restore BACKUP_DIR"
            exit 1
        fi
        restore_data "$2"
        ;;
    update)
        update_images "${2:-production}"
        ;;
    reset)
        reset_environment "${2:-production}"
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
