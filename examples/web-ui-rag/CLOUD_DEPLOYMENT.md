# VittoriaDB RAG Web UI - Cloud Deployment Guide

This guide explains how to deploy the VittoriaDB RAG Web UI to various cloud platforms using the pre-built Docker images from GitHub Container Registry.

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose installed
- GitHub Container Registry access (public images)
- API keys (OpenAI, GitHub Token, etc.)

### 1. Deploy Locally with Cloud Images
```bash
# Clone the repository
git clone https://github.com/antonellof/VittoriaDB.git
cd VittoriaDB/examples/web-ui-rag

# Set required environment variables
export OPENAI_API_KEY=your_openai_api_key_here
export GITHUB_TOKEN=your_github_token_here  # optional
export OLLAMA_URL=http://ollama:11434        # optional

# Deploy using cloud images
./deploy-cloud.sh
```

### 2. Access the Application
- **Web UI**: http://localhost:3000
- **Backend API**: http://localhost:8501
- **VittoriaDB**: http://localhost:8080
- **Ollama**: http://localhost:11434

## 🏗️ Architecture

The application consists of 4 main services:

1. **VittoriaDB v0.5.0** (`ghcr.io/antonellof/vittoriadb:v0.5.0`)
   - High-performance vector database
   - Unified configuration system
   - I/O optimizations (SIMD, memory-mapped storage)
   - Parallel search engine

2. **RAG Backend** (FastAPI)
   - Document processing and indexing
   - RAG query processing
   - Web research capabilities
   - GitHub repository indexing

3. **Frontend** (Next.js)
   - ChatGPT-like interface
   - Real-time streaming responses
   - File upload and management

4. **Ollama** (Optional)
   - Local ML models for embeddings
   - No API costs, works offline

## ☁️ Cloud Platform Deployment

### AWS ECS (Elastic Container Service)

#### 1. Push Images to ECR (Optional)
```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com

# Tag and push (or use ghcr.io directly)
docker tag ghcr.io/antonellof/vittoriadb:v0.5.0 <account-id>.dkr.ecr.us-east-1.amazonaws.com/vittoriadb:v0.5.0
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/vittoriadb:v0.5.0
```

#### 2. Create ECS Task Definition
```json
{
  "family": "vittoriadb-rag",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "containerDefinitions": [
    {
      "name": "vittoriadb",
      "image": "ghcr.io/antonellof/vittoriadb:v0.5.0",
      "portMappings": [{"containerPort": 8080}],
      "environment": [
        {"name": "VITTORIADB_HOST", "value": "0.0.0.0"},
        {"name": "VITTORIA_PERF_ENABLE_SIMD", "value": "true"},
        {"name": "VITTORIA_SEARCH_PARALLEL_MAX_WORKERS", "value": "16"}
      ]
    }
  ]
}
```

### Google Cloud Run

#### 1. Deploy VittoriaDB
```bash
# Deploy to Cloud Run
gcloud run deploy vittoriadb \
  --image=ghcr.io/antonellof/vittoriadb:v0.5.0 \
  --platform=managed \
  --region=us-central1 \
  --allow-unauthenticated \
  --port=8080 \
  --set-env-vars="VITTORIA_PERF_ENABLE_SIMD=true,VITTORIA_SEARCH_PARALLEL_MAX_WORKERS=16"
```

#### 2. Deploy Backend and Frontend
```bash
# Build and deploy backend
cd backend
gcloud run deploy rag-backend \
  --source=. \
  --platform=managed \
  --region=us-central1 \
  --allow-unauthenticated \
  --set-env-vars="VITTORIADB_URL=https://vittoriadb-xxx.run.app"

# Build and deploy frontend
cd ../frontend
gcloud run deploy rag-frontend \
  --source=. \
  --platform=managed \
  --region=us-central1 \
  --allow-unauthenticated \
  --set-env-vars="NEXT_PUBLIC_API_URL=https://rag-backend-xxx.run.app"
```

### Azure Container Instances

#### 1. Create Resource Group
```bash
az group create --name vittoriadb-rag --location eastus
```

#### 2. Deploy Container Group
```bash
az container create \
  --resource-group vittoriadb-rag \
  --name vittoriadb-rag-app \
  --image ghcr.io/antonellof/vittoriadb:v0.5.0 \
  --ports 8080 \
  --dns-name-label vittoriadb-rag \
  --environment-variables \
    VITTORIADB_HOST=0.0.0.0 \
    VITTORIA_PERF_ENABLE_SIMD=true \
    VITTORIA_SEARCH_PARALLEL_MAX_WORKERS=16
```

### DigitalOcean App Platform

#### 1. Create App Spec (`app.yaml`)
```yaml
name: vittoriadb-rag
services:
- name: vittoriadb
  image:
    registry_type: GHCR
    registry: ghcr.io
    repository: antonellof/vittoriadb
    tag: v0.5.0
  http_port: 8080
  instance_count: 1
  instance_size_slug: basic-xxs
  env:
  - key: VITTORIADB_HOST
    value: "0.0.0.0"
  - key: VITTORIA_PERF_ENABLE_SIMD
    value: "true"
  - key: VITTORIA_SEARCH_PARALLEL_MAX_WORKERS
    value: "16"

- name: backend
  source_dir: backend
  github:
    repo: antonellof/VittoriaDB
    branch: main
  build_command: pip install -r requirements.txt
  run_command: uvicorn main:app --host 0.0.0.0 --port 8080
  http_port: 8080
  instance_count: 1
  instance_size_slug: basic-xxs
  env:
  - key: VITTORIADB_URL
    value: "${vittoriadb.PUBLIC_URL}"
  - key: OPENAI_API_KEY
    value: "${OPENAI_API_KEY}"

- name: frontend
  source_dir: frontend
  github:
    repo: antonellof/VittoriaDB
    branch: main
  build_command: npm install && npm run build
  run_command: npm start
  http_port: 3000
  instance_count: 1
  instance_size_slug: basic-xxs
  env:
  - key: NEXT_PUBLIC_API_URL
    value: "${backend.PUBLIC_URL}"
```

