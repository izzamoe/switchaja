#!/bin/bash
# GitHub Copilot Agent Environment Setup Script
# This script ensures the development environment is properly configured
# for the switchaja IoT PlayStation Rental Management System

set -e

echo "🚀 Setting up GitHub Copilot agent environment for switchaja..."

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed. Please install Go 1.24+ first."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go version: $GO_VERSION"

# Ensure MCP server is available
if ! command -v godoc-mcp &> /dev/null; then
    echo "📦 Installing godoc-mcp MCP server..."
    go install github.com/mrjoshuak/godoc-mcp@latest
    echo "✅ godoc-mcp installed successfully"
else
    echo "✅ godoc-mcp is already installed"
fi

# Verify project dependencies
echo "🔍 Checking project dependencies..."
if go mod verify; then
    echo "✅ Go modules verified"
else
    echo "⚠️  Go module verification failed, running go mod tidy..."
    go mod tidy
fi

# Check if we can build the project
echo "🔨 Testing project build..."
if make build; then
    echo "✅ Project builds successfully"
else
    echo "⚠️  Build failed - check for compilation errors"
fi

echo ""
echo "🎉 Environment setup complete!"
echo ""
echo "📋 Configuration Summary:"
echo "   • Agent config: .github/copilot/agent.yml"
echo "   • MCP server: godoc-mcp (Go documentation)"
echo "   • Build system: Makefile"
echo "   • File exclusions: .copilotignore"
echo ""
echo "🛠️  Available Development Tools:"
echo "   • Go: $(go version)"
echo "   • Make: $(make --version | head -n1)"
echo "   • godoc-mcp: Enhanced Go documentation support"
echo ""
echo "📁 Project Structure:"
echo "   • cmd/server/     → Main application entry point"
echo "   • internal/       → Core application logic (clean architecture)"
echo "   • web/static/     → Frontend web assets"
echo "   • test/          → Test files and utilities"
echo ""
echo "🔧 Key Environment Variables:"
echo "   • PORT           → Server port (default: 8080)"
echo "   • DB_PATH        → Database file path (default: heheswitch.db)"
echo "   • MQTT_BROKER    → MQTT broker URL for IoT integration"
echo "   • MQTT_PREFIX    → MQTT topic prefix (default: ps)"
echo ""
echo "🤖 GitHub Copilot agent is ready for enhanced assistance!"