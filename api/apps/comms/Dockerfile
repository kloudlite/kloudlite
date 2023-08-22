# syntax=docker/dockerfile:1.4
FROM golang:1.18.3-alpine3.16 AS base
RUN apk add make gcc libc-dev
USER 1001
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
COPY --chown=1001 --from=project-root ./go.mod ./go.sum ./tools.go ./
COPY --chown=1001 --from=project-root common ./common
COPY --chown=1001 --from=project-root grpc-interfaces ./grpc-interfaces
COPY --chown=1001 --from=project-root pkg ./pkg
COPY --chown=1001 email-templates ./email-templates
RUN go mod download -x
ARG APP
RUN mkdir -p ./apps/$APP
WORKDIR /tmp/app/apps/$APP
COPY ./ ./
#TODO: remove kafka from messaging, to force enable CGO_ENABLED=0
ENV CGO_ENABLED=1
RUN GOOS=linux GOARCH=amd64 go build -tags musl -o /tmp/bin/$APP ./main.go
RUN chmod +x /tmp/bin/$APP

FROM gcr.io/distroless/static:nonroot
WORKDIR /tmp
USER 1001:1001
ARG APP
COPY --from=base /tmp/bin/$APP ./comms
CMD ["./comms"]
