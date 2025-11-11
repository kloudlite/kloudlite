FROM alpine:latest AS builder
RUN apk add --no-cache go git make
WORKDIR /app
RUN <<EOF
  git clone --depth=1 https://codeberg.org/eduVPN/proxyguard
  cd proxyguard && CGO_ENABLED=0 make server
EOF

FROM alpine:latest
RUN apk add --no-cache wireguard-tools-wg-quick iputils iptables procps libcap curl # procps for sysctl
COPY --from=builder /app/proxyguard/proxyguard-server /usr/local/bin/proxyguard
