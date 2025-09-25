#!/bin/bash

# VittoriaDB Configuration Endpoint Demo
# This script demonstrates the new GET /config endpoint

echo "🔧 VittoriaDB Configuration Endpoint Demo"
echo "=========================================="

# Check if VittoriaDB is running
echo ""
echo "📡 1. Checking if VittoriaDB is running..."
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "   ✅ VittoriaDB is running"
else
    echo "   ❌ VittoriaDB is not running. Please start it with: ./vittoriadb run"
    exit 1
fi

# Test the configuration endpoint
echo ""
echo "🔧 2. Fetching current configuration..."
echo "   📡 GET http://localhost:8080/config"
echo ""

# Get configuration and format it nicely
CONFIG_RESPONSE=$(curl -s http://localhost:8080/config)

if [ $? -eq 0 ] && [ -n "$CONFIG_RESPONSE" ]; then
    echo "✅ Configuration retrieved successfully!"
    echo ""
    
    # Extract key information using jq if available
    if command -v jq > /dev/null 2>&1; then
        echo "📊 Key Configuration Summary:"
        echo "   • Config Source: $(echo "$CONFIG_RESPONSE" | jq -r '.metadata.source // "default"')"
        echo "   • Server: $(echo "$CONFIG_RESPONSE" | jq -r '.config.server.host'):$(echo "$CONFIG_RESPONSE" | jq -r '.config.server.port')"
        echo "   • Data Directory: $(echo "$CONFIG_RESPONSE" | jq -r '.config.data_dir')"
        echo ""
        
        echo "⚡ Performance Features:"
        echo "   • Parallel Search: $(echo "$CONFIG_RESPONSE" | jq -r '.features.parallel_search')"
        echo "   • Search Cache: $(echo "$CONFIG_RESPONSE" | jq -r '.features.search_cache')"
        echo "   • Memory-Mapped I/O: $(echo "$CONFIG_RESPONSE" | jq -r '.features.memory_mapped_io')"
        echo "   • SIMD Optimizations: $(echo "$CONFIG_RESPONSE" | jq -r '.features.simd_optimizations')"
        echo "   • Async I/O: $(echo "$CONFIG_RESPONSE" | jq -r '.features.async_io')"
        echo ""
        
        echo "📈 Performance Settings:"
        echo "   • Max Workers: $(echo "$CONFIG_RESPONSE" | jq -r '.performance.max_workers')"
        echo "   • Cache Entries: $(echo "$CONFIG_RESPONSE" | jq -r '.performance.cache_entries')"
        echo "   • Cache TTL: $(echo "$CONFIG_RESPONSE" | jq -r '.performance.cache_ttl')"
        echo "   • Max Concurrency: $(echo "$CONFIG_RESPONSE" | jq -r '.performance.max_concurrency')"
        echo ""
        
        echo "📄 Full Configuration (formatted):"
        echo "$CONFIG_RESPONSE" | jq .
    else
        echo "📄 Raw Configuration Response:"
        echo "$CONFIG_RESPONSE"
        echo ""
        echo "💡 Tip: Install 'jq' for formatted JSON output: brew install jq"
    fi
else
    echo "❌ Failed to retrieve configuration"
    exit 1
fi

echo ""
echo "🎉 Configuration endpoint demo completed!"
echo ""
echo "📚 Usage Examples:"
echo "   # Get full configuration"
echo "   curl http://localhost:8080/config"
echo ""
echo "   # Get specific feature status (with jq)"
echo "   curl -s http://localhost:8080/config | jq '.features.parallel_search'"
echo ""
echo "   # Get performance settings"
echo "   curl -s http://localhost:8080/config | jq '.performance'"
echo ""
echo "   # Get configuration metadata"
echo "   curl -s http://localhost:8080/config | jq '.metadata'"
