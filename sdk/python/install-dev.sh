#!/bin/bash
# VittoriaDB Development Installation Script
# This installs the Python library in editable mode for development

set -e

echo "🚀 Installing VittoriaDB Python Library in Development Mode"
echo "=========================================================="

# Check if we're in the right directory
if [ ! -f "setup.py" ]; then
    echo "❌ Error: Please run this script from the python/ directory"
    exit 1
fi

# Check Python version
python_version=$(python3 -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
echo "📍 Python version: $python_version"

if [ "$(python3 -c "import sys; print(sys.version_info >= (3, 7))")" != "True" ]; then
    echo "❌ Error: Python 3.7+ required"
    exit 1
fi

# Install in editable mode
echo "📦 Installing VittoriaDB Python library in editable mode..."
pip3 install -e .

echo ""
echo "✅ Installation complete!"
echo ""
echo "📋 You can now use VittoriaDB in your Python code:"
echo "   import vittoriadb"
echo "   db = vittoriadb.connect()"
echo ""
echo "🔄 Changes to the library code will be immediately available"
echo "🧪 Run examples: python3 ../examples/basic_usage.py"
echo ""
