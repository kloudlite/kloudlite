FROM --platform=$TARGETPLATFORM alpine:3.18
RUN apk add iputils-ping
ARG BINARY TARGETARCH
COPY ./${BINARY}-${TARGETARCH} /manager
ENTRYPOINT [ "/manager" ]
