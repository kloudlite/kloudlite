# WireGuard TLS Proxy

A TLS termination proxy that integrates ProxyGuard library to enable HTTPS connections to WireGuard over WebSocket.

## Architecture

```
Client (HTTPS/WSS) → TLS Proxy (443) → ProxyGuard (localhost:51821/HTTP) → WireGuard (51820/UDP)
                      ↑________________Same Go Process_________________↑
```

### Components (Single Binary)

1. **TLS Termination Layer**
   - Listens on port 443 (HTTPS)
   - Terminates TLS 1.3 connections
   - Handles WebSocket upgrade requests
   - Proxies decrypted traffic to internal ProxyGuard server

2. **ProxyGuard Library**
   - Runs as goroutine in same process
   - Listens on localhost:51821 (HTTP only, not exposed externally)
   - Converts HTTP WebSocket to UDP packets
   - Forwards UDP packets to WireGuard

3. **WireGuard Interface**
   - Standard WireGuard kernel interface (wg0)
   - Listens on UDP port 51820
   - Handles VPN tunneling

## Usage

### Command Line Flags

```bash
wireguard-tls-proxy \
  --tls-listen=:443 \
  --tls-cert=/certs/tls.crt \
  --tls-key=/certs/tls.key \
  --http-listen=127.0.0.1:51821 \
  --wireguard-target=127.0.0.1:51820
```

**Flags:**
- `--tls-listen`: External TLS endpoint (default: `:443`)
- `--tls-cert`: Path to TLS certificate (default: `/certs/tls.crt`)
- `--tls-key`: Path to TLS private key (default: `/certs/tls.key`)
- `--http-listen`: Internal ProxyGuard HTTP endpoint (default: `127.0.0.1:51821`)
- `--wireguard-target`: WireGuard UDP endpoint (default: `127.0.0.1:51820`)

### Kubernetes Deployment

The tunnel server deployment should mount TLS certificates at `/certs/`:
- `/certs/tls.crt` - TLS certificate
- `/certs/tls.key` - TLS private key

Domain: `vpn-connect.{subdomain}.khost.dev`

### Environment Variables

None required - all configuration via command-line flags.

## Security

- Enforces TLS 1.3 minimum
- Uses certificates from Kubernetes cert-manager
- All traffic between client and proxy is encrypted
- ProxyGuard runs on localhost only (not exposed externally)

## Building

```bash
# Local build
go build -o wireguard-tls-proxy .

# Docker build
docker build -t wireguard-tls-proxy .
```

## Implementation Details

### Single Process Architecture

All components run as goroutines in a single Go process:
- **Main goroutine**: Signal handling and coordination
- **ProxyGuard goroutine**: HTTP WebSocket to UDP conversion (via library)
- **TLS proxy goroutine**: HTTPS/TLS termination and forwarding

### TLS Termination

The proxy creates an HTTPS server with:
- TLS 1.3 minimum version enforcement
- Certificates loaded from files
- Zero timeouts for long-lived WebSocket connections
- HTTP/1.1 protocol for WebSocket compatibility

### WebSocket Proxying Flow

1. Client connects via HTTPS with WebSocket upgrade request
2. TLS layer decrypts and verifies "Upgrade: websocket" header
3. Connects to internal ProxyGuard HTTP server (localhost:51821)
4. Forwards upgrade request over plain HTTP
5. Hijacks client connection for raw byte streaming
6. Bidirectional byte copying between client and ProxyGuard

### ProxyGuard Integration

- Uses `codeberg.org/eduVPN/proxyguard` library
- Calls `proxyguard.Server(ctx, listen, target)` function
- Runs in goroutine alongside TLS proxy
- Converts WebSocket frames to UDP packets
- Forwards UDP to WireGuard interface

### Graceful Shutdown

- Handles SIGTERM/SIGINT signals
- Cancels shared context to stop all goroutines
- TLS server shutdown with 5-second timeout
- ProxyGuard stops when context is cancelled

## Troubleshooting

### Check if TLS proxy is running
```bash
ss -tlnp | grep :443
```

### Check if ProxyGuard is running
```bash
ss -tlnp | grep :51821
```

### Test TLS connection
```bash
openssl s_client -connect vpn-connect.example.khost.dev:443 -servername vpn-connect.example.khost.dev
```

### View logs
```bash
kubectl logs -n kloudlite deployment/tunnel-server -f
```
