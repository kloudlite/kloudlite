FROM node:alpine as build
RUN npm i -g pnpm
WORKDIR  /app
COPY ./package.json ./package.json
RUN pnpm i -p

FROM node:alpine
RUN npm i -g pnpm
WORKDIR  /app
ARG APP
ENV APP=${APP}
COPY --from=build /app/node_modules ./node_modules
COPY ./static/common/. ./public
COPY ./static/${APP}/. ./public

# lib
COPY ./lib ./lib

# design system
COPY ./src/design-system/components ./src/design-system/components
COPY ./src/design-system/index.css ./src/design-system/index.css
COPY ./src/design-system/tailwind-base.js ./src/design-system/tailwind-base.js
COPY ./src/design-system/tailwind.config.js ./src/design-system/tailwind.config.js

# app
COPY ./src/apps/${APP} ./src/apps/${APP}
COPY ./tailwind.config.js ./tailwind.config.js
COPY ./remix.config.js ./remix.config.js
COPY ./pnpm-lock.yaml ./pnpm-lock.yaml
COPY ./package.json ./package.json
COPY ./jsconfig.json ./jsconfig.json

RUN pnpm build

ENTRYPOINT pnpm serve
