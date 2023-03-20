FROM golang:1.18-alpine as builder
WORKDIR /workspace
COPY --from=project go.mod go.sum ./
COPY --from=project pkg ./pkg
COPY --from=project apis ./apis
COPY --from=project ./operator ./operator
RUN go mod download -x

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

ARG name
RUN mkdir -p ./operators/$name
COPY ./ ./operators/$name
RUN go build -o manager ./operators/$name/main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/manager /manager
COPY ./helm-charts /tmp/helm-charts
ENV HELM_CHARTS_DIR=/tmp/helm-charts
USER 65532:65532
ENTRYPOINT ["/manager"]
