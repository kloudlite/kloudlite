FROM node:alpine as build
RUN npm i -g pnpm
WORKDIR  /app
COPY ./package.json ./package.json

# typecheck
ARG APP
ENV APP=${APP}
COPY ./src/generated/package.json ./src/generated/package.json
COPY ./src/generated/plugin/package.json ./src/generated/plugin/package.json

RUN pnpm i -p

FROM node:alpine as tscheck
RUN npm i -g pnpm
WORKDIR  /app
ARG APP
ENV APP=${APP}
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/src/generated/node_modules ./src/generated/node_modules
COPY --from=build /app/src/generated/plugin/node_modules ./src/generated/plugin/node_modules
COPY ./static/common/. ./public
COPY ./static/${APP}/. ./public

# lib
COPY ./lib ./lib

# design system
COPY ./src/design-system/components ./src/design-system/components
COPY ./src/design-system/index.css ./src/design-system/index.css
COPY ./src/design-system/css ./src/design-system/css
COPY ./src/design-system/tailwind-base.js ./src/design-system/tailwind-base.js
COPY ./src/design-system/tailwind.config.js ./src/design-system/tailwind.config.js

# typecheck
COPY ./src/design-system/.eslintrc.yml ./src/design-system/.eslintrc.yml
COPY ./src/design-system/tsconfig.json ./src/design-system/tsconfig.json
COPY ./src/design-system/jsconfig.json ./src/design-system/jsconfig.json
COPY ./src/design-system/package.json ./src/design-system/package.json
COPY ./src/generated ./src/generated
COPY ./gql-queries-generator/loader.ts ./gql-queries-generator/loader.ts
COPY ./gql-queries-generator/${APP}.ts ./gql-queries-generator/index.ts
COPY ./tsconfig-compile.json ./tsconfig-compile.json

RUN ls ./src/design-system

# app
COPY ./src/apps/${APP} ./src/apps/${APP}
COPY ./tailwind.config.js ./tailwind.config.js
COPY ./remix.config.js ./remix.config.js
COPY ./pnpm-lock.yaml ./pnpm-lock.yaml
COPY ./package.json ./package.json
COPY ./jsconfig.json ./jsconfig.json
COPY ./tsconfig.json ./tsconfig.json
COPY ./remix.env.d.ts ./remix.env.d.ts
RUN pnpm build

# FROM node:alpine
# RUN npm i -g pnpm
# WORKDIR  /app
# ARG APP
# ENV APP=${APP}
# COPY --from=build /app/node_modules ./node_modules
# COPY --from=tscheck ./package.json ./package.json
# COPY ./static/common/. ./public
# COPY ./static/${APP}/. ./public
#
# # lib
# COPY ./lib ./lib
#
# # design system
# COPY ./src/design-system/components ./src/design-system/components
# COPY ./src/design-system/index.css ./src/design-system/index.css
# COPY ./src/design-system/css ./src/design-system/css
# COPY ./src/design-system/tailwind-base.js ./src/design-system/tailwind-base.js
# COPY ./src/design-system/tailwind.config.js ./src/design-system/tailwind.config.js
#
# # typecheck
# COPY ./src/generated ./src/generated
# COPY ./gql-queries-generator/loader.ts ./gql-queries-generator/loader.ts
# COPY ./gql-queries-generator/${APP}.ts ./gql-queries-generator/index.ts
#
# # app
# COPY ./src/apps/${APP} ./src/apps/${APP}
# COPY ./tailwind.config.js ./tailwind.config.js
# COPY ./remix.config.js ./remix.config.js
# COPY ./pnpm-lock.yaml ./pnpm-lock.yaml
# COPY ./package.json ./package.json
# COPY ./jsconfig.json ./jsconfig.json
#
# RUN pnpm build

ENTRYPOINT pnpm serve
