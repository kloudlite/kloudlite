FROM golang AS base
WORKDIR /app
ARG APP
COPY go.mod ./go.mod
RUN mkdir -p apps/$APP
COPY apps/$APP ./apps/$APP
COPY pkg ./pkg
ENV APP=$APP
ARG CMD_ARGS
ENV CMD_ARGS=$CMD_ARGS
RUN go mod tidy
RUN go build -o bin/$APP apps/$APP/main.go
# RUN make build.$APP -e APP=$APP -e CMD_ARGS=$CMD_ARGS

FROM golang
WORKDIR /app
# RUN go get -u gopkg.in/confluentinc/confluent-kafka-go.v1/kafka
ARG APP
ENV APP=$APP
RUN mkdir bin
COPY --from=base /app/bin/$APP ./bin/$APP
COPY ./Makefile ./Makefile
ENTRYPOINT make start.$APP -e APP=$APP -e CMD_ARGS=$CMD_ARGS
