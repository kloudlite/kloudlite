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
RUN CGO_ENABLED=0 go build -tags musl -o /tmp/bin/$APP ./main.go
RUN chmod +x /tmp/bin/$APP

## context-dir: <project-root>
#FROM golang:1.18-alpine AS base
#RUN apk add make gcc libc-dev
#USER 1001
#ENV GOPATH=/tmp/go
#ENV GOCACHE=/tmp/go-cache
#WORKDIR /tmp/app
## COPY --chown=1001 ./go.mod ./go.sum ./tools.go pkg common grpc-interfaces ./
#COPY --chown=1001 ./ ./
#ARG APP_DIR
#WORKDIR $APP_DIR
#RUN go build -tags musl -o /tmp/bin/console ./main.go
#RUN chmod +x /tmp/bin/console

#FROM golang:1.18-alpine
#RUN mkdir /tmp/app
#RUN chown -R 1001 /tmp/app
#USER 1001
#WORKDIR /tmp/app
#COPY --from=base --chown=1001 /tmp/bin/console ./
#ENTRYPOINT ["./console"]

FROM gcr.io/distroless/static-debian11
USER 1001
WORKDIR /tmp/app
COPY --from=base --chown=1001 /tmp/bin/console ./console
CMD ["./console"]
