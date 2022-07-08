# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine AS base
USER 1001
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
COPY --chown=1001 --from=project-root ./go.mod ./go.sum ./tools.go ./
RUN go mod download -x
COPY --chown=1001 --from=project-root common ./common
COPY --chown=1001 --from=project-root grpc-interfaces ./grpc-interfaces
COPY --chown=1001 --from=project-root pkg ./pkg
ARG APP
RUN mkdir -p ./apps/$APP
WORKDIR /tmp/app/apps/$APP
COPY --chown=1001 ./  ./
RUN CGO_ENABLED=0 go build -tags musl -o /tmp/bin/ci ./main.go
RUN chmod +x /tmp/bin/ci

#FROM golang:1.18-alpine
#FROM gcr.io/distroless/static-debian11
FROM alpine
USER 1001
WORKDIR /tmp/app
COPY --from=base --chown=1001 /tmp/bin/ci ./ci
CMD ["./ci"]
