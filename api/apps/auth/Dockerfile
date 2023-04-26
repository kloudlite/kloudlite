# syntax=docker/dockerfile:1.4
FROM golang:1.18.3-alpine3.16 AS base
# RUN apk add make gcc libc-dev
USER 1001
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
ENV GOPRIVATE=github.com/kloudlite/operator,github.com/kloudlite/wg-operator,github.com/kloudlite/cluster-operator

WORKDIR /tmp/app
COPY --chown=1001 --from=project-root ./go.mod ./go.sum ./tools.go ./
COPY --chown=1001 --from=project-root common ./common
COPY --chown=1001 --from=project-root grpc-interfaces ./grpc-interfaces
COPY --chown=1001 --from=project-root pkg ./pkg
RUN go mod download -x
ARG APP
ARG GH_ACCESS_TOKEN
ENV GH_ACCESS_TOKEN=$GH_ACCESS_TOKEN
RUN mkdir -p ./apps/$APP
WORKDIR /tmp/app/apps/$APP
COPY ./ ./
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 go build -tags musl -o /tmp/bin/$APP ./main.go
RUN chmod +x /tmp/bin/$APP

#FROM gcr.io/distroless/static:nonroot
FROM golang:1.18.3-alpine3.16
USER 1001:1001
ARG APP
COPY --from=base /tmp/bin/$APP /auth
CMD ["/auth"]
