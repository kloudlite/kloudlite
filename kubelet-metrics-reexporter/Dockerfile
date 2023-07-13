FROM golang:alpine AS base
WORKDIR /workspace
COPY . ./
RUN go mod tidy
RUN go build -ldflags="-s -w" -o ./kubelet-metrics-reexporter

FROM gcr.io/distroless/static:nonroot
COPY --from=base /workspace/kubelet-metrics-reexporter /kubelet-metrics-reexporter
ENTRYPOINT [ "/kubelet-metrics-reexporter" ]

