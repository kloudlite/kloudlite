FROM gcr.io/distroless/static-debian12:nonroot

COPY bin/api-server /app/api-server

ENTRYPOINT ["/app/api-server"]
CMD ["server", "api"]
