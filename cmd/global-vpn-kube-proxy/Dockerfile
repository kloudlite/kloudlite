#syntax=docker/dockerfile:1
FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/static:latest-glibc
ARG BINARY
COPY  --chown=1001 $BINARY ./global-vpn-kube-proxy
ENTRYPOINT ["./global-vpn-kube-proxy"]
