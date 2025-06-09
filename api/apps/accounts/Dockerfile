# syntax=docker/dockerfile:1.4
FROM gcr.io/distroless/static:nonroot
ARG BINARY TARGETARCH
COPY $BINARY-$TARGETARCH ./accounts
ENTRYPOINT [ "./accounts" ]
