FROM gcr.io/distroless/static:nonroot
ARG BINARY TARGETARCH
COPY ${BINARY}-${TARGETARCH} /webhook-server
ENTRYPOINT ["/webhook-server"]
