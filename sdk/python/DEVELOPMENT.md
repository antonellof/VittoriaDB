# VittoriaDB Python SDK - Development Guide

This guide is for contributors and maintainers who want to develop, build, and deploy the VittoriaDB Python SDK.

## 🛠️ Development Setup

### Prerequisites
- Python 3.7+
- Git
- PyPI account (for publishing)

### Installation for Development

```bash
# Clone the repository
git clone https://github.com/antonellof/VittoriaDB.git
cd VittoriaDB/sdk/python

# Install in development mode
pip install -e .

# Or use the development script
./install-dev.sh
```

This installs the package in editable mode, so changes to the code are immediately available.

## 🚀 Building and Publishing

### Quick Deploy Commands

**🚀 One-Command Deploy with Version Bumping:**
```bash
# Deploy current version to Test PyPI
./deploy.sh test

# Bump patch version (0.1.0 -> 0.1.1) and deploy to Production
./deploy.sh patch

# Bump minor version (0.1.0 -> 0.2.0) and deploy to Production
./deploy.sh minor

# Bump major version (0.1.0 -> 1.0.0) and deploy to Production
./deploy.sh major

# Deploy current version to Production PyPI (no version bump)
./deploy.sh
```

### Version Management

**📦 Standalone Version Bumping:**
```bash
# Just bump version without deploying
./bump-version.sh patch   # 0.1.0 -> 0.1.1
./bump-version.sh minor   # 0.1.0 -> 0.2.0
./bump-version.sh major   # 0.1.0 -> 1.0.0
```

**🔍 Check Current Version:**
```bash
python -c "import vittoriadb; print(vittoriadb.__version__)"
```

### PyPI Authentication Setup

**🔑 First-time Setup:**
```bash
# Interactive authentication setup
./setup-pypi.sh
```

**Manual Setup Options:**

1. **Environment Variables (temporary):**
```bash
export TWINE_USERNAME=__token__
export TWINE_PASSWORD=pypi-your-api-token-here
```

2. **~/.pypirc File (permanent):**
```bash
cat > ~/.pypirc << EOF
[distutils]
index-servers = pypi testpypi

[pypi]
username = __token__
password = pypi-your-api-token-here

[testpypi]
repository = https://test.pypi.org/legacy/
username = __token__
password = pypi-your-api-token-here
EOF
chmod 600 ~/.pypirc
```

Get your API token from: https://pypi.org/manage/account/token/

## 🔄 Development Workflow

### Recommended Development Process

1. **Setup Development Environment:**
```bash
./install-dev.sh
./setup-pypi.sh  # First time only
```

2. **Make Your Changes:**
- Edit code in `vittoriadb/` directory
- Update version in `vittoriadb/__init__.py` if needed
- Update `README.md` for user-facing changes
- Update this `DEVELOPMENT.md` for dev changes

3. **Test Your Changes:**
```bash
# Test with existing VittoriaDB server
python -c "import vittoriadb; print('✅ Import works')"

# Test basic functionality
cd ../../examples/python
python 00_basic_usage_manual_vectors.py
```

4. **Deploy to Test PyPI:**
```bash
./deploy.sh test
```

5. **Test Installation from Test PyPI:**
```bash
pip install -i https://test.pypi.org/simple/ vittoriadb
```

6. **Deploy to Production:**
```bash
./deploy.sh patch  # or minor/major for version bumps
```

### Version Bumping Guidelines

Follow [Semantic Versioning](https://semver.org/):

- **Patch** (`0.1.0` → `0.1.1`): Bug fixes, small improvements
- **Minor** (`0.1.0` → `0.2.0`): New features, backward compatible
- **Major** (`0.1.0` → `1.0.0`): Breaking changes

### What the Deploy Script Does

**✨ The deploy script automatically:**
- Bumps version (if specified)
- Cleans build artifacts
- Installs build dependencies (`build`, `twine`)
- Builds the package (source + wheel)
- Validates the package with twine
- Checks for PyPI authentication
- Uploads to PyPI with confirmation prompts
- Shows install instructions

## 🧪 Testing

### Manual Testing

```bash
# Test import
python -c "import vittoriadb; print(vittoriadb.__version__)"

# Test connection (requires running VittoriaDB server)
python -c "
import vittoriadb
db = vittoriadb.connect(auto_start=False)
print('✅ Connection works')
db.close()
"
```

### Integration Testing

```bash
# Run example scripts
cd ../../examples/python
python 02_server_side_embeddings_basic.py
python 07_rag_complete_workflow.py
```

## 🔧 Troubleshooting

### Common Issues

**Import Errors:**
```bash
# Reinstall in development mode
pip uninstall vittoriadb
pip install -e .
```

**Authentication Errors:**
```bash
# Check authentication
./setup-pypi.sh  # Choose option 3 to check status

# Reset authentication
rm ~/.pypirc
./setup-pypi.sh
```

**Version Already Exists:**
```bash
# Bump version before deploying
./deploy.sh patch  # or minor/major
```

**Build Errors:**
```bash
# Clean and retry
rm -rf build/ dist/ *.egg-info/
./deploy.sh test
```

## 📁 File Structure

```
sdk/python/
├── vittoriadb/              # Main package
│   ├── __init__.py         # Package exports and version
│   ├── client.py           # Main VittoriaDB client
│   ├── types.py            # Data types and enums
│   └── configure.py        # Configuration builders
├── setup.py                # Package configuration
├── README.md               # User documentation (PyPI)
├── DEVELOPMENT.md          # This file (dev guide)
├── install-dev.sh          # Development setup script
├── deploy.sh               # Main deploy script
├── bump-version.sh         # Version bumping only
└── setup-pypi.sh           # PyPI authentication setup
```

## 🤝 Contributing Guidelines

### Code Style
- Follow PEP 8 style guidelines
- Use type hints where appropriate
- Add docstrings to public methods
- Keep imports organized

### Commit Messages
- Use clear, descriptive commit messages
- Reference issue numbers when applicable
- Use conventional commit format when possible

### Pull Requests
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Update documentation
6. Submit pull request

## 🔗 Useful Links

- **Main Repository**: https://github.com/antonellof/VittoriaDB
- **PyPI Project**: https://pypi.org/project/vittoriadb/
- **Test PyPI**: https://test.pypi.org/project/vittoriadb/
- **PyPI API Tokens**: https://pypi.org/manage/account/token/
- **Twine Documentation**: https://twine.readthedocs.io/
- **Python Packaging Guide**: https://packaging.python.org/

---

**Happy developing! 🚀**