#### 2. Deploy
```bash
doctl apps create --spec app.yaml
```

### Railway

#### 1. Deploy via CLI
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login and deploy
railway login
railway project new vittoriadb-rag

# Deploy VittoriaDB
railway service create vittoriadb
railway service deploy --image ghcr.io/antonellof/vittoriadb:v0.5.0

# Deploy backend and frontend
railway service create backend
railway service create frontend
# Connect GitHub repo and deploy
```

### Render

#### 1. Create `render.yaml`
```yaml
services:
- type: web
  name: vittoriadb
  env: docker
  dockerfilePath: ./Dockerfile
  dockerContext: ../../
  plan: starter
  port: 8080
  envVars:
  - key: VITTORIADB_HOST
    value: "0.0.0.0"
  - key: VITTORIA_PERF_ENABLE_SIMD
    value: "true"

- type: web
  name: rag-backend
  env: python
  buildCommand: pip install -r requirements.txt
  startCommand: uvicorn main:app --host 0.0.0.0 --port $PORT
  plan: starter
  envVars:
  - key: VITTORIADB_URL
    fromService:
      type: web
      name: vittoriadb
      property: host

- type: web
  name: rag-frontend
  env: node
  buildCommand: npm install && npm run build
  startCommand: npm start
  plan: starter
  envVars:
  - key: NEXT_PUBLIC_API_URL
    fromService:
      type: web
      name: rag-backend
      property: host
```

## 🔧 Configuration

### Environment Variables

All API keys and sensitive configuration should be set as system environment variables for security.

#### Required Environment Variables
```bash
# Required for AI functionality
export OPENAI_API_KEY=your_openai_api_key_here

# Optional but recommended
export GITHUB_TOKEN=your_github_token_here    # For private repos and higher rate limits
export OLLAMA_URL=http://ollama:11434          # For local ML models
```

#### VittoriaDB v0.5.0 Performance Settings
```bash
# Enable SIMD optimizations
VITTORIA_PERF_ENABLE_SIMD=true

# Parallel search workers
VITTORIA_SEARCH_PARALLEL_MAX_WORKERS=16

# Memory-mapped I/O
VITTORIA_PERF_IO_USE_MEMORY_MAP=true

# Search caching
VITTORIA_SEARCH_CACHE_ENABLED=true
VITTORIA_SEARCH_CACHE_SIZE=1000
```

#### Application Settings
These are automatically passed from your system environment variables to the containers:

```bash
# Required - must be set in your system environment
OPENAI_API_KEY=your_openai_api_key

# Optional - will use defaults if not set
GITHUB_TOKEN=your_github_token
OLLAMA_URL=http://ollama:11434  # defaults to this if not set
```

### Resource Requirements

#### Minimum Requirements
- **CPU**: 1 vCPU
- **Memory**: 1GB RAM
- **Storage**: 5GB (for data and uploads)

#### Recommended for Production
- **CPU**: 2+ vCPUs
- **Memory**: 4GB+ RAM
- **Storage**: 20GB+ SSD

### Performance Tuning

#### VittoriaDB v0.5.0 Optimizations
```bash
# For high-throughput scenarios
VITTORIA_SEARCH_PARALLEL_MAX_WORKERS=32
VITTORIA_SEARCH_CACHE_SIZE=5000
VITTORIA_PERF_BATCH_SIZE=1000

# For memory-constrained environments
VITTORIA_SEARCH_PARALLEL_MAX_WORKERS=4
VITTORIA_SEARCH_CACHE_SIZE=500
VITTORIA_PERF_IO_USE_MEMORY_MAP=false
```

## 🔍 Monitoring and Logging

### Health Checks
- **VittoriaDB**: `GET /health`
- **Backend**: `GET /health`
- **Frontend**: `GET /` (returns 200)

### Metrics Endpoints
- **VittoriaDB**: `GET /stats` - Database statistics
- **VittoriaDB**: `GET /config` - Current configuration
- **Backend**: `GET /metrics` - Application metrics

### Log Aggregation
```bash
# View logs in Docker Compose
docker-compose -f docker-compose.cloud.yml logs -f

# Export logs for analysis
docker-compose -f docker-compose.cloud.yml logs > app.log
```

## 🔐 Security Considerations

### API Keys Management
- Use cloud provider secret management (AWS Secrets Manager, etc.)
- Never commit API keys to version control
- Rotate keys regularly

### Network Security
- Use HTTPS in production
- Configure proper firewall rules
- Consider VPC/private networking

### Container Security
- Images are built with non-root user
- Regular security updates via automated builds
- Minimal attack surface with Alpine Linux base

## 🚀 Scaling

### Horizontal Scaling
- Multiple VittoriaDB instances with load balancer
- Separate backend instances for different functions
- CDN for frontend static assets

### Vertical Scaling
- Increase CPU/memory for compute-intensive workloads
- SSD storage for better I/O performance
- GPU instances for Ollama ML models

## 📞 Support

- **Documentation**: [VittoriaDB Docs](../../docs/)
- **Issues**: [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues)
- **Discussions**: [GitHub Discussions](https://github.com/antonellof/VittoriaDB/discussions)

---

**Happy deploying! 🚀**
