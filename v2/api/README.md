# Kloudlite API v2

Backend API server using Kubernetes CRDs as persistence layer.

## Quick Start

**Prerequisites:**
- Go 1.24+
- Docker with colima context
- kubectl

**Setup:**
```bash
# Start K3s cluster
docker context use colima
docker-compose up -d k3s

# Run API server
go run ./cmd/server
```

**API Endpoints:**
- `GET /health` - Health check
- `GET /api/v1/users` - List users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:name` - Get user
- `PUT /api/v1/users/:name` - Update user
- `DELETE /api/v1/users/:name` - Delete user

**Environment Variables:**
See `.env` file for Kubernetes configuration options.