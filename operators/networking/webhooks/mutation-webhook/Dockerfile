FROM gcr.io/distroless/static:nonroot
ARG BINARY
COPY ./${BINARY} /webhook-server
ENTRYPOINT ["/webhook-server"]
