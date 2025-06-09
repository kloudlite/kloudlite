#syntax=docker/dockerfile:1
FROM cgr.dev/chainguard/static:latest-glibc
ARG BINARY TARGETARCH
COPY  --chown=1001 $BINARY-$TARGETARCH ./gateway-kube-proxy
ENTRYPOINT ["./gateway-kube-proxy"]
