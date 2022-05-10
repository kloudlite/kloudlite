# context-dir: <project-root>
FROM golang:1.18-alpine AS base
RUN apk add make gcc libc-dev
USER 1001
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
# COPY --chown=1001 ./go.mod ./go.sum ./tools.go pkg common grpc-interfaces ./
COPY --chown=1001 ./ ./
ARG APP_DIR
WORKDIR $APP_DIR
RUN go build -tags musl -o /tmp/bin/comms ./main.go
RUN chmod +x /tmp/bin/comms

FROM golang:1.18-alpine
USER 1001
WORKDIR /tmp/app
COPY --from=base --chown=1001 /tmp/bin/comms ./
ENTRYPOINT ["./comms"]
