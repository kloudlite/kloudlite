FROM ghcr.io/kloudlite/hub/wireguard:latest

COPY healthcheck.sh /usr/local/bin/healthcheck.sh

RUN chmod +x /usr/local/bin/healthcheck.sh

HEALTHCHECK --interval=2s --timeout=2s --retries=1  CMD /usr/local/bin/healthcheck.sh
