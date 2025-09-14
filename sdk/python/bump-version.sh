#!/bin/bash
# Simple version bumping script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_step() { echo -e "${BLUE}🔄 $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_error() { echo -e "${RED}❌ $1${NC}"; }

# Show usage
if [ $# -eq 0 ] || [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Version Bumper"
    echo "Usage: $0 <patch|minor|major>"
    echo ""
    echo "Examples:"
    echo "  $0 patch     # 0.1.0 -> 0.1.1"
    echo "  $0 minor     # 0.1.0 -> 0.2.0"
    echo "  $0 major     # 0.1.0 -> 1.0.0"
    exit 0
fi

# Check if we're in the right directory
if [ ! -f "vittoriadb/__init__.py" ]; then
    print_error "Must be run from the SDK python directory"
    exit 1
fi

# Get current version
current_version=$(python -c "
with open('vittoriadb/__init__.py') as f:
    for line in f:
        if line.startswith('__version__'):
            print(line.split('=')[1].strip().strip('\"').strip(\"'\"))
            break
")

if [ -z "$current_version" ]; then
    print_error "Could not read current version"
    exit 1
fi

# Parse version parts
IFS='.' read -r major minor patch <<< "$current_version"

# Bump version based on type
case "$1" in
    "patch")
        patch=$((patch + 1))
        ;;
    "minor")
        minor=$((minor + 1))
        patch=0
        ;;
    "major")
        major=$((major + 1))
        minor=0
        patch=0
        ;;
    *)
        print_error "Invalid bump type: $1"
        echo "Use: patch, minor, or major"
        exit 1
        ;;
esac

new_version="$major.$minor.$patch"

print_step "Current version: $current_version"
print_step "New version: $new_version"

# Confirm
read -p "Bump version? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_error "Version bump cancelled"
    exit 0
fi

# Update version in __init__.py
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS sed
    sed -i '' "s/__version__ = [\"'].*[\"']/__version__ = \"$new_version\"/" vittoriadb/__init__.py
else
    # Linux sed
    sed -i "s/__version__ = [\"'].*[\"']/__version__ = \"$new_version\"/" vittoriadb/__init__.py
fi

print_success "Version bumped from $current_version to $new_version"
