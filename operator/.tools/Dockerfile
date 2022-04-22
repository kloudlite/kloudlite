FROM golang:1.17-alpine
RUN apk add curl make gcc libc-dev lsof net-tools util-linux bash zsh
RUN env GOBIN=/usr/local/bin/ go install github.com/go-task/task/v3/cmd/task@latest
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
RUN curl -L0 https://github.com/operator-framework/operator-sdk/releases/download/v1.18.1/operator-sdk_linux_amd64 > operator-sdk && chmod +x operator-sdk && mv operator-sdk /usr/local/bin/
RUN apk add pkgconfig librdkafka
RUN curl -L0 https://get.helm.sh/helm-v3.8.2-linux-amd64.tar.gz > helm.tar.gz && tar -xf helm.tar.gz && mv linux-amd64/helm /usr/local/bin/
RUN adduser \
  --disabled-password \
  -h /home/nxtcoder17 \
  --gecos "" \
  --uid "1000" \
  nxtcoder17
USER nxtcoder17
