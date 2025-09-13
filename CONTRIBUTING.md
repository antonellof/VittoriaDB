# Contributing to VittoriaDB

Thank you for your interest in contributing to VittoriaDB! We welcome contributions from everyone, whether you're fixing a bug, adding a feature, improving documentation, or helping with testing.

## 🚀 Quick Start

1. **Fork** the repository on GitHub
2. **Clone** your fork locally
3. **Create** a new branch for your changes
4. **Make** your changes and test them
5. **Submit** a pull request

## 📋 Ways to Contribute

### 🐛 Bug Reports
- Use the [GitHub Issues](https://github.com/antonellof/VittoriaDB/issues) page
- Search existing issues first to avoid duplicates
- Include clear steps to reproduce the bug
- Provide system information (OS, Go version, etc.)

### ✨ Feature Requests
- Open an issue with the "enhancement" label
- Describe the feature and its use case
- Discuss the implementation approach if you have ideas

### 🔧 Code Contributions
- Bug fixes
- New features
- Performance improvements
- Documentation updates
- Test improvements

### 📖 Documentation
- Fix typos or unclear explanations
- Add examples or tutorials
- Improve API documentation
- Translate documentation (future)

## 🛠️ Development Setup

### Prerequisites
- **Go 1.21+** for the main codebase
- **Python 3.7+** for the Python SDK
- **Git** for version control

### Setup Steps
```bash
# 1. Fork the repo on GitHub, then clone your fork
git clone https://github.com/YOUR_USERNAME/VittoriaDB.git
cd VittoriaDB

# 2. Add the original repo as upstream
git remote add upstream https://github.com/antonellof/VittoriaDB.git

# 3. Create a new branch
git checkout -b feature/your-feature-name

# 4. Install dependencies and build
go mod download
go build -o vittoriadb ./cmd/vittoriadb

# 5. Install Python SDK (optional)
cd sdk/python && ./install-dev.sh
```

## ✅ Before Submitting

### Code Quality
- [ ] Code follows Go and Python style guidelines
- [ ] All tests pass: `go test ./...`
- [ ] New features include tests
- [ ] Documentation is updated if needed
- [ ] Commit messages are clear and descriptive

### Testing
```bash
# Run Go tests
go test ./... -v

# Run Python tests (if applicable)
cd sdk/python && python -m pytest tests/ -v

# Test your changes manually
./vittoriadb run
curl http://localhost:8080/health
```

## 📝 Pull Request Process

1. **Update** your branch with the latest changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push** your changes:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create** a pull request on GitHub with:
   - Clear title describing the change
   - Description of what was changed and why
   - Link to any related issues
   - Screenshots/examples if applicable

4. **Respond** to review feedback promptly
5. **Squash** commits if requested before merging

## 🎯 Contribution Guidelines

### Code Style
- **Go**: Follow `gofmt` and `golangci-lint` standards
- **Python**: Follow PEP 8 and use type hints
- **Comments**: Write clear, helpful comments for complex logic
- **Naming**: Use descriptive variable and function names

### Commit Messages
Use clear, descriptive commit messages:
```
feat: add HNSW index support for large collections
fix: resolve memory leak in vector search
docs: update API documentation for batch operations
test: add integration tests for Python client
```

### Branch Naming
Use descriptive branch names:
- `feature/add-hnsw-index`
- `fix/memory-leak-search`
- `docs/update-api-guide`
- `test/integration-python`

## 🏷️ Issue Labels

We use these labels to organize issues:
- **bug**: Something isn't working
- **enhancement**: New feature or request
- **documentation**: Improvements to docs
- **good first issue**: Good for newcomers
- **help wanted**: Extra attention is needed
- **question**: Further information is requested

## 🤝 Code of Conduct

### Our Standards
- **Be respectful** and inclusive
- **Be constructive** in feedback
- **Be patient** with newcomers
- **Be collaborative** and helpful

### Not Acceptable
- Harassment or discriminatory language
- Personal attacks or trolling
- Spam or off-topic content
- Publishing private information

## 🆘 Getting Help

### Questions?
- **GitHub Discussions**: For general questions and ideas
- **GitHub Issues**: For bug reports and feature requests
- **Documentation**: Check the [`docs/`](docs/) directory first

### Need Support?
- Review the [Development Guide](docs/development.md)
- Check existing issues and discussions
- Ask questions in GitHub Discussions
- Tag maintainers in issues if urgent

## 🎉 Recognition

Contributors are recognized in:
- GitHub contributors list
- Release notes for significant contributions
- Special thanks in documentation

## 📄 License

By contributing to VittoriaDB, you agree that your contributions will be licensed under the same [MIT License](LICENSE) that covers the project.

---

**Thank you for contributing to VittoriaDB! 🚀**

*Every contribution, no matter how small, makes a difference and is greatly appreciated.*
