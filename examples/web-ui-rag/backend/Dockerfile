# VittoriaDB RAG Backend Dockerfile
FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    git \
    build-essential \
    libmagic1 \
    && rm -rf /var/lib/apt/lists/*

# Install VittoriaDB binary manually
RUN ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = "amd64" ]; then ARCH="amd64"; elif [ "$ARCH" = "arm64" ]; then ARCH="arm64"; fi && \
    wget -O vittoriadb.tar.gz "https://github.com/antonellof/VittoriaDB/releases/download/v0.3.0/vittoriadb-v0.3.0-linux-${ARCH}.tar.gz" && \
    tar -xzf vittoriadb.tar.gz && \
    mv vittoriadb-v0.3.0-linux-${ARCH} /usr/local/bin/vittoriadb && \
    chmod +x /usr/local/bin/vittoriadb && \
    rm vittoriadb.tar.gz

# Copy requirements first for better caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Create necessary directories
RUN mkdir -p /app/data /app/uploads /app/logs

# Expose port
EXPOSE 8501

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8501/health || exit 1

# Start command
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8501", "--reload"]
