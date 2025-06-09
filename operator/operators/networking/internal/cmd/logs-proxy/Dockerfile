# FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/static:latest-glibc
FROM --platform=$TARGETPLATFORM busybox:stable-glibc
ARG BINARY TARGETARCH
COPY  --chown=1001 $BINARY-$TARGETARCH ./logs-proxy
ENV EXECUTABLE=/logs-proxy
ENTRYPOINT ["/logs-proxy"]
