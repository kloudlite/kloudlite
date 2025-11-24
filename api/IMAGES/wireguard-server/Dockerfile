# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server binary (includes integrated watcher)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /wireguard-tls-proxy \
    ./IMAGES/wireguard-server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add \
    ca-certificates \
    wireguard-tools-wg-quick \
    iptables \
    ip6tables \
    iputils \
    procps \
    libcap \
    curl \
    bash

WORKDIR /

# Copy binary from builder
COPY --from=builder /wireguard-tls-proxy /usr/local/bin/wireguard-tls-proxy

# Make it executable
RUN chmod +x /usr/local/bin/wireguard-tls-proxy

# Expose HTTPS port and WireGuard port
EXPOSE 443
EXPOSE 51820/udp

# Run the server
ENTRYPOINT ["/usr/local/bin/wireguard-tls-proxy"]
