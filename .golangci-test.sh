#!/bin/bash
# Test script for golangci-lint configuration

set -e

echo "Testing golangci-lint configuration..."
echo ""

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "Error: golangci-lint is not installed"
    echo "Install it with: brew install golangci-lint"
    exit 1
fi

# Show version
echo "golangci-lint version:"
golangci-lint --version
echo ""

# Validate config
echo "Validating configuration..."
golangci-lint config verify
echo "Configuration is valid!"
echo ""

# Run linters
echo "Running linters..."
golangci-lint run --timeout=5m

echo ""
echo "Linting completed successfully!"
