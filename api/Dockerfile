FROM golang:alpine
RUN apk add make
WORKDIR /app
ARG APP
COPY go.mod ./go.mod
RUN mkdir -p apps/$APP
COPY apps/$APP ./apps/$APP
COPY pkg ./pkg
RUN go mod tidy
ENV APP=$APP 
ARG CMD_ARGS
ENV CMD_ARGS=$CMD_ARGS
COPY Makefile ./Makefile
RUN make build.$APP -e APP=$APP -e CMD_ARGS=$CMD_ARGS
ENTRYPOINT make start.$APP -e APP=$APP -e CMD_ARGS=$CMD_ARGS
