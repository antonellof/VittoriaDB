# VittoriaDB RAG Web UI - Docker Guide

Complete Docker containerization setup for the VittoriaDB RAG Web UI with production-ready configurations, development environments, and comprehensive monitoring.

## 🚀 Quick Start

### Production Deployment
```bash
# Clone and navigate to the project
cd examples/web-ui-rag

# Start production environment
./docker-start.sh

# Monitor services
./docker-monitor.sh
```

### Development Environment
```bash
# Start development environment with hot reload
./docker-start.sh -e development

# Monitor in development mode
./docker-monitor.sh
```

## 📋 Prerequisites

- **Docker**: Version 20.10+ ([Install Docker](https://docs.docker.com/get-docker/))
- **Docker Compose**: Version 2.0+ ([Install Compose](https://docs.docker.com/compose/install/))
- **System Requirements**:
  - RAM: 4GB minimum, 8GB recommended
  - Storage: 10GB free space
  - CPU: 2 cores minimum, 4 cores recommended

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ Nginx (Production)                                          │
│ ├─ Load balancing                                           │
│ ├─ SSL termination                                          │
│ └─ Static file serving                                      │
└─────────────────┬───────────────────────────────────────────┘
                  │ HTTP/HTTPS
┌─────────────────▼───────────────────────────────────────────┐
│ Next.js Frontend (Port 3000)                              │
│ ├─ React components with shadcn/ui                         │
│ ├─ Real-time chat interface                                │
│ └─ Hot reload (development)                                 │
└─────────────────┬───────────────────────────────────────────┘
                  │ HTTP/WebSocket
┌─────────────────▼───────────────────────────────────────────┐
│ FastAPI Backend (Port 8501)                               │
│ ├─ RAG system integration                                  │
│ ├─ File processing pipeline                                │
│ ├─ Web research & GitHub indexing                          │
│ └─ Real-time WebSocket streaming                           │
└─────────────────┬───────────────────────────────────────────┘
                  │ Python SDK
┌─────────────────▼───────────────────────────────────────────┐
│ VittoriaDB Server (Port 8080)                             │
│ ├─ Vector storage & search                                 │
│ ├─ Collection management                                   │
│ └─ High-performance indexing                               │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Supporting Services                                         │
│ ├─ Ollama (Port 11434) - Local embeddings                 │
│ ├─ Redis (Port 6379) - Caching & sessions                 │
│ └─ Redis Commander (Dev) - Database debugging             │
└─────────────────────────────────────────────────────────────┘
```

## 🐳 Docker Configurations

### Available Environments

| Environment | Compose File | Features |
|-------------|--------------|----------|
| **Production** | `docker-compose.prod.yml` | Multi-stage builds, resource limits, security hardening |
| **Development** | `docker-compose.dev.yml` | Hot reload, debug logging, development tools |
| **Default** | `docker-compose.yml` | Basic setup for quick testing |

### Container Images

| Service | Base Image | Size | Purpose |
|---------|------------|------|---------|
| **Frontend** | `node:22-alpine` | ~150MB | Next.js application |
| **Backend** | `python:3.11-slim` | ~800MB | FastAPI + ML libraries |
| **VittoriaDB** | `antonellof/vittoriadb:v0.4.0` | ~50MB | Vector database |
| **Ollama** | `ollama/ollama:latest` | ~2GB | Local LLM inference |
| **Redis** | `redis:7-alpine` | ~30MB | Caching & sessions |

## 🛠️ Scripts Overview

### Main Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| `docker-start.sh` | Start services | `./docker-start.sh [options]` |
| `docker-monitor.sh` | Monitor services | `./docker-monitor.sh [command]` |
| `docker-utils.sh` | Utility functions | `./docker-utils.sh [command]` |

### docker-start.sh Options

```bash
./docker-start.sh [OPTIONS]

Options:
  -e, --env ENV        Environment: development, production (default: production)
  -f, --foreground     Run in foreground (not detached)
  --no-build          Skip building images
  --no-models         Skip pulling Ollama models
  -h, --help          Show help message

Examples:
  ./docker-start.sh                           # Production mode
  ./docker-start.sh -e development            # Development with hot reload
  ./docker-start.sh -e production -f          # Production in foreground
  ./docker-start.sh --no-build --no-models    # Quick restart
```

### docker-monitor.sh Commands

```bash
./docker-monitor.sh [COMMAND]

Commands:
  monitor, m          Live monitoring dashboard (default)
  status, s           Show current status
  urls, u             Show service URLs
  logs, l [service]   Show logs (all or specific service)
  help, h             Show help

Examples:
  ./docker-monitor.sh                  # Live monitoring
  ./docker-monitor.sh status           # Quick status check
  ./docker-monitor.sh logs backend     # Backend logs
```

### docker-utils.sh Commands

```bash
./docker-utils.sh [COMMAND]

Commands:
  cleanup [--force]           Clean up Docker resources
  backup                      Backup data and configuration
  restore BACKUP_DIR          Restore from backup
  update [ENV]                Update images (dev/prod)
  reset [ENV]                 Reset environment (dev/prod)
  help                        Show help

Examples:
  ./docker-utils.sh cleanup                  # Gentle cleanup
  ./docker-utils.sh backup                   # Create backup
  ./docker-utils.sh update development       # Update dev images
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file or use the provided templates:

```bash
# Copy appropriate template
cp env.production .env    # For production
cp env.development .env   # For development
```

#### Required Variables

```bash
# OpenAI API Key (required)
OPENAI_API_KEY=your_openai_api_key_here

# Optional: GitHub token for private repos
GITHUB_TOKEN=your_github_token_here
```

#### Service URLs (Auto-configured)

```bash
# Internal service communication
VITTORIADB_URL=http://vittoriadb:8080
OLLAMA_URL=http://ollama:11434
REDIS_URL=redis://redis:6379

# Frontend configuration
NEXT_PUBLIC_API_URL=http://localhost:8501
NEXT_PUBLIC_WS_URL=ws://localhost:8501
```

### Volume Mounts

| Volume | Purpose | Persistence |
|--------|---------|-------------|
| `vittoriadb_data` | Vector database storage | Persistent |
| `ollama_data` | LLM models and cache | Persistent |
| `redis_data` | Cache and session data | Persistent |
| `backend_logs` | Application logs | Persistent |
| `./uploads` | File uploads | Host mount |

## 🚀 Deployment Scenarios

### Local Development

```bash
# Start development environment
./docker-start.sh -e development

# Features:
# - Hot reload for frontend and backend
# - Debug logging enabled
# - Redis Commander for debugging
# - Source code mounted for live editing
```

### Production Deployment

```bash
# Start production environment
./docker-start.sh -e production

# Features:
# - Multi-stage optimized builds
# - Resource limits and security hardening
# - Health checks and monitoring
# - Nginx reverse proxy (optional)
```

### CI/CD Integration

```bash
# Build and test
docker-compose -f docker-compose.prod.yml build
docker-compose -f docker-compose.prod.yml up -d
./docker-monitor.sh status

# Cleanup after tests
./docker-utils.sh cleanup
```

## 📊 Monitoring & Health Checks

### Built-in Health Checks

All services include comprehensive health checks:

- **VittoriaDB**: `GET /health`
- **Backend**: `GET /health`
- **Frontend**: `GET /` (Next.js)
- **Ollama**: `GET /api/tags`
- **Redis**: `redis-cli ping`

### Live Monitoring

```bash
# Start live monitoring dashboard
./docker-monitor.sh

# Features:
# - Real-time container status
# - Resource usage (CPU, memory, network)
# - Service health checks
# - Recent logs from all services
# - Volume usage statistics
```

### Log Management

```bash
# View all logs
./docker-monitor.sh logs

# View specific service logs
./docker-monitor.sh logs backend
./docker-monitor.sh logs frontend

# Follow logs in real-time
docker-compose -f docker-compose.prod.yml logs -f
```

## 🔒 Security Features

### Production Security

- **Non-root containers**: All services run as non-root users
- **Resource limits**: CPU and memory constraints
- **Network isolation**: Services communicate via internal network
- **Volume permissions**: Proper file ownership and permissions
- **Health checks**: Automatic service recovery

### Development Security

- **Isolated environment**: Separate from production
- **Debug access**: Enhanced logging and debugging tools
- **Hot reload**: Safe code changes without container rebuilds

## 🛠️ Troubleshooting

### Common Issues

#### Services Not Starting

```bash
# Check service status
./docker-monitor.sh status

# View logs for failing service
./docker-monitor.sh logs [service_name]

# Restart specific service
docker-compose restart [service_name]
```

#### Port Conflicts

```bash
# Check port usage
netstat -tulpn | grep -E "(3000|8501|8080|11434|6379)"

# Stop conflicting services
sudo systemctl stop [service_name]
```

#### Resource Issues

```bash
# Check Docker resources
docker system df
docker stats

# Clean up unused resources
./docker-utils.sh cleanup
```

#### Volume Issues

```bash
# Reset volumes (WARNING: Data loss)
docker-compose down -v

# Backup before reset
./docker-utils.sh backup
./docker-utils.sh reset production
```

### Performance Optimization

#### Resource Allocation

```yaml
# Adjust in docker-compose.prod.yml
deploy:
  resources:
    limits:
      memory: 2G
      cpus: '1.0'
    reservations:
      memory: 1G
      cpus: '0.5'
```

#### Cache Optimization

```bash
# Optimize Redis cache
# Edit docker-compose.yml Redis command:
command: redis-server --maxmemory 200mb --maxmemory-policy allkeys-lru
```

## 📚 Advanced Usage

### Custom Configurations

#### Nginx Reverse Proxy

```bash
# Create nginx.conf for production
# Enable nginx service in docker-compose.prod.yml
# Configure SSL certificates in ./ssl/
```

#### GPU Support (Ollama)

```yaml
# Uncomment in docker-compose files:
deploy:
  resources:
    reservations:
      devices:
        - driver: nvidia
          count: 1
          capabilities: [gpu]
```

### Backup & Restore

```bash
# Create backup
./docker-utils.sh backup

# Restore from backup
./docker-utils.sh restore backups/20231201_120000

# Automated backups (add to crontab)
0 2 * * * cd /path/to/web-ui-rag && ./docker-utils.sh backup
```

### Scaling

```bash
# Scale specific services
docker-compose -f docker-compose.prod.yml up --scale backend=3

# Load balancing with Nginx
# Configure upstream servers in nginx.conf
```

## 🆘 Support

### Getting Help

1. **Check logs**: `./docker-monitor.sh logs`
2. **Verify configuration**: `./docker-monitor.sh status`
3. **Clean and restart**: `./docker-utils.sh cleanup && ./docker-start.sh`
4. **Reset environment**: `./docker-utils.sh reset`

### Useful Commands

```bash
# Quick status check
./docker-monitor.sh status

# View resource usage
docker stats

# Access container shell
docker exec -it [container_name] /bin/bash

# View container configuration
docker inspect [container_name]
```

### Performance Monitoring

```bash
# Monitor resource usage
./docker-monitor.sh

# Check disk usage
docker system df -v

# Analyze logs
docker-compose logs --since=1h | grep ERROR
```

---

## 📄 Files Reference

### Configuration Files

- `docker-compose.yml` - Default configuration
- `docker-compose.prod.yml` - Production optimized
- `docker-compose.dev.yml` - Development with hot reload
- `env.example` - Environment template
- `env.production` - Production environment template
- `env.development` - Development environment template

### Docker Files

- `backend/Dockerfile` - Production backend image
- `backend/Dockerfile.dev` - Development backend image
- `frontend/Dockerfile` - Production frontend image
- `frontend/Dockerfile.dev` - Development frontend image
- `backend/.dockerignore` - Backend build exclusions
- `frontend/.dockerignore` - Frontend build exclusions

### Scripts

- `docker-start.sh` - Main startup script
- `docker-monitor.sh` - Monitoring and status
- `docker-utils.sh` - Utility functions

This Docker setup provides a complete, production-ready containerized environment for the VittoriaDB RAG Web UI with comprehensive monitoring, backup capabilities, and development support.
