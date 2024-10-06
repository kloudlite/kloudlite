FROM golang:1.22 as builder

WORKDIR /app

ENV CGO_ENABLED=0

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY main.go main.go

RUN go build -o main .



FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/main ./main

ENTRYPOINT ["./main"]
