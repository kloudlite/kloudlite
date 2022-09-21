# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine as builder
WORKDIR /workspace
RUN apk add curl
ARG name
COPY --from=project go.mod go.sum ./
COPY --from=project lib ./lib
COPY --from=project apis ./apis
COPY --from=project env ./env
RUN mkdir -p ./operators/$name
COPY ./ ./operators/$name
RUN --mount=type=cache,target=/root/.cache/operator.gomod go mod download -json -x
RUN mkdir -p ./operators/$name
COPY ./ ./operators/$name
RUN --mount=type=cache,target=/root/.cache/operator.go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -work -v -a -o manager ./operators/$name/main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/manager /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
