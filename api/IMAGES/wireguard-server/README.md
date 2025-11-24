# WireGuard TLS Proxy Server (UDP-over-WebSocket)

A custom UDP-over-WebSocket proxy server for tunneling WireGuard traffic over HTTPS/TLS.

## Architecture

```
Client (WireGuard) → UDP-over-WebSocket Client → TLS/WSS → UDP-over-WebSocket Server → WireGuard (51820/UDP)
```

### Components

1. **UDP-over-WebSocket Server**
   - Listens on port 443 (HTTPS/WSS)
   - Terminates TLS 1.3 connections
   - Accepts WebSocket connections
   - Converts WebSocket messages to UDP packets
   - Forwards UDP packets to WireGuard

2. **Protocol**
   - Client initiates WebSocket connection over TLS
   - Client sends: `CONNECT_UDP <target_addr>\n`
   - Server responds: `OK\n` or `ERR\n`
   - Binary frames: `[2-byte length][UDP packet data]`
   - Bidirectional packet forwarding

## Usage

### Command Line Flags

```bash
wireguard-tls-proxy \
  --listen=:443 \
  --tls-cert=/certs/tls.crt \
  --tls-key=/certs/tls.key \
  --wireguard-target=127.0.0.1:51820
```

**Flags:**
- `--listen`: WebSocket server listen address (default: `:443`)
- `--tls-cert`: Path to TLS certificate (default: `/certs/tls.crt`)
- `--tls-key`: Path to TLS private key (default: `/certs/tls.key`)
- `--wireguard-target`: WireGuard UDP endpoint (default: `127.0.0.1:51820`)

### Kubernetes Deployment

The server deployment should mount TLS certificates at `/certs/`:
- `/certs/tls.crt` - TLS certificate
- `/certs/tls.key` - TLS private key

Domain: `vpn-connect.{subdomain}.khost.dev`

## Security

- Enforces TLS 1.3 minimum
- Uses certificates from Kubernetes cert-manager
- All traffic between client and server is encrypted
- WebSocket protocol for firewall-friendly tunneling

## Building

```bash
# Local build
go build -o wireguard-tls-proxy ./IMAGES/wireguard-server

# Docker build
docker build -t wireguard-tls-proxy -f IMAGES/wireguard-server/Dockerfile .
```

## Protocol Details

### Connection Handshake

1. Client connects via TLS WebSocket (wss://)
2. Client sends ASCII header: `CONNECT_UDP 127.0.0.1:51820\n`
3. Server connects to target UDP endpoint
4. Server responds: `OK\n` (success) or `ERR\n` (failure)

### Packet Format

All data packets use binary WebSocket frames:

```
[2 bytes: packet length (big-endian)] [N bytes: UDP packet data]
```

- Length field: 16-bit unsigned integer, big-endian
- Maximum packet size: 65535 bytes
- Both directions use same format

### Session Management

- One WebSocket connection per UDP client
- Automatic cleanup on connection close
- Ping/pong for keepalive (30s interval)

## Differences from ProxyGuard

- Custom implementation (no external dependencies)
- Simpler protocol (length-prefixed packets)
- Direct WebSocket binary frames (no HTTP upgrade complexity)
- Better error handling and logging
- Native Go implementation for better performance

## Troubleshooting

### Check if server is running
```bash
ss -tlnp | grep :443
```

### Test TLS connection
```bash
openssl s_client -connect vpn-connect.example.khost.dev:443 -servername vpn-connect.example.khost.dev
```

### View logs
```bash
kubectl logs -n kloudlite deployment/tunnel-server -f
```
