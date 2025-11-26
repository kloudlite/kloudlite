#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Generating self-signed TLS certificates..."

# Generate private key and certificate
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout tls.key \
  -out tls.crt \
  -days 365 \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

chmod 600 tls.key
chmod 644 tls.crt

echo "✓ Certificate generated successfully!"
echo "  - Certificate: $(pwd)/tls.crt"
echo "  - Private key: $(pwd)/tls.key"
echo ""
echo "Note: This is a self-signed certificate for development/testing only."
echo "For production, use certificates from a trusted CA."
