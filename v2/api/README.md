# Kloudlite API v2

## Overview
Backend API server for Kloudlite v2 platform built with Go.

## Structure
```
api/
├── cmd/server/     # Main application entry point
├── pkg/test/       # Test utilities and helpers
└── Taskfile.yml    # Build automation
```

## Quick Start

### Prerequisites
- Go 1.21+

### Setup
```bash
# Install dependencies
task deps

# Run the server
task run
```

### Development
```bash
# Run with automatic reload
task dev

# Run tests
task test

# Format code
task fmt

# Run linters
task lint

# Show all available tasks
task --list
```

### Build
```bash
# Build binary
task build

# Build for production
task prod
```

## Configuration
Configuration is managed through environment variables. See `.env.example` for available options.

## API Endpoints

### Health
- `GET /health` - Health check
- `GET /ready` - Readiness check

### API v1
- `GET /api/v1/info` - API information