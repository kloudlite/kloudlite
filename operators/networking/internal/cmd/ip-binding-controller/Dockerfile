FROM --platform=$BUILDPLATFORM debian:stable-slim AS build
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
COPY ./${BINARY}-${TARGETARCH} /manager
# RUN setcap cap_net_bind_service=+ep /manager
RUN setcap cap_net_raw=+ep /manager
# RUN chmod +x /manager
# RUN setcap cap_net_raw=+ep /manager

# FROM gcr.io/distroless/static:nonroot
# FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian11:latest
FROM --platform=$TARGETPLATFORM gcr.io/distroless/static-debian11:latest
COPY --from=build /manager /manager
ENTRYPOINT [ "/manager" ]
