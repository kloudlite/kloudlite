# Build the manager binary
FROM golang:1.18-alpine as builder
RUN apk add curl
WORKDIR /workspace
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" > \
    ./kubectl && chmod +x ./kubectl
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download -x
COPY . ./
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go
RUN mkdir /tmp/types

FROM vectorized/redpanda:v22.1.6 as redpanda

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/kubectl /usr/local/bin/kubectl
COPY --from=redpanda /usr/bin/rpk /usr/local/bin/rpk

COPY --from=builder /workspace/manager /manager
COPY --from=builder /tmp/lib /tmp/lib
#RUN mkdir -p /tmp/types
COPY --from=builder --chown=65532:65532 /workspace/lib/templates  /tmp/lib/templates
ENV TEMPLATES_DIR=/tmp/lib/templates
USER 65532:65532

ENTRYPOINT ["/manager"]
