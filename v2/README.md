# Kloudlite v2 Development Setup

## Prerequisites

- Docker Desktop installed and running
- Go 1.21+ installed
- Node.js 18+ and pnpm installed

## Step 1: Setup Infrastructure

Start K3s and auto-configure everything:

```bash
docker-compose up -d
```

This automatically:
- Starts K3s cluster
- Installs User CRD
- Configures RBAC
- Generates kubeconfig files

## Step 2: Run Backend API

```bash
cd v2/api
task run
```

The API server will start on http://localhost:8080

## Step 3: Run Frontend

```bash
cd v2/web
pnpm install  # First time only
pnpm dev
```

The frontend will start on http://localhost:3000

## Quick Commands

```bash
# Start everything
docker-compose up -d

# Stop everything
docker-compose down

# Clean everything (including data)
docker-compose down -v

# Check logs
docker-compose logs -f

# Access K3s directly
kubectl --kubeconfig=v2/api/kubeconfig/k3s.yaml get nodes
```

## Architecture

```
Frontend (3000) → Backend API (8080) → K3s (6443)
```

## Configuration Files

- `docker-compose.yml` - Infrastructure setup
- `v2/api/kubeconfig/` - Auto-generated kubeconfig files (git-ignored)
- `v2/api/.env` - Backend configuration
- `v2/web/.env` - Frontend configuration