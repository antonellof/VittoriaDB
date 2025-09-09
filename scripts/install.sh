#!/bin/bash

# VittoriaDB Installation Script
# Automatically detects platform and installs the latest release

set -e

# Configuration
REPO="antonellof/VittoriaDB"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
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

# Detect platform
detect_platform() {
    local os arch
    
    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          log_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

# Get latest version from GitHub
get_latest_version() {
    if [ "$VERSION" = "latest" ]; then
        log_info "Fetching latest version..."
        local latest_version
        latest_version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        
        if [ -z "$latest_version" ]; then
            log_error "Failed to fetch latest version"
            exit 1
        fi
        
        echo "$latest_version"
    else
        echo "$VERSION"
    fi
}

# Download and install
install_vittoriadb() {
    local platform version download_url filename
    
    platform=$(detect_platform)
    version=$(get_latest_version)
    
    log_info "Installing VittoriaDB ${version} for ${platform}..."
    
    # Construct download URL
    if [[ "$platform" == *"windows"* ]]; then
        filename="vittoriadb-${version}-${platform}.zip"
    else
        filename="vittoriadb-${version}-${platform}.tar.gz"
    fi
    
    download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    
    # Create temporary directory
    local temp_dir
    temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # Download
    log_info "Downloading from ${download_url}..."
    if ! curl -L -o "$filename" "$download_url"; then
        log_error "Failed to download VittoriaDB"
        exit 1
    fi
    
    # Extract
    log_info "Extracting..."
    if [[ "$filename" == *.zip ]]; then
        unzip -q "$filename"
        binary_name="vittoriadb-${version}-${platform}.exe"
    else
        tar -xzf "$filename"
        binary_name="vittoriadb-${version}-${platform}"
    fi
    
    # Install
    log_info "Installing to ${INSTALL_DIR}..."
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        log_warning "Creating directory ${INSTALL_DIR} (may require sudo)"
        sudo mkdir -p "$INSTALL_DIR"
    fi
    
    # Copy binary
    if [ -w "$INSTALL_DIR" ]; then
        cp "$binary_name" "${INSTALL_DIR}/vittoriadb"
        chmod +x "${INSTALL_DIR}/vittoriadb"
    else
        log_warning "Installing to ${INSTALL_DIR} (requires sudo)"
        sudo cp "$binary_name" "${INSTALL_DIR}/vittoriadb"
        sudo chmod +x "${INSTALL_DIR}/vittoriadb"
    fi
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$temp_dir"
    
    log_success "VittoriaDB ${version} installed successfully!"
    
    # Verify installation
    if command -v vittoriadb >/dev/null 2>&1; then
        log_success "Installation verified: $(vittoriadb version)"
    else
        log_warning "VittoriaDB installed but not in PATH. Add ${INSTALL_DIR} to your PATH or use full path: ${INSTALL_DIR}/vittoriadb"
    fi
    
    # Show quick start
    echo ""
    echo -e "${BLUE}🚀 Quick Start:${NC}"
    echo "  vittoriadb run                    # Start the server"
    echo "  vittoriadb version                # Show version info"
    echo "  vittoriadb --help                 # Show help"
    echo ""
    echo -e "${BLUE}📖 Documentation:${NC}"
    echo "  https://github.com/${REPO}#readme"
    echo ""
}

# Main execution
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                    VittoriaDB Installer                     ║"
    echo "║              Local Vector Database for AI                   ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v tar >/dev/null 2>&1 && ! command -v unzip >/dev/null 2>&1; then
        log_error "tar or unzip is required but not installed"
        exit 1
    fi
    
    install_vittoriadb
}

# Handle command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --help)
            echo "VittoriaDB Installation Script"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --version VERSION     Install specific version (default: latest)"
            echo "  --install-dir DIR     Installation directory (default: /usr/local/bin)"
            echo "  --help               Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  VERSION              Same as --version"
            echo "  INSTALL_DIR          Same as --install-dir"
            echo ""
            echo "Examples:"
            echo "  $0                           # Install latest version"
            echo "  $0 --version v0.1.0          # Install specific version"
            echo "  $0 --install-dir ~/.local/bin # Install to custom directory"
            echo ""
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

main
