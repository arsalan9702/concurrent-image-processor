#!/bin/bash

set -e

echo "Building Concurrent Image Processor..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Get project root directory
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)

# Set build information
VERSION=$(git describe --tags --dirty --always 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | cut -d' ' -f3)

# Build flags
LDFLAGS="-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GoVersion=$GO_VERSION"

echo "Building version: $VERSION"
echo "Build time: $BUILD_TIME"
echo "Go version: $GO_VERSION"

# Build for current platform
echo "Building for current platform..."
go build -ldflags "$LDFLAGS" -o bin/processor cmd/processor/main.go

# Build for multiple platforms (optional)
if [ "$1" = "all" ]; then
    echo "Building for multiple platforms..."
    
    # Linux AMD64
    echo "Building for Linux AMD64..."
    GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/processor-linux-amd64 cmd/processor/main.go
    
    # Linux ARM64
    echo "Building for Linux ARM64..."
    GOOS=linux GOARCH=arm64 go build -ldflags "$LDFLAGS" -o bin/processor-linux-arm64 cmd/processor/main.go
    
    # Windows AMD64
    echo "Building for Windows AMD64..."
    GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/processor-windows-amd64.exe cmd/processor/main.go
    
    # macOS AMD64
    echo "Building for macOS AMD64..."
    GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/processor-darwin-amd64 cmd/processor/main.go
    
    # macOS ARM64 (Apple Silicon)
    echo "Building for macOS ARM64..."
    GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o bin/processor-darwin-arm64 cmd/processor/main.go
fi

echo "Build completed successfully!"
echo "Binary location: $(pwd)/bin/processor"

# Make the binary executable
chmod +x bin/processor*

echo "To run the processor:"
echo "  ./bin/processor -help"