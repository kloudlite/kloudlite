FROM golang:1.18-alpine as builder
COPY . ./
ENV CGO_ENABLED=0
RUN go build -o /lb ./main.go

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /lb /lb
ENTRYPOINT ["/lb"]
