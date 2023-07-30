FROM gcr.io/distroless/static:nonroot
ARG BIN
COPY ${BIN} /kubelet-metrics-reexporter
ENTRYPOINT [ "/kubelet-metrics-reexporter" ]

