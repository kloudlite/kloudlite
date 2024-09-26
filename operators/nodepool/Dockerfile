FROM --platform=$TARGETPLATFORM gcr.io/distroless/static:nonroot
ARG BINARY TARGETARCH
COPY ${BINARY}-$TARGETARCH /nodepool-operator
ENTRYPOINT ["/nodepool-operator"]
