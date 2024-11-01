FROM node:22.10.0-alpine AS remix
WORKDIR  /app
COPY ./package-production.json ./package.json
RUN npm i --frozen-lockfile

FROM node:22.10.0-alpine AS install
RUN npm i -g pnpm@9.12.2
RUN apk add gcc python3 make
WORKDIR  /app
COPY ./package.json ./package.json
COPY ./pnpm-lock.yaml ./pnpm-lock.yaml

# typecheck
ARG APP
ENV APP=${APP}
# COPY ./src/generated/package.json ./src/generated/package.json
# COPY ./src/generated/plugin/package.json ./src/generated/plugin/package.json
COPY ./src/generated/package.json ./src/generated/pnpm-lock.yaml ./src/generated/
COPY ./src/generated/plugin/package.json  ./src/generated/plugin/pnpm-lock.yaml ./src/generated/plugin/

RUN pnpm i -p --frozen-lockfile

FROM node:22.10.0-alpine AS build
RUN npm i -g pnpm@9.12.2
WORKDIR  /app
ARG APP
ENV APP=${APP}
COPY --from=install /app/node_modules ./node_modules
COPY ./src/generated ./src/generated
COPY --from=install /app/src/generated/node_modules ./src/generated/node_modules
COPY --from=install /app/src/generated/plugin/node_modules ./src/generated/plugin/node_modules
COPY ./static/common/. ./public
COPY ./static/${APP}/. ./public

# lib
COPY ./lib ./lib

# typecheck
COPY ./gql-queries-generator/loader.ts ./gql-queries-generator/loader.ts
COPY ./fake-data-generator/gen.ts ./fake-data-generator/gen.ts
COPY ./gql-queries-generator/${APP}.ts ./gql-queries-generator/index.ts
COPY ./tsconfig-compile.json ./tsconfig-compile.json


# app
COPY ./src/apps/${APP} ./src/apps/${APP}
COPY ./tailwind.config.js ./tailwind.config.js
COPY ./tailwind-base.js ./tailwind-base.js
COPY ./remix.config.js ./remix.config.js
COPY ./pnpm-lock.yaml ./pnpm-lock.yaml
COPY ./package.json ./package.json
COPY ./jsconfig.json ./jsconfig.json
COPY ./tsconfig.json ./tsconfig.json
COPY ./remix.env.d.ts ./remix.env.d.ts
COPY ./css-plugins ./css-plugins
COPY ./index.css ./index.css

RUN pnpm build:ts

FROM node:22.10.0-alpine
WORKDIR  /app
ARG APP
ENV APP=${APP}
COPY ./package-production.json ./package.json
COPY ./static/common/. ./public
COPY ./static/${APP}/. ./public
COPY --from=build /app/public ./public
COPY --from=remix /app/node_modules ./node_modules

ENTRYPOINT ["npm", "run", "serve"]
