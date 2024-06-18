FROM --platform=$TARGETPLATFORM linuxserver/wireguard:latest
RUN mkdir -p /config/wg_confs
RUN apk add --no-cache nginx-mod-stream
ENV NGINX_STREAMS_DIR=/etc/nginx/stream.d
RUN cat > /etc/nginx/nginx.conf <<EOF
user nginx;
wsorker_processes auto;
pcre_jit on;
error_log /var/log/nginx/error.log warn;
# Includes files with directives to load dynamic modules.
include /etc/nginx/modules/*.conf;

# Include files with config snippets into the root context.
include /etc/nginx/conf.d/*.conf;

events {
  worker_connections 1024;
}
EOF
ARG BINARY TARGETARCH
COPY ${BINARY}-${TARGETARCH} /usr/local/bin/ip-manager
# ENTRYPOINT ["ip-manager"]
RUN cat >> /entrypoint.sh <<EOF
#!/usr/bin/env sh
nginx &
pid1=$!
echo "[#] ip-manager \$@"
ip-manager \$@ &
pid2=$!

trap "eval kill $pid1 $pid2" SIGINT SIGTERM
wait $pid1 $pid2
EOF
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/entrypoint.sh"]
