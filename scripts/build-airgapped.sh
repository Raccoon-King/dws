#!/bin/bash

# Build script for air-gapped environments
# This script ensures all dependencies are vendored before building

set -e

echo "=== Building DWS for Air-Gapped Environments ==="

# Check if vendor directory exists
if [ ! -d "vendor" ]; then
    echo "ERROR: vendor directory not found!"
    echo "Run 'go mod vendor' in a networked environment first"
    exit 1
fi

echo "✓ Vendor directory found"

# Verify Go modules
if ! go mod verify; then
    echo "ERROR: Go module verification failed"
    exit 1
fi

echo "✓ Go modules verified"

# Build binary using vendored dependencies
echo "Building binary..."
CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -o dws .

echo "✓ Binary built successfully: ./dws"

# Build Docker image using airgapped Dockerfile
if command -v docker >/dev/null 2>&1; then
    echo "Building Docker image for air-gapped deployment..."
    docker build -f Dockerfile.airgapped -t dws:airgapped .
    echo "✓ Docker image built: dws:airgapped"
else
    echo "⚠ Docker not available, skipping image build"
fi

echo ""
echo "=== Air-Gapped Build Complete ==="
echo "Binary: ./dws"
echo "Docker image: dws:airgapped (if Docker available)"
echo ""
echo "To run the binary directly:"
echo "  export RULES_FILE=rules.yaml"
echo "  export PORT=8080" 
echo "  ./dws"
echo ""
echo "To run with Docker:"
echo "  docker run -d -p 8080:8080 -v ./rules.yaml:/app/rules.yaml dws:airgapped"