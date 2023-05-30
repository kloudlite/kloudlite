FROM golang:alpine
RUN apk add bash git curl
WORKDIR /workspace
ENV GOBIN=/usr/local/bin
RUN go install github.com/go-task/task/v3/cmd/task@latest
RUN curl -L0 https://github.com/helm/chart-releaser/releases/download/v1.5.0/chart-releaser_1.5.0_linux_amd64.tar.gz > /tmp/chart-releaser.tar.gz && tar xf /tmp/chart-releaser.tar.gz -C /tmp && mv /tmp/cr $GOBIN
COPY Taskfile.yml ./Taskfile.yml
COPY cmd ./cmd
RUN task setup
# USER 1000
