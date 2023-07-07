# syntax=docker/dockerfile:1.4
FROM golang:1.19-alpine as builder
WORKDIR /workspace

COPY --from=project go.mod go.sum ./
COPY --from=project pkg ./pkg
COPY --from=project apis ./apis
RUN go mod download -x
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -v -a -o env-route-switcher

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/env-route-switcher ./workspace-route-switcher
ENTRYPOINT ["./workspace-route-switcher"]
