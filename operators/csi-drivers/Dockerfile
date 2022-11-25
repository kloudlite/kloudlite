# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine as builder
RUN apk add curl
WORKDIR /workspace
COPY --from=project go.mod go.sum ./
COPY --from=project lib ./lib
COPY --from=project apis ./apis
COPY --from=project ./operator ./operator
RUN go mod download -x

ARG name
RUN mkdir -p ./operators/$name
COPY ./ ./operators/$name
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o manager ./operators/$name/main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/manager /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
