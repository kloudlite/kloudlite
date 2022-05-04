# Build the manager binary
FROM golang:1.17 as builder
RUN apt update && apt install -y librdkafka-dev

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . ./
# COPY main.go main.go
# COPY api/ api/
# COPY controllers/ controllers/
# COPY apis apis/
# COPY lib lib/
# COPY agent agent/

# Build
#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go
RUN GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/static:nonroot
FROM golang:1.17
RUN apt update && apt install -y kubernetes-client
WORKDIR /
COPY --from=builder /workspace/manager .
RUN mkdir -p /tmp/lib
COPY --from=builder --chown=65532:65532 /workspace/lib/templates  /tmp/lib/templates
ENV TEMPLATES_DIR=/tmp/lib/templates
USER 65532:65532

ENTRYPOINT ["/manager"]
