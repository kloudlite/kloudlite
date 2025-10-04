# Development Setup

## Prerequisites

- Go 1.24+
- Node.js 18+
- pnpm
- Docker & Docker Compose
- kubectl
- [Task](https://taskfile.dev) (install: `brew install go-task` or see [installation guide](https://taskfile.dev/installation/))

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/kloudlite/kloudlite.git
cd kloudlite
```

### 2. Install Frontend Dependencies

```bash
cd devenv
task web:install
```

### 3. Start Infrastructure (K3s + Pre-setup)

**Terminal 1:**

```bash
cd devenv
docker-compose up k3s pre-app
```

Wait for pre-app to complete:
- ✓ K3s cluster starts
- ✓ Kubeconfig extracted to `devenv/k3s-config/k3s.yaml`
- ✓ CRDs installed
- ✓ Pre-app exits with "PRE-APP SETUP COMPLETE"

### 4. Start Backend API

**Terminal 2:**

```bash
cd devenv
task api:dev
```

Wait for API to start on `http://localhost:8080`

### 5. Complete Post-setup

**Back to Terminal 1:**

```bash
docker-compose up post-app
```

This will:
- Generate TLS certificates
- Deploy nginx in k3s
- Wait for backend API (already running)
- Create default users and machine types
- Configure admission webhooks

Wait for "POST-APP SETUP COMPLETE"

### 6. Start Frontend

**Terminal 3:**

```bash
cd devenv
task web:dev
```

Frontend runs on `http://localhost:3000`

## Default Users

| Email | Password | Role |
|-------|----------|------|
| super-admin@kloudlite.io | AdminPass123! | super-admin |
| admin@kloudlite.io | AdminPass123! | admin |
| user@kloudlite.io | DevPass123! | user |

## Useful Commands

### Development

```bash
# Check cluster status
task status

# View k3s logs
docker logs k3s-dev -f

# Restart infrastructure
docker-compose down
docker-compose up k3s pre-app -d

# Stop all containers
docker-compose down

# Clean up everything (removes volumes)
docker-compose down -v
```

### API

```bash
# Run tests
task api:test

# Run with coverage
task api:test

# Lint code
task api:lint

# Format code
task api:fmt

# Generate CRDs
task api:generate-crds

# Generate deepcopy code
task api:generate-deepcopy
```

### Testing

Access the application at `http://localhost:3000` and login with any of the default users.

## Troubleshooting

### Post-app stuck waiting for API

If post-app is stuck at "Waiting for API...", make sure:
1. Backend API is running on port 8080
2. Check API logs for errors

```bash
# Check if API is running
curl http://localhost:8080/health
```

### K3s not ready

```bash
docker-compose down
docker-compose up k3s pre-app
```

### Port conflicts

Kill processes on ports 3000, 6443, or 8080:

```bash
lsof -ti:3000 | xargs kill -9
lsof -ti:8080 | xargs kill -9
lsof -ti:6443 | xargs kill -9
```

### Complete reset

```bash
# Stop all containers and remove volumes
docker-compose down -v

# Remove kubeconfig
rm -rf devenv/k3s-config/

# Reinstall frontend dependencies
cd devenv && task web:install

# Start from step 3
```
