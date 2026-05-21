FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY ./bin/oci-installer /app/oci-installer
RUN chmod +x /app/oci-installer

USER 65534:65534

ENTRYPOINT ["/app/oci-installer"]
