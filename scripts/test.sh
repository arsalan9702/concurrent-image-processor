#!/bin/bash

set -e

echo "Running tests for Concurrent Image Processor..."

# Get project root directory
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)

# Create test directories
mkdir -p test/coverage

# Run tests with coverage
echo "Running unit tests with coverage..."
go test -v -race -coverprofile=test/coverage/coverage.out ./...

# Generate coverage report
echo "Generating coverage report..."
go tool cover -html=test/coverage/coverage.out -o test/coverage/coverage.html

# Display coverage summary
echo "Coverage summary:"
go tool cover -func=test/coverage/coverage.out

# Check for race conditions
echo "Running race condition tests..."
go test -race ./...

# Run benchmarks
echo "Running benchmarks..."
go test -bench=. -benchmem ./...

# Lint code (if golangci-lint is installed)
if command -v golangci-lint &> /dev/null; then
    echo "Running linter..."
    golangci-lint run
else
    echo "golangci-lint not found, skipping linting"
    echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# Run go vet
echo "Running go vet..."
go vet ./...

# Format check
echo "Checking code formatting..."
if [ -n "$(gofmt -l .)" ]; then
    echo "Code is not properly formatted. Run 'go fmt ./...' to fix."
    gofmt -l .
    exit 1
else
    echo "Code formatting is correct."
fi

# Check for unused dependencies
echo "Checking for unused dependencies..."
go mod tidy
if [ -n "$(git diff go.mod go.sum)" ]; then
    echo "go.mod or go.sum is not up to date. Run 'go mod tidy'."
    exit 1
else
    echo "Dependencies are up to date."
fi

echo "All tests passed successfully!"
echo "Coverage report: $(pwd)/test/coverage/coverage.html"