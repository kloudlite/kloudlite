# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /wireguard-tls-proxy \
    ./IMAGES/wireguard-server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wireguard-tools

WORKDIR /

# Copy the binary from builder
COPY --from=builder /wireguard-tls-proxy /wireguard-tls-proxy

# Expose HTTPS port
EXPOSE 443

# Run the server
ENTRYPOINT ["/wireguard-tls-proxy"]
