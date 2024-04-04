FROM ghcr.io/linuxserver/wireguard
WORKDIR /apps
ARG APP
COPY ./bin/${APP} /apps/app
RUN mkdir -p /config/wg_confs
ENTRYPOINT ["/apps/app"]
