# syntax=docker/dockerfile:1.4
FROM golang:1.18-alpine as builder
RUN apk add curl
WORKDIR /workspace
#RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" > \
#    ./kubectl && chmod +x ./kubectl
#
COPY --from=project go.mod go.sum ./
COPY --from=project lib ./lib
COPY --from=project apis ./apis
COPY --from=project ./operator ./operator

#COPY ./go.mod ./go.sum ./
#COPY ./pkg ./pkg
#COPY ./apis ./apis
#COPY ./operator ./operator
RUN go mod download -x

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

ARG name
RUN mkdir -p ./operators/$name
COPY ./ ./operators/$name
RUN go build -o manager ./operators/$name/main.go

#RUN ls -al
#COPY ./operators/$name ./operators/$name
#RUN go build -o manager ./operators/$name/main.go

FROM gcr.io/distroless/static:nonroot
#COPY --from=builder /workspace/kubectl /usr/local/bin/kubectl
COPY --from=builder /workspace/manager /manager
USER 65532:65532

ENTRYPOINT ["/manager"]
