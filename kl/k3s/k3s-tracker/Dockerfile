FROM golang:1.22 as builder

WORKDIR /app

ENV CGO_ENABLED=0
ARG TARGETARCH
ENV GOARCH=$TARGETARCH

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go

ARG VERSION

RUN go build -ldflags "-X main.version=$VERSION" -o main .

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/main ./main

ENTRYPOINT ["./main"]
