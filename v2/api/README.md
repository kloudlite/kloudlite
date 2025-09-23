# Kloudlite API v2

## Overview
Backend API server for Kloudlite v2 platform built with Go and clean architecture.

## Structure
```
api/
├── cmd/
│   └── server/         # Application entry points
├── internal/
│   ├── domain/         # Business logic and entities
│   ├── app/           # Application layer (use cases)
│   └── infra/         # Infrastructure layer (k8s, external services)
├── pkg/
│   ├── utils/         # Shared utilities
│   └── errors/        # Error handling
└── Taskfile.yml       # Build automation
```

## Quick Start

### Prerequisites
- Go 1.21+
- Kubernetes cluster (local K3s or remote)
- kubectl configured

### Setup
```bash
# Initial setup
task setup

# Or manually:
cp .env.example .env
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

# Build Docker image
task docker-build
```

## Configuration
Configuration is managed through environment variables. See `.env.example` for available options.

## API Endpoints

### Health
- `GET /health` - Health check
- `GET /ready` - Readiness check

### API v1
- `GET /api/v1/info` - API information