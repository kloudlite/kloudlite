# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine as builder
WORKDIR /workspace
RUN apk add curl
ARG name
COPY --from=project go.mod go.sum ./
COPY --from=project pkg ./pkg
COPY --from=project apis ./apis
RUN mkdir -p ./$name
WORKDIR ./$name
COPY ./ ./
#RUN --mount=type=cache,target=/root/.cache/operator.gomod go mod download -json -x
RUN go mod download -json -x
#RUN --mount=type=cache,target=/root/.cache/operator.go-build CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -work -v -a -o manager ./operators/$name/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -work -v -a -o /workspace/manager ./main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/manager /manager
USER 65532:65532
ENTRYPOINT ["/manager"]
