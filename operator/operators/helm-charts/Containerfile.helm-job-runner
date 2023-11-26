# vim: set ft=dockerfile:
FROM docker.io/alpine/helm:3.12.3
RUN apk add bash curl
RUN curl -L0 -o /usr/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x /usr/bin/kubectl
