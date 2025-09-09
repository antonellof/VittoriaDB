#!/bin/bash

# VittoriaDB Release Build Script
# Builds executables for multiple platforms

set -e

VERSION=${1:-"v0.1.0"}
BUILD_DIR="./releases/${VERSION}"
BINARY_NAME="vittoriadb"

echo "🚀 Building VittoriaDB ${VERSION} for multiple platforms..."

# Create build directory
mkdir -p "${BUILD_DIR}"

# Build information
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD)
GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "dev")

# Build flags
LDFLAGS="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT} -X main.GitTag=${GIT_TAG} -s -w"

# Platform configurations
declare -a PLATFORMS=(
    "linux/amd64"
    "linux/arm64" 
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

echo "📦 Building for platforms: ${PLATFORMS[@]}"

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    
    # Set output filename
    OUTPUT_NAME="${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    OUTPUT_PATH="${BUILD_DIR}/${OUTPUT_NAME}"
    
    echo "🔨 Building ${GOOS}/${GOARCH}..."
    
    # Build the binary
    env GOOS="$GOOS" GOARCH="$GOARCH" CGO_ENABLED=0 go build \
        -ldflags="${LDFLAGS}" \
        -o "${OUTPUT_PATH}" \
        ./cmd/vittoriadb
    
    # Create compressed archive
    ARCHIVE_NAME="${BINARY_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        ARCHIVE_NAME="${ARCHIVE_NAME}.zip"
        (cd "${BUILD_DIR}" && zip -q "${ARCHIVE_NAME}" "$(basename "${OUTPUT_PATH}")")
    else
        ARCHIVE_NAME="${ARCHIVE_NAME}.tar.gz"
        (cd "${BUILD_DIR}" && tar -czf "${ARCHIVE_NAME}" "$(basename "${OUTPUT_PATH}")")
    fi
    
    # Calculate checksums
    if command -v sha256sum >/dev/null 2>&1; then
        (cd "${BUILD_DIR}" && sha256sum "${ARCHIVE_NAME}" >> "checksums-${VERSION}.txt")
    elif command -v shasum >/dev/null 2>&1; then
        (cd "${BUILD_DIR}" && shasum -a 256 "${ARCHIVE_NAME}" >> "checksums-${VERSION}.txt")
    fi
    
    echo "✅ Built ${OUTPUT_PATH} ($(du -h "${OUTPUT_PATH}" | cut -f1))"
done

echo ""
echo "🎉 Build complete! Files created in ${BUILD_DIR}:"
ls -la "${BUILD_DIR}"

echo ""
echo "📋 Checksums:"
if [ -f "${BUILD_DIR}/checksums-${VERSION}.txt" ]; then
    cat "${BUILD_DIR}/checksums-${VERSION}.txt"
fi

echo ""
echo "🚀 Ready for GitHub release!"
echo "   Upload the archives and checksums file to GitHub releases."
