# syntax=docker/dockerfile:1

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the pre-built server binary from GitHub Actions
COPY server /app/server

# Expose API port
EXPOSE 8080

# Run as non-root user
USER nonroot:nonroot

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD ["/app/server", "health"]

# Run the server
CMD ["/app/server"]
