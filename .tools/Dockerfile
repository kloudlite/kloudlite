FROM golang:1.17-buster
RUN apt update
RUN curl -L0 https://github.com/operator-framework/operator-sdk/releases/download/v1.18.1/operator-sdk_linux_amd64 > operator-sdk && chmod +x operator-sdk && mv operator-sdk /usr/local/bin/
RUN apt install psmisc net-tools lsof
RUN env GOBIN=/usr/local/bin/ go install github.com/go-task/task/v3/cmd/task@latest
RUN apt install -y inotify-tools
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
USER 1000
