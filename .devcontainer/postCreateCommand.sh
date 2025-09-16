# Install dependencies and build
go mod download
go build -o vittoriadb ./cmd/vittoriadb

# Install Python SDK (optional)
cd sdk/python && ./install-dev.sh
cd ../..