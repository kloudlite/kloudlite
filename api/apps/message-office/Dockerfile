# syntax=docker/dockerfile:1.4
FROM gcr.io/distroless/static-debian11:nonroot
ARG BINARY TARGETARCH
COPY --chown=1001 $BINARY-$TARGETARCH ./message-office
ENTRYPOINT ["./message-office"]
