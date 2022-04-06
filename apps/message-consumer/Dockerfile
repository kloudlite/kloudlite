FROM shared-lib:api-go AS base
WORKDIR /app
ARG APP_DIR
RUN mkdir -p ${APP_DIR}
WORKDIR ${APP_DIR}
COPY [^Makefile]* ./
COPY internal ./internal
RUN go mod tidy
COPY Makefile ./
RUN make build

FROM golang:alpine
RUN apk add make
# RUN apk add make gcc libc-dev
WORKDIR /app
ARG APP_DIR
COPY --from=base /app/${APP_DIR}/bin ./
COPY ./Makefile ./Makefile
ENTRYPOINT ["make", "start"]
