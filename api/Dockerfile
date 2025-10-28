# syntax=docker/dockerfile:1

FROM golang:1.24 AS builder

# Set working directory
WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Final image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the server binary
COPY --from=builder /workspace/server /app/server

# Expose API port
EXPOSE 8080

# Run as non-root user
USER nonroot:nonroot

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD ["/app/server", "health"]

# Run the server
CMD ["/app/server"]
