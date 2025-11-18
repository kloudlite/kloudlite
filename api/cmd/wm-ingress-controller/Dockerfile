# syntax=docker/dockerfile:1

# NOTE: Build the binary outside the container for faster iteration
# Build command: CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/api-server ./cmd/server
# Then copy the pre-built binary into the image

# Runtime image
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the pre-built server binary (built outside Docker)
COPY ./bin/wm-ingress /app/wm-ingress

# Expose API port
EXPOSE 8080

# Run as non-root user
USER nonroot:nonroot

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD ["/app/wm-ingress", "health"]

# Run the wm-ingress
CMD ["/app/wm-ingress"]
ENTRYPOINT ["/app/wm-ingress"]
