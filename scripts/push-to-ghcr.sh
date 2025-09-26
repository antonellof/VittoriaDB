#!/bin/bash

# VittoriaDB - Push Docker Images to GitHub Container Registry
# This script builds and pushes VittoriaDB images to ghcr.io

set -e

VERSION=${1:-"v0.5.0"}
REPO="antonellof/vittoriadb"

echo "🚀 Pushing VittoriaDB $VERSION to GitHub Container Registry"
echo "=========================================================="

# Check if we're logged in to ghcr.io
echo "🔐 Checking GitHub Container Registry authentication..."
if ! docker info | grep -q "ghcr.io"; then
    echo "⚠️  Not logged in to ghcr.io. Please authenticate first:"
    echo ""
    echo "1. Create a Personal Access Token with 'write:packages' permission:"
    echo "   https://github.com/settings/tokens/new?scopes=write:packages"
    echo ""
    echo "2. Login to ghcr.io:"
    echo "   echo \$GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin"
    echo ""
    read -p "Press Enter after authenticating..."
fi

# Build the image
echo "🔨 Building VittoriaDB $VERSION..."
docker build -t ghcr.io/$REPO:$VERSION -t ghcr.io/$REPO:latest .

# Push the images
echo "📤 Pushing images to ghcr.io..."
docker push ghcr.io/$REPO:$VERSION
docker push ghcr.io/$REPO:latest

echo ""
echo "✅ Successfully pushed VittoriaDB images to GitHub Container Registry!"
echo ""
echo "📋 Available images:"
echo "   • ghcr.io/$REPO:$VERSION"
echo "   • ghcr.io/$REPO:latest"
echo ""
echo "🔗 View on GitHub:"
echo "   https://github.com/$REPO/pkgs/container/vittoriadb"
echo ""
echo "🚀 Ready for cloud deployment!"
