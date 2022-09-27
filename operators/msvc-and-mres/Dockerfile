# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine as builder
RUN apk add curl
WORKDIR /workspace
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" > \
    ./kubectl && chmod +x ./kubectl

COPY --from=project go.mod go.sum ./
RUN go mod download -x
COPY --from=project env ./env
COPY --from=project lib ./lib
COPY --from=project apis ./apis

ARG name
RUN mkdir -p ./operators/$name
COPY --from=project ./operators/operator ./operators/operator
COPY ./ ./operators/$name
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o manager ./operators/$name/main.go
RUN mkdir /tmp/types

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/kubectl /usr/local/bin/kubectl
#COPY --from=redpanda /usr/bin/rpk /usr/local/bin/rpk
COPY --from=builder /workspace/manager /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
