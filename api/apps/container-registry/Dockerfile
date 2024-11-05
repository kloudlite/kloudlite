# FROM gcr.io/distroless/static-debian12:nonroot
FROM cgr.dev/chainguard/static
ARG BINARY TARGETARCH
COPY ${BINARY}-${TARGETARCH} ./container-registry
ENTRYPOINT ["./container-registry"]
