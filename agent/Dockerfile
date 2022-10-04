# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine as builder
RUN apk add curl
WORKDIR /workspace
COPY --from=project ./go.mod ./go.mod
COPY --from=project ./go.sum ./go.sum
RUN ls -al
RUN go mod download -x

COPY ./main.go ./main.go
COPY --from=project ./lib/errors ./lib/errors
COPY --from=project ./lib/logging ./lib/logging
COPY --from=project ./lib/redpanda ./lib/redpanda
COPY --from=project ./lib/functions ./lib/functions

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o kloudlite-agent main.go
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" > \
    ./kubectl && chmod +x ./kubectl

# FROM gcr.io/distroless/static:nonroot
FROM alpine:latest
COPY --from=builder /workspace/kubectl /usr/local/bin/kubectl
COPY --from=builder /workspace/kloudlite-agent /kloudlite-agent
USER 65532:65532
ENTRYPOINT ["/kloudlite-agent"]
