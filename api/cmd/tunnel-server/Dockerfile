FROM alpine:latest

RUN apk --no-cache add \
    ca-certificates \
    wireguard-tools-wg-quick \
    iptables \
    ip6tables \
    iputils \
    procps \
    libcap \
    curl \
    bash

WORKDIR /

COPY ./bin/tunnel-server /usr/local/bin/tunnel-server

# Make it executable
RUN chmod +x /usr/local/bin/tunnel-server

# Expose HTTPS port (WireGuard traffic tunneled inside)
EXPOSE 443

# Run the server
ENTRYPOINT ["/usr/local/bin/tunnel-server"]

