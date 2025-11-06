# Building Workspace Comprehensive Image

## Overview

The workspace-comprehensive image includes the `kl` CLI tool built outside of Docker for faster builds and better caching.

## Build Process

### Automated Build (Recommended)

Use the provided build script:

```bash
./build.sh
```

This script will:
1. Build the `kl` binary for Linux (amd64)
2. Build the Docker image with the pre-built binary

### Manual Build

If you need more control:

```bash
# 1. Build the kl binary
cd ../../  # Go to api root
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o bin/kl-linux \
  cmd/kl/main.go

# 2. Build the Docker image
cd workspace-images/comprehensive
docker build \
  -f Dockerfile \
  -t kloudlite/workspace-comprehensive:latest \
  --build-arg BASE_IMAGE=kloudlite/workspace-base:latest \
  ../..
```

### Multi-Architecture Build

For ARM64 support:

```bash
# Build for both amd64 and arm64
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f Dockerfile \
  -t kloudlite/workspace-comprehensive:latest \
  --push \
  ../..
```

**Note**: For multi-arch builds, you'll need to build separate binaries for each architecture and modify the Dockerfile to copy the appropriate binary based on `TARGETARCH`.

## CI/CD Integration

In GitHub Actions or other CI systems:

```yaml
- name: Build kl binary
  run: |
    cd api
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags="-s -w" \
      -o bin/kl-linux \
      cmd/kl/main.go

- name: Build Docker image
  run: |
    cd api/workspace-images/comprehensive
    docker build \
      -f Dockerfile \
      -t kloudlite/workspace-comprehensive:latest \
      ../..
```

## Benefits of External Build

1. **Faster Builds**: Go binary is built once outside Docker, reusing Go build cache
2. **Better Caching**: Docker layers don't include Go source changes
3. **Smaller Context**: Docker build context doesn't need Go modules
4. **Easier Debugging**: Binary can be tested independently before image build
5. **CI/CD Friendly**: Binary artifact can be cached across builds

## Troubleshooting

### Binary Not Found Error

If you get an error about `bin/kl-linux` not found:

```bash
# Make sure you're in the api directory when building
cd /path/to/api
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/kl-linux cmd/kl/main.go
```

### Permission Denied

If the binary is not executable in the container:

```bash
# The Dockerfile already includes chmod +x, but you can verify:
docker run --rm kloudlite/workspace-comprehensive:latest kl version
```
