# GitHub Copilot Agent Configuration

This directory contains configuration files for customizing the GitHub Copilot agent environment for the switchaja project.

## Files

### `agent.yml`
Main configuration file that defines:
- **MCP Servers**: Configures the `godoc-mcp` server for Go documentation context
- **Tools**: Specifies available development tools (Go, Make)
- **Context**: Project-specific information for better AI assistance
- **Environment Variables**: Important configuration options for the application
- **Code Conventions**: Development standards and best practices

### `setup.sh`
Environment setup script that:
- Verifies Go installation and version
- Installs the required MCP server (`godoc-mcp`)
- Validates project dependencies
- Tests the build process
- Provides environment overview

### `../.copilotignore`
Specifies files and directories to exclude from Copilot analysis:
- Build artifacts and binaries
- Dependencies and vendor directories
- Temporary files and logs
- IDE-specific files

## MCP Integration

The configuration includes the `godoc-mcp` server which provides:
- Go package documentation lookup
- Standard library reference
- Project-specific Go code context
- Enhanced code completion and suggestions

## Usage

To set up the environment manually:
```bash
# Run the setup script
./.github/copilot/setup.sh

# Or install the MCP server directly
go install github.com/mrjoshuak/godoc-mcp@latest
```

## Project Context

The agent is configured with knowledge of:
- **Architecture**: Clean architecture with separate layers
- **Technologies**: Go, Fiber, SQLite, MQTT, WebSocket
- **Purpose**: IoT PlayStation rental management system
- **Key Components**: Device control, real-time updates, rental management

## Environment Variables

The agent understands these key configuration variables:
- `PORT`: Server port (default: 8080)
- `DB_PATH`: Database file path
- `SQLITE_MODE`: Database performance mode
- `MQTT_*`: IoT device communication settings

This configuration enhances the GitHub Copilot experience by providing project-specific context and Go language support through the MCP integration.