FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* main.go ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o wireguard-tls-proxy -ldflags="-s -w" .

FROM alpine:latest
RUN apk add --no-cache wireguard-tools-wg-quick iputils iptables procps libcap curl

# Copy binary
COPY --from=builder /app/wireguard-tls-proxy /usr/local/bin/wireguard-tls-proxy

# Set TLS proxy as the entrypoint
ENTRYPOINT ["/usr/local/bin/wireguard-tls-proxy"]
