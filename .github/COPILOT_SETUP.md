# GitHub Copilot Agent Environment Configuration

This repository is configured to customize GitHub Copilot's development environment following the [official GitHub documentation best practices](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/customize-the-agent-environment).

## Overview

GitHub Copilot has access to its own ephemeral development environment, powered by GitHub Actions, where it can explore code, make changes, execute automated tests, and run linters. This repository customizes that environment to provide optimal support for Go development.

## Configuration Files

### `.github/workflows/copilot-setup-steps.yml`
This is the main configuration file that defines how Copilot's environment should be set up. The workflow:

- **Installs Go 1.24** with caching for faster subsequent runs
- **Downloads Go dependencies** using `go mod download`
- **Builds the application** using the project's Makefile
- **Runs tests** to ensure the environment is properly configured
- **Installs additional Go tools** like `goimports`, `golint`, and `golangci-lint`

### `.copilotignore`
Specifies files and directories that should be excluded from Copilot's context, including:
- Build artifacts (`dist/`, binaries)
- Dependencies (`vendor/`, `node_modules/`)
- Temporary files (`.tmp`, `.log`)
- Database files (`.db`, `.sqlite`)
- IDE files (`.vscode/`, `.idea/`)

## Environment Setup

The setup workflow automatically:

1. **Checks out the repository code**
2. **Sets up Go 1.24** with module caching enabled
3. **Downloads all Go dependencies** specified in `go.mod`
4. **Builds the application** using the existing Makefile
5. **Runs the test suite** to validate the environment
6. **Installs additional development tools** for code quality

## Benefits for Copilot

This configuration enables GitHub Copilot to:

- **Understand project dependencies** through pre-downloaded modules
- **Build and test changes** before suggesting them
- **Use Go development tools** for code analysis and formatting
- **Access project context** without build artifacts cluttering the view
- **Work efficiently** with a pre-configured, ready-to-use environment

## Project Context

### Technology Stack
- **Go 1.24+** - Primary language
- **Fiber web framework** - HTTP server and routing
- **SQLite** - Embedded database with WAL mode
- **MQTT** - IoT device communication protocol
- **WebSocket** - Real-time client updates

### Architecture
This is an IoT PlayStation rental management system with:
- **Clean architecture** with separation of concerns
- **IoT integration** via MQTT for device control
- **Real-time features** using WebSocket connections
- **Web interface** for management and monitoring

### Key Directories
- `cmd/server/` - Application entry point
- `internal/api/` - HTTP API handlers and routes
- `internal/iot/` - MQTT and IoT device integration
- `internal/db/` - Database models and operations
- `web/static/` - Frontend web assets
- `test/` - Test files and utilities

## Validation

The setup workflow runs automatically when:
- The workflow file is modified
- Changes are pushed to the main branch
- Manually triggered from the Actions tab

You can monitor the setup process in the repository's Actions tab to ensure everything works correctly.

## Official Documentation

This configuration follows the official GitHub documentation:
- [Customizing the development environment for GitHub Copilot coding agent](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/customize-the-agent-environment)
- [Workflow syntax for GitHub Actions](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)

## Troubleshooting

If the setup fails:
1. Check the Actions tab for detailed error logs
2. Ensure all dependencies in `go.mod` are accessible
3. Verify the build process works locally with `make build`
4. Test the workflow manually using the "Run workflow" button in Actions

The workflow is designed to be self-validating and will show any issues in the GitHub Actions logs.