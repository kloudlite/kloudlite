# vim: set ft=dockerfile:
FROM --platform=$TARGETPLATFORM docker.io/alpine/helm:3.12.3
WORKDIR /workspace
RUN apk add bash curl
# # FIXME: this url is for amd64 only
ARG KUBECTL_VERSION=v1.30.0 TARGETPLATFORM
RUN curl -L0 https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/${TARGETPLATFORM}/kubectl > /usr/bin/kubectl
RUN chmod +x /usr/bin/kubectl
