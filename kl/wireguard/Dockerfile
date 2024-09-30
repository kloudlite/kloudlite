FROM --platform=$TARGETPLATFORM golang:1.21 as builder
ARG VERSION
ENV CGO_ENABLED=0
WORKDIR /wireguard
COPY . .
RUN go build -o ./wireguard main.go
RUN chmod +x ./wireguard


FROM --platform=$TARGETPLATFORM ghcr.io/kloudlite/hub/wireguard
COPY --from=builder ./wireguard/wireguard /wireguard
COPY healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh
CMD ["/wireguard"]

