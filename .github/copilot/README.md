# GitHub Copilot Agent Configuration

This directory contains configuration files for customizing the GitHub Copilot agent environment for the switchaja project, following the [official best practices](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/customize-the-agent-environment).

## Files Overview

### `agent.yml`
Main configuration file that defines the agent environment according to GitHub's specification:

- **MCP Servers**: Configures the `godoc-mcp` server for enhanced Go documentation context
- **Project Information**: Defines project type, description, and key technologies
- **Development Environment**: Specifies tools, key files, environment variables, and conventions
- **Project Structure**: Maps important directories and their purposes

### `setup.sh`
Automated environment setup script that:
- ✅ Verifies Go installation and version compatibility
- 📦 Installs the required MCP server (`godoc-mcp`)
- 🔍 Validates project dependencies with `go mod verify`
- 🔨 Tests the build process to ensure everything works
- 📋 Provides a comprehensive environment overview

### `../.copilotignore`
File exclusion rules to optimize Copilot performance by ignoring:
- Build artifacts and compiled binaries
- Dependencies and vendor directories  
- Temporary files, logs, and IDE configurations
- Generated code and large data files

## MCP Integration

The configuration includes the `godoc-mcp` server which provides:
- 📚 Go package documentation lookup
- 🔍 Standard library reference and examples
- 🧠 Project-specific Go code context
- ⚡ Enhanced code completion and intelligent suggestions

## Quick Start

Run the setup script to configure your environment:

```bash
# Make the script executable and run it
chmod +x .github/copilot/setup.sh
./.github/copilot/setup.sh
```

Or install the MCP server manually:
```bash
go install github.com/mrjoshuak/godoc-mcp@latest
```

## Project Context Enhancement

The agent is configured with comprehensive knowledge of:

### 🏗️ Architecture
- **Clean Architecture**: Separate layers for better maintainability
- **IoT Integration**: MQTT-based device communication
- **Web Interface**: Real-time updates with WebSockets

### 🛠️ Technologies
- **Backend**: Go 1.24+, Fiber web framework
- **Database**: SQLite with optimized configurations
- **IoT**: MQTT protocol for PlayStation device control
- **Frontend**: WebSocket for real-time rental management

### 📁 Project Structure
```
switchaja/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── api/             # HTTP API handlers and routes
│   ├── iot/             # MQTT and IoT device integration
│   └── db/              # Database models and operations
├── web/static/          # Frontend web assets
└── test/                # Test files and utilities
```

## Environment Configuration

The agent understands these key configuration variables:

### 🌐 Server Settings
- `PORT`: Server port (default: 8080)
- `DB_PATH`: Database file path (default: heheswitch.db)
- `SQLITE_MODE`: Database performance mode (default: balanced)

### 📡 IoT/MQTT Settings
- `MQTT_BROKER`: MQTT broker URL for device communication
- `MQTT_PREFIX`: Topic prefix for organizing messages (default: ps)
- `MQTT_USERNAME/PASSWORD`: Authentication credentials
- `MQTT_CLIENT_ID`: Unique client identifier

## Development Best Practices

The configuration promotes these conventions:
- ✅ Use Go modules for dependency management
- 🏗️ Follow clean architecture principles
- 📝 Implement structured logging for debugging
- 🛡️ Comprehensive error handling throughout
- 🧪 Write unit tests for business logic
- ⚙️ Use environment variables for configuration
- 📋 Follow Go naming conventions and idioms

This setup enhances the GitHub Copilot experience by providing rich project context and specialized Go language support through MCP integration, resulting in more accurate and contextually relevant code suggestions.