FROM --platform=$TARGETPLATFORM gcr.io/distroless/static:nonroot
ARG BIN TARGETARCH
COPY ./bin/${BIN}-${TARGETARCH} /kubelet-metrics-reexporter
ENTRYPOINT [ "/kubelet-metrics-reexporter" ]
