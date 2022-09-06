# context-dir: <project-root>
FROM golang:1.18.3-alpine3.16 AS base
RUN apk add make gcc libc-dev
USER root
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
# COPY --chown=1001 ./go.mod ./go.sum ./tools.go pkg common grpc-interfaces ./
COPY --chown=1001 ./ ./
ARG APP_DIR
WORKDIR $APP_DIR
ENV CGO_ENABLED=0
RUN go build -tags musl -o /tmp/bin/dns-load-balancer ./main.go
RUN chmod +x /tmp/bin/dns-load-balancer

FROM ghcr.io/linuxserver/wireguard:latest
WORKDIR /tmp/bin
COPY --from=base /tmp/bin/dns-load-balancer ./
ENTRYPOINT ["./dns-load-balancer"]
