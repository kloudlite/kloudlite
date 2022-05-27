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
RUN go build -tags musl -o /tmp/bin/console ./main.go
RUN chmod +x /tmp/bin/console

FROM golang:1.18-alpine
RUN mkdir /tmp/app
RUN chown -R 1001 /tmp/app
USER 1001
WORKDIR /tmp/app
COPY --from=base --chown=1001 /tmp/bin/console ./
ENTRYPOINT ["./console"]
