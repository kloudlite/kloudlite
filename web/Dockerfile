FROM node:alpine as base
RUN npm i -g pnpm
WORKDIR  /app
COPY package.json pnpm-lock.yaml ./
COPY postcss.config.js ./
COPY remix.config.js ./
COPY tailwind.config.js ./
RUN pnpm i


FROM node:alpine as build
RUN npm i -g pnpm
WORKDIR  /app
COPY --from=base /app/node_modules ./node_modules
COPY --from=base /app/pnpm-lock.yaml ./
COPY --from=base /app/package.json ./
COPY --from=base /app/postcss.config.js ./
COPY --from=base /app/remix.config.js ./
COPY --from=base /app/tailwind.config.js ./
ARG APP
COPY ./src/apps/${APP} ./src/apps/${APP}
COPY ./src/components ./src/components
COPY ./src/index.css ./src/index.css
COPY ./src/tailwind.config.js ./src/tailwind.config.js
ENV APP=${APP}
RUN pnpm run build


FROM node:alpine
RUN npm i -g pnpm
WORKDIR  /app
COPY --from=base /app/pnpm-lock.yaml ./
COPY --from=base /app/package.json ./
COPY --from=base /app/postcss.config.js ./
COPY --from=base /app/remix.config.js ./
COPY --from=base /app/tailwind.config.js ./
RUN pnpm install -P
ARG APP
COPY --from=build /app/src/apps/${APP}/build ./build
ENTRYPOINT [ "npm", "run", "start" ]