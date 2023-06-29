FROM node:current-alpine3.16
RUN npm i -g pnpm
WORKDIR  /app

COPY package.json pnpm-lock.yaml ./

ENV STORE_DIR=/app/pnpm
RUN pnpm i -P  --virtual-store-dir ${STORE_DIR}

ARG APP
RUN mkdir ${APP}
WORKDIR /app/${APP}
COPY ./src/apps/${APP}/package.json ./src/apps/${APP}/pnpm-lock.yaml ./
RUN pnpm i -P --virtual-store-dir ${STORE_DIR}
COPY ./src/apps/${APP}/public ./public
COPY ./src/apps/${APP}/build ./build

ENTRYPOINT [ "npm", "run", "start" ]