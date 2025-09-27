# API Proxy Configuration

This nginx proxy enables K3s containers to access the API server running on the host machine.

## How it Works

1. The proxy listens on port 3001
2. It forwards requests to `host.api:8080` (your API server)
3. The `host.api` DNS name is mapped via docker-compose `extra_hosts`

## Configuration

### Default (Docker Desktop)

No configuration needed. Docker Desktop automatically resolves `host-gateway` to the host machine.

```bash
docker-compose up -d
```

### Custom Host IP

Set the HOST_IP environment variable:

```bash
# Option 1: Export environment variable
export HOST_IP=192.168.1.100
docker-compose up -d

# Option 2: Use .env file
echo "HOST_IP=192.168.1.100" > .env
docker-compose up -d

# Option 3: Inline
HOST_IP=192.168.1.100 docker-compose up -d
```

### Finding Your Host IP

```bash
# On macOS
ifconfig | grep "inet " | grep -v 127.0.0.1

# On Linux
ip addr show | grep "inet " | grep -v 127.0.0.1

# Docker bridge network (Linux)
docker network inspect bridge | grep Gateway
```

## Testing

Test the proxy is working:

```bash
# From host
curl http://localhost:3001/health

# From inside K3s
docker exec kloudlite-k3s curl http://api-proxy:3001/health
```

## Usage in Webhooks

When the API server is running on the host, webhooks can access it via:
- Internal: `http://api-proxy:3001`
- External: `http://localhost:3001`

Note: Kubernetes webhooks require HTTPS, so TLS must be configured for production use.