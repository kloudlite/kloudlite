#syntax=docker/dockerfile:1.4
FROM node:alpine3.18
WORKDIR /app
RUN apk add bash curl && rm -rf /var/cache/apk/*
RUN npm i -g zx
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x kubectl && mv kubectl /usr/local/bin/
COPY script.mjs ./script.mjs
RUN chmod +x ./script.mjs
# USER 1717
ENTRYPOINT ["zx", "./script.mjs"]
