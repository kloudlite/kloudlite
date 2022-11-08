# syntax=docker/dockerfile:1.4
FROM golang:1.18.3-alpine3.16 AS base
USER 1001
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
COPY --link --chown=1001 --from=project-root ./go.mod ./go.sum ./tools.go ./
RUN go mod download -x
COPY --link --chown=1001 --from=project-root common ./common
COPY --link --chown=1001 --from=project-root grpc-interfaces ./grpc-interfaces
COPY --link --chown=1001 --from=project-root pkg ./pkg
ARG APP
RUN mkdir -p ./apps/$APP
WORKDIR /tmp/app/apps/$APP
COPY --link --chown=1001 ./  ./
ENV CGO_ENABLED=0
RUN go build -tags musl -o /tmp/bin/ci ./main.go
# ENV GOBIN=/tmp/bin
# RUN go install .
# RUN ls -al /tmp/bin
RUN chmod +x /tmp/bin/ci

#FROM golang:1.18-alpine
# FROM gcr.io/distroless/static-debian11
FROM alpine
USER 1001
#WORKDIR /tmp/app
COPY --link --from=base --chown=1001 /tmp/bin/ci ./ci
CMD ["./ci"]
