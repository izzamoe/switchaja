# GitHub Copilot Environment Configuration

This directory contains the configuration files for GitHub Copilot Agent to properly understand and work with the switchaja project.

## Structure

```
.copilot/
├── environment.yml    # Main environment configuration
├── README.md         # This documentation file
└── context.md        # Project context for Copilot
```

## Files

### environment.yml
The main environment configuration file that defines:
- Go development environment setup
- MCP server integration for enhanced documentation
- Build and test configurations
- Required tools and dependencies

### context.md
Additional project context to help Copilot understand the codebase structure and purpose.

## Usage

This configuration is automatically recognized by GitHub Copilot Agent and used to:
1. Set up the proper Go development environment
2. Install and configure the Go documentation MCP server
3. Provide enhanced code assistance with project-specific context
4. Enable proper build and test operations

## MCP Server Integration

The configuration includes setup for the Go documentation MCP server (`godoc-mcp`) which provides:
- Enhanced Go documentation access
- Symbol and function lookup
- Code completion assistance
- Project-aware suggestions

This server is automatically installed and configured when the environment is initialized.