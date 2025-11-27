# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the kl-tun-proxy binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /kl-tun-proxy \
    ./cmd/kl-tun-proxy/main.go

# Runtime stage - minimal image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

# Copy binary from builder
COPY --from=builder /kl-tun-proxy /usr/local/bin/kl-tun-proxy

# Run as non-root user
USER nonroot:nonroot

# Expose WireGuard UDP port
EXPOSE 51820/udp

# Run the client
ENTRYPOINT ["/usr/local/bin/kl-tun-proxy"]
