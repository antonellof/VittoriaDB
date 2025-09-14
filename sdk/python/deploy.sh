#!/bin/bash
# VittoriaDB Python Package - One-Command Deploy
# Usage: ./deploy.sh [test]

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
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }

# Show usage if help requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "VittoriaDB Python Package Deploy"
    echo "Usage: $0 [test|patch|minor|major]"
    echo ""
    echo "Commands:"
    echo "  $0           Deploy to Production PyPI (current version)"
    echo "  $0 test      Deploy to Test PyPI (current version)"
    echo "  $0 patch     Bump patch version (0.1.0 -> 0.1.1) and deploy"
    echo "  $0 minor     Bump minor version (0.1.0 -> 0.2.0) and deploy"
    echo "  $0 major     Bump major version (0.1.0 -> 1.0.0) and deploy"
    echo ""
    echo "Version bumping examples:"
    echo "  $0 patch     # 0.1.0 -> 0.1.1"
    echo "  $0 minor     # 0.1.0 -> 0.2.0" 
    echo "  $0 major     # 0.1.0 -> 1.0.0"
    exit 0
fi

# Check if we're in the right directory
if [ ! -f "setup.py" ] || [ ! -d "vittoriadb" ]; then
    print_error "Must be run from the SDK python directory"
    exit 1
fi

# Function to bump version
bump_version() {
    local bump_type="$1"
    local current_version
    
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
    case "$bump_type" in
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
    esac
    
    local new_version="$major.$minor.$patch"
    
    print_step "Bumping version: $current_version -> $new_version"
    
    # Update version in __init__.py
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS sed
        sed -i '' "s/__version__ = [\"'].*[\"']/__version__ = \"$new_version\"/" vittoriadb/__init__.py
    else
        # Linux sed
        sed -i "s/__version__ = [\"'].*[\"']/__version__ = \"$new_version\"/" vittoriadb/__init__.py
    fi
    
    print_success "Version updated to $new_version"
    
    # Confirm the change
    echo -e "${YELLOW}⚠️  Version will be bumped from $current_version to $new_version${NC}"
    read -p "Continue with deployment? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        # Revert the change
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/__version__ = [\"'].*[\"']/__version__ = \"$current_version\"/" vittoriadb/__init__.py
        else
            sed -i "s/__version__ = [\"'].*[\"']/__version__ = \"$current_version\"/" vittoriadb/__init__.py
        fi
        print_warning "Version bump cancelled, reverted to $current_version"
        exit 0
    fi
}

# Handle version bumping
case "$1" in
    "patch"|"minor"|"major")
        bump_version "$1"
        TARGET="production"
        REPO_FLAG=""
        print_warning "Bumping version and deploying to Production PyPI"
        ;;
    "test")
        TARGET="test"
        REPO_FLAG="--repository testpypi"
        print_warning "Deploying to Test PyPI"
        ;;
    *)
        TARGET="production"
        REPO_FLAG=""
        print_warning "Deploying to Production PyPI"
        ;;
esac

echo -e "${BLUE}🚀 VittoriaDB Python Package Deploy${NC}"
echo "=================================="

# 1. Clean
print_step "Cleaning build artifacts..."
rm -rf build/ dist/ *.egg-info/ __pycache__/ vittoriadb/__pycache__/
find . -name "*.pyc" -delete 2>/dev/null || true

# 2. Install build dependencies
print_step "Installing build dependencies..."
pip install --upgrade build twine

# 3. Build
print_step "Building package..."
python -m build

# 4. Check
print_step "Validating package..."
if [ ! -d "dist" ] || [ -z "$(ls -A dist/ 2>/dev/null)" ]; then
    print_error "No distribution files found. Build may have failed."
    exit 1
fi
python -m twine check dist/*

# 5. Get version for confirmation
VERSION=$(python -c "
import sys
sys.path.insert(0, '.')
with open('vittoriadb/__init__.py') as f:
    for line in f:
        if line.startswith('__version__'):
            print(line.split('=')[1].strip().strip('\"').strip(\"'\"))
            break
" 2>/dev/null || echo "unknown")

if [ "$VERSION" = "unknown" ]; then
    print_error "Could not determine package version"
    exit 1
fi

print_step "Ready to deploy vittoriadb v$VERSION to $TARGET PyPI"

# 6. Confirm (only for production)
if [ "$TARGET" = "production" ]; then
    echo -e "${YELLOW}⚠️  This will upload to PRODUCTION PyPI!${NC}"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Deploy cancelled"
        exit 0
    fi
fi

# 7. Upload
print_step "Uploading to $TARGET PyPI..."

# Check for authentication
if [ -z "$TWINE_USERNAME" ] && [ -z "$TWINE_PASSWORD" ] && [ ! -f "$HOME/.pypirc" ]; then
    print_warning "No PyPI credentials found!"
    echo ""
    echo "To authenticate, you can:"
    echo "1. Set environment variables:"
    echo "   export TWINE_USERNAME=__token__"
    echo "   export TWINE_PASSWORD=pypi-your-api-token-here"
    echo ""
    echo "2. Or create ~/.pypirc file:"
    echo "   [pypi]"
    echo "   username = __token__"
    echo "   password = pypi-your-api-token-here"
    echo ""
    echo "Get your API token from: https://pypi.org/manage/account/token/"
    echo ""
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_warning "Upload cancelled - please setup authentication first"
        exit 0
    fi
fi

python -m twine upload $REPO_FLAG dist/*

# 8. Success
print_success "Successfully deployed vittoriadb v$VERSION!"

if [ "$TARGET" = "test" ]; then
    echo -e "${BLUE}Test install: pip install -i https://test.pypi.org/simple/ vittoriadb==$VERSION${NC}"
else
    echo -e "${BLUE}Install: pip install vittoriadb${NC}"
fi
