#!/bin/bash
# GitHub Copilot Agent Environment Setup Script
# This script ensures the development environment is properly configured
# for the switchaja IoT PlayStation Rental Management System

set -e

echo "Setting up GitHub Copilot agent environment for switchaja..."

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.24+ first."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "Go version: $GO_VERSION"

# Ensure MCP server is available
if ! command -v godoc-mcp &> /dev/null; then
    echo "Installing godoc-mcp MCP server..."
    go install github.com/mrjoshuak/godoc-mcp@latest
    echo "godoc-mcp installed successfully"
else
    echo "godoc-mcp is already installed"
fi

# Verify project dependencies
echo "Checking project dependencies..."
go mod verify
go mod tidy

# Check if we can build the project
echo "Testing project build..."
make build

echo "Environment setup complete!"
echo "Available tools:"
echo "  - Go: $(go version)"
echo "  - godoc-mcp: Available for Go documentation"
echo "  - make: Available for build automation"
echo ""
echo "Project structure:"
echo "  - cmd/server: Main application entry point"
echo "  - internal/: Core application logic with clean architecture"
echo "  - web/static: Frontend web assets"
echo "  - test/: Test files and utilities"
echo ""
echo "Key environment variables:"
echo "  - PORT: Server port (default: 8080)"
echo "  - DB_PATH: Database file path (default: heheswitch.db)" 
echo "  - MQTT_BROKER: MQTT broker URL for IoT integration"
echo ""
echo "Ready for GitHub Copilot agent assistance!"