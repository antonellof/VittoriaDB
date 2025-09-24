# VittoriaDB Dockerfile
# Multi-stage build for optimized production image

# Build stage  
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.version=0.4.0 -w -s" \
    -o vittoriadb \
    ./cmd/vittoriadb

# Production stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1001 -S vittoriadb && \
    adduser -u 1001 -S vittoriadb -G vittoriadb

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/vittoriadb .

# Create data directory
RUN mkdir -p /data && chown -R vittoriadb:vittoriadb /data /app

# Switch to non-root user
USER vittoriadb

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Default command
CMD ["./vittoriadb", "run", "--host", "0.0.0.0", "--port", "8080", "--data-dir", "/data"]
