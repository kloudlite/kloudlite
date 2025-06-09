#syntax=docker/dockerfile:1
FROM gcr.io/distroless/static-debian11:nonroot
ARG BINARY TARGETARCH
COPY --chown=1001 $BINARY-$TARGETARCH ./observability
ENTRYPOINT ["./observability"]
