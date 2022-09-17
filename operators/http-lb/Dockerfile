# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine as builder
WORKDIR /workspace
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./lb ./main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /workspace/lb /lb

ENTRYPOINT ["/lb"]
