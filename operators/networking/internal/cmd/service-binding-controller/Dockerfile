FROM gcr.io/distroless/static:nonroot
ARG BINARY
ARG TARGETARCH
COPY ${BINARY}-${TARGETARCH} /manager
ENTRYPOINT [ "/manager" ]
