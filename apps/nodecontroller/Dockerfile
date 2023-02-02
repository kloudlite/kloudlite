# syntax=docker/dockerfile:1.4
FROM golang:1.18.3-alpine3.16 AS base
USER 1001
ENV GOPATH=/tmp/go
ENV GOCACHE=/tmp/go-cache
WORKDIR /tmp/app
COPY --chown=1001 --from=project-root  ./go.mod ./go.sum ./tools.go ./
RUN go mod download -x
COPY --chown=1001 --from=project-root pkg ./pkg
ARG APP
RUN mkdir -p ./apps/$APP
WORKDIR /tmp/app/apps/$APP
COPY --chown=1001 ./  ./
RUN CGO_ENABLED=0 go build -tags musl -o /tmp/bin/$APP ./main.go
RUN chmod +x /tmp/bin/$APP

#FROM gcr.io/distroless/static-debian11
FROM alpine
RUN adduser -D -h /home/nonroot nonroot
RUN apk add curl git openssh-client
RUN cd bin && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl
RUN cd bin && curl -o tf.zip https://releases.hashicorp.com/terraform/1.2.9/terraform_1.2.9_linux_amd64.zip && unzip tf.zip
# RUN cd bin && curl -L0 -o talosctl https://github.com/siderolabs/talos/releases/download/v1.2.3/talosctl-linux-amd64 && chmod +x talosctl
USER nonroot
WORKDIR /tmp/app
COPY --chown=nonroot --from=base /tmp/bin/nodecontroller ./nodecontroller  
COPY --chown=nonroot --from=project-root  ./pkg/infraClient/terraform  /templates/terraform
# COPY --chown=nonroot --from=project-root  ./pkg/infraClient/templates  /tmp/app/pkg/infraClient/templates
CMD ["./nodecontroller"]
