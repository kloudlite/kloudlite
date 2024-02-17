# syntax=docker/dockerfile:1.4
FROM gcr.io/distroless/static-debian11:nonroot
ARG BINARY
COPY ./bin/$BINARY ./worker-audit-logging
ENTRYPOINT [ "./worker-audit-logging" ]
