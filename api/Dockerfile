FROM golang:alpine
RUN apk add make gcc libc-dev
WORKDIR /app
ARG APP
COPY go.mod ./go.mod
COPY pkg ./pkg
