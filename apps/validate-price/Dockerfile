# context-dir: <project-root>
FROM golang:buster AS base
USER 1001
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
# COPY --chown=1001 ./go.mod ./go.sum ./tools.go pkg common grpc-interfaces ./
COPY --chown=1001 ./ ./
ARG APP_DIR
WORKDIR $APP_DIR
RUN go build -tags musl -o /tmp/bin/validate-price ./main.go
RUN chmod +x /tmp/bin/validate-price

FROM golang:buster
RUN mkdir /tmp/app
RUN chown -R 1001 /tmp/app
USER 1001
WORKDIR /tmp/app
COPY --from=base --chown=1001 /tmp/bin/validate-price ./
ENTRYPOINT ["./validate-price"]
