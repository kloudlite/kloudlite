FROM alpine:3.16
WORKDIR /app
RUN apk add bash curl jq coreutils && rm -rf /var/cache/apk/*
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
RUN chmod +x kubectl && mv kubectl /usr/local/bin/
COPY main.sh ./main.sh
RUN chmod +x main.sh
USER 1717
CMD ["./main.sh"]

