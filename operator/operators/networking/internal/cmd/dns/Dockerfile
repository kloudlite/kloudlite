# FROM --platform=$BUILDPLATFORM debian:stable-slim AS build
# RUN export DEBCONF_NONINTERACTIVE_SEEN=true \
#   DEBIAN_FRONTEND=noninteractive \
#   DEBIAN_PRIORITY=critical \
#   TERM=linux ; \
#   apt-get -qq update ; \
#   apt-get -yyqq upgrade ; \
#   apt-get -yyqq install ca-certificates libcap2-bin; \
#   apt-get clean
# ARG BINARY TARGETARCH
# COPY ${BINARY}-${TARGETARCH} /dns
# RUN setcap cap_net_bind_service=+ep /dns
#
# FROM gcr.io/distroless/static:nonroot
# COPY --from=build /dns /dns
# ENTRYPOINT [ "/dns" ]

ARG DEBIAN_IMAGE=debian:stable-slim
ARG BASE=gcr.io/distroless/static-debian11:nonroot
FROM --platform=$BUILDPLATFORM ${DEBIAN_IMAGE} AS build
SHELL [ "/bin/sh", "-ec" ]

RUN export DEBCONF_NONINTERACTIVE_SEEN=true \
  DEBIAN_FRONTEND=noninteractive \
  DEBIAN_PRIORITY=critical \
  TERM=linux ; \
  apt-get -qq update ; \
  apt-get -yyqq upgrade ; \
  apt-get -yyqq install ca-certificates libcap2-bin; \
  apt-get clean
ARG BINARY TARGETARCH
COPY ${BINARY}-${TARGETARCH} /dns
RUN setcap cap_net_bind_service=+ep /dns

FROM --platform=$TARGETPLATFORM ${BASE}
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /dns /dns
USER nonroot:nonroot
EXPOSE 53 53/udp
ENTRYPOINT ["/dns"]
