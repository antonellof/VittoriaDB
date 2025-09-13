# Installation Guide

This guide covers all installation methods for VittoriaDB across different platforms and use cases.

## 📦 Installation Methods

### Option 1: Pre-built Binaries (Recommended)

#### Quick Install Script
```bash
# Install latest version
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash

# Or install specific version
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash -s -- --version v0.2.0
```

#### Manual Download
Download from [GitHub Releases](https://github.com/antonellof/VittoriaDB/releases/latest):

- **Linux AMD64**: `vittoriadb-v0.2.0-linux-amd64.tar.gz`
- **Linux ARM64**: `vittoriadb-v0.2.0-linux-arm64.tar.gz`  
- **macOS Intel**: `vittoriadb-v0.2.0-darwin-amd64.tar.gz`
- **macOS Apple Silicon**: `vittoriadb-v0.2.0-darwin-arm64.tar.gz`
- **Windows**: `vittoriadb-v0.2.0-windows-amd64.zip`

```bash
# Example for Linux
wget https://github.com/antonellof/VittoriaDB/releases/download/v0.2.0/vittoriadb-v0.2.0-linux-amd64.tar.gz
tar -xzf vittoriadb-v0.2.0-linux-amd64.tar.gz
chmod +x vittoriadb-v0.2.0-linux-amd64
./vittoriadb-v0.2.0-linux-amd64 run
```

#### From Source
```bash
go install github.com/antonellof/VittoriaDB/cmd/vittoriadb@latest
vittoriadb run
```

### Option 2: Python Package (Development)

```bash
# Clone the repository
git clone https://github.com/antonellof/VittoriaDB.git
cd VittoriaDB

# Install Python package in development mode
cd sdk/python && ./install-dev.sh

# Or manually install in editable mode
pip install -e ./sdk/python
```

```python
import vittoriadb

# Connect to running server (recommended for development)
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)
collection = db.create_collection("documents", dimensions=384)
```


## 🔧 Platform-Specific Instructions

### Linux

#### Ubuntu/Debian
```bash
# Install dependencies
sudo apt-get update
sudo apt-get install curl wget

# Install VittoriaDB
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash
```

#### CentOS/RHEL/Fedora
```bash
# Install dependencies
sudo yum install curl wget  # CentOS/RHEL
sudo dnf install curl wget  # Fedora

# Install VittoriaDB
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash
```

### macOS

#### Using Homebrew (Planned)
```bash
brew install vittoriadb
```

#### Manual Installation
```bash
# Download for your architecture
# Intel Macs
wget https://github.com/antonellof/VittoriaDB/releases/download/v0.2.0/vittoriadb-v0.2.0-darwin-amd64.tar.gz

# Apple Silicon Macs
wget https://github.com/antonellof/VittoriaDB/releases/download/v0.2.0/vittoriadb-v0.2.0-darwin-arm64.tar.gz

# Extract and install
tar -xzf vittoriadb-v0.2.0-darwin-*.tar.gz
chmod +x vittoriadb-v0.2.0-darwin-*
sudo mv vittoriadb-v0.2.0-darwin-* /usr/local/bin/vittoriadb
```

### Windows

#### PowerShell Installation
```powershell
# Download and extract
Invoke-WebRequest -Uri "https://github.com/antonellof/VittoriaDB/releases/download/v0.2.0/vittoriadb-v0.2.0-windows-amd64.zip" -OutFile "vittoriadb.zip"
Expand-Archive -Path "vittoriadb.zip" -DestinationPath "."
.\vittoriadb-v0.2.0-windows-amd64.exe run
```

#### Manual Installation
1. Download `vittoriadb-v0.2.0-windows-amd64.zip`
2. Extract to desired location
3. Add to PATH (optional)
4. Run `vittoriadb.exe run`

## 🐍 Python SDK Installation

### Development Installation
```bash
# Clone repository
git clone https://github.com/antonellof/VittoriaDB.git
cd VittoriaDB

# Install in development mode
cd sdk/python && ./install-dev.sh

# Verify installation
python -c "import vittoriadb; print('✅ VittoriaDB Python SDK ready!')"
```

### Dependencies
```bash
# Core dependencies
pip install numpy

# For RAG examples with embeddings
pip install sentence-transformers

# For performance benchmarks
pip install psutil

# Optional: for advanced RAG features
pip install openai
```

## 🔍 Verification

### Basic Functionality Test
```bash
# Start the server
./vittoriadb run &
SERVER_PID=$!

# Wait for startup
sleep 2

# Create a test collection
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name": "test", "dimensions": 4, "metric": "cosine"}'

# Insert a test vector
curl -X POST http://localhost:8080/collections/test/vectors \
  -H "Content-Type: application/json" \
  -d '{"id": "test1", "vector": [0.1, 0.2, 0.3, 0.4], "metadata": {"type": "test"}}'

# Search for similar vectors
curl "http://localhost:8080/collections/test/search?vector=0.1,0.2,0.3,0.4&limit=1"

# Cleanup
kill $SERVER_PID
```

### Python Client Test
```python
import vittoriadb

# Connect to server
db = vittoriadb.connect(url="http://localhost:8080", auto_start=False)

# Create collection
collection = db.create_collection("test", dimensions=4)

# Insert and search
collection.insert("test1", [0.1, 0.2, 0.3, 0.4], {"type": "test"})
results = collection.search([0.1, 0.2, 0.3, 0.4], limit=1)

print(f"Found {len(results)} results")
db.close()
```

## 🚨 Troubleshooting

### Common Issues

**Port Already in Use**
```bash
# Check what's using port 8080
lsof -i :8080

# Use a different port
./vittoriadb run --port 9090
```

**Permission Denied**
```bash
# Make binary executable
chmod +x ./vittoriadb

# Or run with explicit path
./vittoriadb run
```

**Python Import Errors**
```bash
# Reinstall Python package
cd sdk/python && ./install-dev.sh

# Or manually reinstall
pip uninstall vittoriadb
pip install -e ./sdk/python
```

## 📋 System Requirements

### Minimum Requirements
- **Operating System**: Linux, macOS, or Windows
- **Memory**: 512MB RAM
- **Disk Space**: 100MB for binary + storage for your data
- **Network**: Port 8080 (configurable)

### Recommended Requirements
- **Memory**: 2GB+ RAM
- **CPU**: Multi-core processor
- **Disk**: SSD for better performance
- **Network**: Dedicated port for production use

### Development Requirements
- **Go**: Version 1.21 or higher (for building from source)
- **Python**: Version 3.7 or higher (for Python client)
- **Git**: For cloning the repository

## 🔄 Upgrading

### Binary Upgrade
```bash
# Download new version
curl -fsSL https://raw.githubusercontent.com/antonellof/VittoriaDB/main/scripts/install.sh | bash

# Or manually download and replace binary
```

### Python SDK Upgrade
```bash
cd sdk/python
git pull origin main
./install-dev.sh
```

## 🗑️ Uninstallation

### Remove Binary
```bash
# If installed via script
rm /usr/local/bin/vittoriadb

# If manually installed
rm /path/to/vittoriadb
```

### Remove Python SDK
```bash
pip uninstall vittoriadb
```

### Remove Data
```bash
# Remove data directory (be careful!)
rm -rf ./data
```
