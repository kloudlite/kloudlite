# Development Setup

## Prerequisites

- Go 1.24+
- Node.js 18+
- pnpm
- Docker & Docker Compose
- kubectl

## Installation

### 1. Clone Repository

```bash
git clone https://github.com/kloudlite/kloudlite-v2.git
cd kloudlite-v2
```

### 2. Start K3s Cluster

```bash
cd devenv
task k3s:up
```

Wait for k3s to be ready (~30 seconds).

### 3. Install Frontend Dependencies

```bash
task web:install
```

### 4. Start Backend API

Open a new terminal:

```bash
cd devenv
task api:dev
```

Backend runs on `http://localhost:8080`

### 5. Start Frontend

Open another terminal:

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
# Check status
task status

# View k3s logs
task k3s:logs

# Restart k3s
task k3s:restart

# Clean up everything
task clean
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

### K3s not ready

```bash
task k3s:down
task k3s:up
```

### Port conflicts

Kill processes on ports 3000, 6443, or 8080:

```bash
lsof -ti:3000 | xargs kill -9
lsof -ti:8080 | xargs kill -9
```

### Reset everything

```bash
task clean
task k3s:up
task web:install
```

Then restart API and frontend servers.
