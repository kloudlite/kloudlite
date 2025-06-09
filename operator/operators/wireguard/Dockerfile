FROM gcr.io/distroless/static:nonroot
# COPY --from=docker.io/bitnami/kubectl /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl
ARG BINARY TARGETARCH
COPY ${BINARY}-${TARGETARCH} /manager
USER 65532:65532
ENTRYPOINT [ "/manager" ]
