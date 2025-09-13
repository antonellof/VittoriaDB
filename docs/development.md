# Development Guide

This guide covers building, testing, and contributing to VittoriaDB from source.

## 🛠️ Prerequisites

Before building VittoriaDB, ensure you have the required tools installed:

### Required Tools
```bash
# Check Go version (required: 1.21+)
go version

# Check Python version (required: 3.7+)
python --version

# Install Git if not already installed
git --version
```

### Optional Tools
```bash
# Make (for build scripts)
make --version

# Docker (for containerized testing)
docker --version

# jq (for JSON processing in examples)
jq --version
```

## 🏗️ Building from Source

### 1. Clone the Repository
```bash
git clone https://github.com/antonellof/VittoriaDB.git
cd VittoriaDB
```

### 2. Build the Go Binary
```bash
# Download Go dependencies
go mod download

# Build the binary
go build -o vittoriadb ./cmd/vittoriadb

# Verify the build
./vittoriadb --version
```

### 3. Build Python Package (Optional)
```bash
# Navigate to Python package directory
cd sdk/python

# Option 1: Use the development installation script (recommended)
./install-dev.sh

# Option 2: Manual installation in development mode
pip install -e .

# Option 3: Install with optional dependencies
pip install -e ".[dev,full]"

# Verify installation
python -c "import vittoriadb; print('VittoriaDB Python client installed successfully')"
```

## 🚀 Running Locally

### Quick Start
```bash
# Start VittoriaDB server
./vittoriadb run

# In another terminal, test the API
curl http://localhost:8080/health
```

### With Custom Configuration
```bash
# Run with custom settings
./vittoriadb run \
  --host 0.0.0.0 \
  --port 9090 \
  --data-dir ./my-data \
  --cors

# Or use a configuration file
./vittoriadb run --config ./config/vittoriadb.yaml
```

### Development Mode
```bash
# Run with verbose logging
./vittoriadb run --log-level debug

# Run with performance monitoring
./vittoriadb run --enable-simd --memory-limit 2GB
```

## 🧪 Testing

### Go Tests
```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./pkg/core -v

# Run benchmarks
go test ./pkg/core -bench=. -benchmem

# Run tests with race detection
go test ./... -race
```

### Python Tests
```bash
# Navigate to Python package directory
cd sdk/python

# Run Python tests
python -m pytest tests/ -v

# Run with coverage
python -m pytest tests/ --cov=vittoriadb --cov-report=html

# Run specific test file
python -m pytest tests/test_client.py -v
```

### Integration Tests
```bash
# Start server for integration tests
./vittoriadb run --port 8081 &
SERVER_PID=$!

# Run integration tests
go test ./tests/integration -v

# Run Python integration tests
cd sdk/python && python -m pytest tests/integration/ -v

# Cleanup
kill $SERVER_PID
```

### Example Tests
```bash
# Test Go examples
cd examples/go && go run basic_usage.go
cd examples/go && go run volume_benchmark.go

# Test Python examples (requires Python SDK)
python examples/python/basic_usage.py
python examples/python/rag_complete_example.py

# Test cURL examples
cd examples/curl && ./basic_usage.sh
cd examples/curl && ./volume_test.sh
```

## 🔨 Build Scripts

### Using Make (if available)
```bash
# Build everything
make build

# Run tests
make test

# Clean build artifacts
make clean

# Build for multiple platforms
make build-all

# Run linting
make lint

# Generate documentation
make docs
```

### Manual Cross-Platform Builds
```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o vittoriadb-linux-amd64 ./cmd/vittoriadb

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o vittoriadb-darwin-amd64 ./cmd/vittoriadb

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o vittoriadb-darwin-arm64 ./cmd/vittoriadb

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o vittoriadb-windows-amd64.exe ./cmd/vittoriadb

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o vittoriadb-linux-arm64 ./cmd/vittoriadb
```

### Release Builds
```bash
# Build with version information
VERSION=$(git describe --tags --always --dirty)
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse HEAD)

go build -ldflags "\
  -X main.Version=${VERSION} \
  -X main.BuildTime=${BUILD_TIME} \
  -X main.GitCommit=${GIT_COMMIT}" \
  -o vittoriadb ./cmd/vittoriadb
```

## 🧹 Code Quality

### Linting
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
golangci-lint run

# Run with auto-fix
golangci-lint run --fix

# Run specific linters
golangci-lint run --enable=gofmt,goimports,govet
```

### Formatting
```bash
# Format Go code
go fmt ./...

# Format imports
goimports -w .

# Format Python code (if black is installed)
cd sdk/python && black .

# Format Python imports (if isort is installed)
cd sdk/python && isort .
```

### Static Analysis
```bash
# Run go vet
go vet ./...

# Run staticcheck
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...

# Run gosec (security analysis)
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
gosec ./...
```

## 🐛 Debugging

### Debug Builds
```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o vittoriadb-debug ./cmd/vittoriadb

# Run with delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv exec ./vittoriadb-debug -- run --port 8080
```

### Profiling
```bash
# CPU profiling
go test ./pkg/core -cpuprofile=cpu.prof -bench=.

