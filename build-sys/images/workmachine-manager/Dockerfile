FROM gcr.io/distroless/static-debian12:nonroot

COPY bin/workmachine-manager /app/workmachine-manager

ENTRYPOINT ["/app/workmachine-manager"]
CMD ["server", "workmachine-manager"]
