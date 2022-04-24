# context-dir: <project-root>
FROM golang:1.18-alpine AS base
RUN apk add make gcc libc-dev
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
# COPY --chown=1001 ./go.mod ./go.sum ./tools.go pkg common grpc-interfaces ./
# COPY --chown=hotspot ./ ./
COPY ./ ./
ARG APP_DIR
WORKDIR $APP_DIR
RUN go build -tags musl -o /tmp/bin/infra ./main.go
RUN chmod +x /tmp/bin/infra

FROM alpine

RUN apk update && apk add curl openssh terraform bash
RUN curl -sLS https://get.k3sup.dev | sh

RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.18.0/bin/linux/amd64/kubectl
RUN chmod +x ./kubectl
RUN mv ./kubectl /usr/local/bin/kubectl
# kubectl version --client

RUN mkdir -p /home/hotspot

RUN adduser \
    --disabled-password \
    --gecos "" \
    --home /home/hotspot \
    --ingroup wheel \
    --uid 1000 \
    hotspot

RUN chown -R hotspot:wheel /home/hotspot

USER hotspot
WORKDIR /tmp/app
COPY --from=base /tmp/bin/infra ./
COPY ./apps/infra/infra-scripts ./infra-scripts
ENTRYPOINT ["./infra"]
