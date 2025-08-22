# Switchaja Project Context

## Project Overview

Switchaja is a Go-based web server project that provides IoT switching capabilities. The project is structured as a modern Go application with embedded web assets.

## Architecture

- **Language**: Go 1.24.x
- **Module Name**: switchiot  
- **Main Entry Point**: `./cmd/server`
- **Web Assets**: Embedded using `static_embed.go`

## Key Directories

```
switchaja/
├── cmd/                 # Command-line applications
│   └── server/         # Main web server application
├── internal/           # Private application code
├── web/               # Web frontend assets
├── test/              # Test files and fixtures
├── deploy/            # Deployment configurations
└── static_embed.go    # Embedded web assets
```

## Development Guidelines

### Building
```bash
go build ./cmd/server
```

### Testing
```bash
go test ./...
```

### Dependencies
The project uses Go modules for dependency management. Run `go mod tidy` to sync dependencies.

## Web Server Features

- IoT device switching capabilities
- Web-based user interface
- Embedded static assets for standalone deployment
- RESTful API endpoints

## Development Environment

This project is configured with:
- Go 1.24.x runtime
- MCP server integration for enhanced development experience
- Automated build and test configurations
- CGO disabled for better portability

## Deployment

The project includes deployment configurations in the `deploy/` directory and supports containerized deployment scenarios.