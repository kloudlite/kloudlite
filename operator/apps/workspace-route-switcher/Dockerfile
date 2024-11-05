# syntax=docker/dockerfile:1.4
FROM golang:1.20-alpine as builder
WORKDIR /workspace
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go mod init example.com
RUN go mod tidy
COPY . .
RUN go build -v -a -ldflags='-s -w' -o workspace-route-switcher

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/workspace-route-switcher ./workspace-route-switcher
ENTRYPOINT ["./workspace-route-switcher"]
