FROM gcr.io/distroless/static:nonroot
ARG BIN
COPY ./bin/${BIN} /kubelet-metrics-reexporter
ENTRYPOINT [ "/kubelet-metrics-reexporter" ]
