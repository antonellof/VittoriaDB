# VittoriaDB Makefile

# Variables
BINARY_NAME=vittoriadb
VERSION?=0.1.0
BUILD_DIR=build
GO_FILES=$(shell find . -name "*.go" -type f)
PYTHON_DIR=python

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building VittoriaDB..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/vittoriadb
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/vittoriadb
	GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/vittoriadb

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/vittoriadb
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/vittoriadb

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/vittoriadb

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running Go tests..."
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Python package
.PHONY: python-install
python-install:
	@echo "Installing Python package in development mode..."
	cd $(PYTHON_DIR) && pip install -e .

.PHONY: python-install-full
python-install-full:
	@echo "Installing Python package with all dependencies..."
	cd $(PYTHON_DIR) && pip install -e ".[dev,full]"

.PHONY: python-test
python-test:
	@echo "Running Python tests..."
	cd $(PYTHON_DIR) && python -m pytest tests/ -v

# Development
.PHONY: run
run: build
	@echo "Starting VittoriaDB server..."
	./$(BUILD_DIR)/$(BINARY_NAME) run

.PHONY: run-dev
run-dev: build
	@echo "Starting VittoriaDB server in development mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) run --log-level debug --cors

# Linting and formatting
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

.PHONY: lint
lint:
	@echo "Running Go linter..."
	golangci-lint run

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache
	go clean -testcache

# Docker
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t vittoriadb:$(VERSION) .
	docker tag vittoriadb:$(VERSION) vittoriadb:latest

.PHONY: docker-run
docker-run:
	@echo "Running VittoriaDB in Docker..."
	docker run -p 8080:8080 -v ./data:/data vittoriadb:latest

# Release
.PHONY: release
release: clean build-all test
	@echo "Creating release artifacts..."
	@mkdir -p $(BUILD_DIR)/release
	cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(BUILD_DIR) && tar -czf release/$(BINARY_NAME)-$(VERSION)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(BUILD_DIR) && zip release/$(BINARY_NAME)-$(VERSION)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release artifacts created in $(BUILD_DIR)/release/"

# Quick test of the built binary
.PHONY: test-binary
test-binary: build
	@echo "Testing binary functionality..."
	@./$(BUILD_DIR)/$(BINARY_NAME) --version
	@echo "Starting server for 5 seconds..."
	@./$(BUILD_DIR)/$(BINARY_NAME) run --port 8081 &
	@SERVER_PID=$$!; \
	sleep 3; \
	echo "Testing health endpoint..."; \
	curl -s http://localhost:8081/health > /dev/null && echo "✅ Health check passed" || echo "❌ Health check failed"; \
	kill $$SERVER_PID 2>/dev/null || true; \
	wait $$SERVER_PID 2>/dev/null || true

# Show help
.PHONY: help
help:
	@echo "VittoriaDB Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary for current platform"
	@echo "  build-all      - Build for all supported platforms"
	@echo "  build-linux    - Build for Linux (amd64, arm64)"
	@echo "  build-darwin   - Build for macOS (amd64, arm64)"
	@echo "  build-windows  - Build for Windows (amd64)"
	@echo ""
	@echo "  test           - Run Go tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  bench          - Run benchmarks"
	@echo "  test-binary    - Test the built binary"
	@echo ""
	@echo "  python-install - Install Python package in dev mode"
	@echo "  python-install-full - Install Python package with all deps"
	@echo "  python-test    - Run Python tests"
	@echo ""
	@echo "  run            - Build and run the server"
	@echo "  run-dev        - Run in development mode"
	@echo ""
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter"
	@echo "  vet            - Run go vet"
	@echo ""
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run in Docker"
	@echo ""
	@echo "  clean          - Clean build artifacts"
	@echo "  release        - Create release artifacts"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION        - Version to build (default: $(VERSION))"
	@echo ""
	@echo "Examples:"
	@echo "  make build VERSION=1.0.0"
	@echo "  make test"
	@echo "  make run-dev"
