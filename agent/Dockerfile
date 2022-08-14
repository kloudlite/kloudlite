# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine as builder
WORKDIR /workspace
COPY --from=project ./go.mod ./go.mod
COPY --from=project ./go.sum ./go.sum
RUN ls -al
RUN go mod download -x

RUN mkdir ./lib
COPY ./main.go ./main.go
COPY --from=project ./lib/errors ./lib/errors
COPY --from=project ./lib/logging ./lib/logging
COPY --from=project ./lib/redpanda ./lib/redpanda

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o kloudlite-agent main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/kloudlite-agent /kloudlite-agent
USER 65532:65532
ENTRYPOINT ["/kloudlite-agent"]
