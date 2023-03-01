FROM golang:1.18.3-alpine3.16 AS base
USER 1001
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
COPY .  .
RUN go mod download -x
ARG APP
RUN mkdir -p ./apps/$APP
WORKDIR /tmp/app/apps/$APP
COPY --chown=1001 ./  ./
RUN CGO_ENABLED=0 go build -tags musl -o /tmp/bin/$APP ./main.go
RUN chmod +x /tmp/bin/$APP

#FROM gcr.io/distroless/static-debian11
FROM alpine
ARG APP
RUN adduser -D -h /home/nonroot nonroot
USER nonroot
WORKDIR /tmp/app
COPY --chown=nonroot --from=base /tmp/bin/$APP ./$APP  
CMD ["./admission-webhook"]
