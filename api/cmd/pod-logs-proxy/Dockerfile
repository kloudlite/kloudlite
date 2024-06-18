FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/static:latest-glibc
ARG BINARY TARGETARCH
COPY  --chown=1001 $BINARY-$TARGETARCH ./pod-logs-proxy
ENTRYPOINT ["./pod-logs-proxy"]
