#!/bin/bash
set -euo pipefail

# Build script for workspace-comprehensive image
# This script builds the kl binary outside Docker and then builds the image

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
API_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "==> Building kl binary for Linux..."
cd "$API_ROOT"

# Build the kl binary for Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o bin/kl-linux \
  cmd/kl/main.go

echo "==> kl binary built successfully at: $API_ROOT/bin/kl-linux"
ls -lh "$API_ROOT/bin/kl-linux"

echo ""
echo "==> Building Docker image..."
cd "$SCRIPT_DIR"

# Build the Docker image
docker build \
  -f Dockerfile \
  -t kloudlite/workspace-comprehensive:latest \
  --build-arg BASE_IMAGE=kloudlite/workspace-base:latest \
  ../..

echo ""
echo "==> Build complete!"
echo "    Image: kloudlite/workspace-comprehensive:latest"
