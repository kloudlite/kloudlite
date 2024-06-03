FROM linuxserver/wireguard:latest
RUN mkdir -p /config/wg_confs
RUN apk add --no-cache nginx-mod-stream
ENV NGINX_STREAMS_DIR=/etc/nginx/stream.d
ARG BINARY
COPY ${BINARY} /usr/local/bin/ip-manager
# ENTRYPOINT ["ip-manager"]
RUN cat >> /entrypoint.sh <<EOF
#!/usr/bin/env sh
nginx &
pid1=$!
ip-manager $@ &
pid2=$!

trap "eval kill $pid1 $pid2" SIGINT SIGTERM
wait $pid1 $pid2
EOF
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