# Memory profiling
go test ./pkg/core -memprofile=mem.prof -bench=.

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Logging
```bash
# Enable debug logging
./vittoriadb run --log-level debug

# Log to file
./vittoriadb run --log-level debug > debug.log 2>&1

# Structured logging analysis
./vittoriadb run --log-level debug | jq 'select(.level == "error")'
```

## 🚨 Troubleshooting

### Common Build Issues

**1. Go Module Issues**
```bash
# Clean module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod tidy

# Verify module integrity
go mod verify
```

**2. Python Package Issues**
```bash
# Reinstall Python package
cd sdk/python && ./install-dev.sh

# Clear Python cache
find . -name "__pycache__" -type d -exec rm -rf {} +
find . -name "*.pyc" -delete

# Check Python path
python -c "import sys; print(sys.path)"
```

**3. Port Conflicts**
```bash
# Check what's using port 8080
lsof -i :8080

# Use a different port
./vittoriadb run --port 9090
```

**4. Permission Issues**
```bash
# Make binary executable
chmod +x ./vittoriadb

# Check data directory permissions
ls -la ./data

# Fix permissions
chmod -R 755 ./data
```

### Performance Issues

**Memory Usage**
```bash
# Monitor memory usage
./vittoriadb run --memory-limit 1GB

# Profile memory usage
go test ./pkg/core -memprofile=mem.prof -bench=BenchmarkInsert
go tool pprof mem.prof
```

**CPU Usage**
```bash
# Enable SIMD optimizations
./vittoriadb run --enable-simd

# Profile CPU usage
go test ./pkg/core -cpuprofile=cpu.prof -bench=BenchmarkSearch
go tool pprof cpu.prof
```

## 📁 Project Structure

```
vittoriadb/
├── cmd/vittoriadb/           # Main binary
│   └── main.go
├── pkg/
│   ├── core/                 # Core database engine
│   │   ├── database.go
│   │   ├── collection.go
│   │   └── types.go
│   ├── storage/              # File storage layer
│   │   ├── engine.go
│   │   ├── wal.go
│   │   └── cache.go
│   ├── index/                # Vector indexing
│   │   ├── flat.go
│   │   ├── hnsw.go
│   │   └── distance.go
│   ├── server/               # HTTP API server
│   │   └── server.go
│   ├── processor/            # Document processing
│   │   ├── pdf.go
│   │   ├── docx.go
│   │   └── text.go
│   └── embeddings/           # Embedding integrations
├── sdk/python/vittoriadb/    # Python SDK package
│   ├── __init__.py
│   ├── client.py
│   └── types.py
├── examples/                 # Code examples
│   ├── python/              # Python client examples
│   ├── go/                  # Go native examples
│   ├── curl/                # HTTP API examples
│   └── documents/           # Sample documents
├── docs/                     # Documentation
│   ├── installation.md
│   ├── api.md
│   ├── configuration.md
│   ├── performance.md
│   ├── cli.md
│   └── development.md
├── tests/                    # Test suites
├── scripts/                  # Build and utility scripts
├── Makefile                  # Build automation
├── go.mod                    # Go module definition
└── README.md                 # Main documentation
```

## 🤝 Contributing

### Development Workflow

1. **Fork and Clone**
   ```bash
   git clone https://github.com/yourusername/VittoriaDB.git
   cd VittoriaDB
   ```

2. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Changes**
   - Follow Go and Python coding standards
   - Add tests for new functionality
   - Update documentation as needed

4. **Test Changes**
   ```bash
   # Run all tests
   make test
   
   # Run linting
   make lint
   
   # Test examples
   make test-examples
   ```

5. **Commit and Push**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   git push origin feature/your-feature-name
   ```

6. **Submit Pull Request**
   - Create PR with clear description
   - Include tests and documentation
   - Ensure CI passes

### Code Standards

**Go Code**
- Follow `gofmt` formatting
- Use `golangci-lint` for linting
- Write comprehensive tests
- Document public APIs
- Handle errors appropriately

**Python Code**
- Follow PEP 8 style guide
- Use type hints
- Write docstrings
- Include unit tests
- Handle exceptions properly

### Testing Requirements

- All new features must include tests
- Maintain >80% test coverage
- Include integration tests for API changes
- Test cross-platform compatibility
- Benchmark performance-critical code

### Documentation Requirements

- Update README.md for user-facing changes
- Add/update API documentation
- Include code examples
- Update CLI help text
- Write clear commit messages

## 🔄 Release Process

### Preparing a Release

1. **Update Version**
   ```bash
   # Update version in relevant files
   # Create changelog entry
   ```

2. **Test Release**
   ```bash
   # Run full test suite
   make test-all
   
   # Test cross-platform builds
   make build-all
   
   # Test examples
   make test-examples
   ```

3. **Create Release**
   ```bash
   # Tag release
   git tag v0.2.1
   git push origin v0.2.1
   
   # GitHub Actions will build and publish
   ```

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] Version numbers updated
- [ ] Changelog updated
- [ ] Cross-platform builds tested
- [ ] Examples tested
- [ ] Performance benchmarks run
- [ ] Security scan passed
