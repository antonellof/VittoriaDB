#!/bin/bash
# PyPI Authentication Setup Helper

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

echo -e "${BLUE}🔑 PyPI Authentication Setup${NC}"
echo "=============================="
echo ""

# Check current status
print_step "Checking current authentication status..."

if [ -f "$HOME/.pypirc" ]; then
    print_success "Found ~/.pypirc file"
elif [ -n "$TWINE_USERNAME" ] && [ -n "$TWINE_PASSWORD" ]; then
    print_success "Found environment variables"
else
    print_warning "No authentication configured"
fi

echo ""
echo "🔐 PyPI Authentication Options:"
echo ""
echo "1. Environment Variables (temporary)"
echo "2. ~/.pypirc file (permanent)"
echo "3. Show current status"
echo "4. Exit"
echo ""

read -p "Choose option (1-4): " choice

case $choice in
    1)
        echo ""
        print_step "Setting up environment variables..."
        echo ""
        echo "First, get your API token from: https://pypi.org/manage/account/token/"
        echo ""
        read -p "Enter your PyPI API token (starts with pypi-): " token
        
        if [ -z "$token" ]; then
            print_error "No token provided"
            exit 1
        fi
        
        echo ""
        echo "Add these lines to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo ""
        echo "export TWINE_USERNAME=__token__"
        echo "export TWINE_PASSWORD=$token"
        echo ""
        echo "Or run these commands for this session:"
        echo ""
        echo -e "${GREEN}export TWINE_USERNAME=__token__${NC}"
        echo -e "${GREEN}export TWINE_PASSWORD=$token${NC}"
        echo ""
        ;;
        
    2)
        echo ""
        print_step "Setting up ~/.pypirc file..."
        echo ""
        echo "First, get your API token from: https://pypi.org/manage/account/token/"
        echo ""
        read -p "Enter your PyPI API token (starts with pypi-): " token
        
        if [ -z "$token" ]; then
            print_error "No token provided"
            exit 1
        fi
        
        # Backup existing file if it exists
        if [ -f "$HOME/.pypirc" ]; then
            cp "$HOME/.pypirc" "$HOME/.pypirc.backup"
            print_warning "Backed up existing ~/.pypirc to ~/.pypirc.backup"
        fi
        
        # Create .pypirc file
        cat > "$HOME/.pypirc" << EOF
[distutils]
index-servers = pypi testpypi

[pypi]
username = __token__
password = $token

[testpypi]
repository = https://test.pypi.org/legacy/
username = __token__
password = $token
EOF
        
        chmod 600 "$HOME/.pypirc"
        print_success "Created ~/.pypirc file with secure permissions"
        ;;
        
    3)
        echo ""
        print_step "Current authentication status:"
        echo ""
        
        if [ -f "$HOME/.pypirc" ]; then
            print_success "~/.pypirc file exists"
            echo "Content (passwords hidden):"
            sed 's/password = .*/password = [HIDDEN]/' "$HOME/.pypirc"
        else
            print_warning "No ~/.pypirc file found"
        fi
        
        echo ""
        
        if [ -n "$TWINE_USERNAME" ]; then
            print_success "TWINE_USERNAME = $TWINE_USERNAME"
        else
            print_warning "TWINE_USERNAME not set"
        fi
        
        if [ -n "$TWINE_PASSWORD" ]; then
            print_success "TWINE_PASSWORD = [HIDDEN]"
        else
            print_warning "TWINE_PASSWORD not set"
        fi
        ;;
        
    4)
        echo "Goodbye!"
        exit 0
        ;;
        
    *)
        print_error "Invalid option"
        exit 1
        ;;
esac

echo ""
print_success "Setup complete!"
echo ""
echo "💡 Next steps:"
echo "1. Test with: ./deploy.sh test"
echo "2. Deploy to production: ./deploy.sh"
echo ""
echo "📚 Useful links:"
echo "• PyPI API tokens: https://pypi.org/manage/account/token/"
echo "• Test PyPI: https://test.pypi.org/"
echo "• Twine docs: https://twine.readthedocs.io/"
