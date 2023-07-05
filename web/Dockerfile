FROM node:alpine as build
RUN npm i -g pnpm
WORKDIR  /app
COPY ./scripts ./scripts
COPY ./remix.config.js ./tailwind.config.js ./
COPY package.json pnpm-lock.yaml ./
RUN pnpm i
ARG APP
COPY ./src/apps/${APP} ./src/apps/${APP}
COPY ./src/components ./src/components
COPY lib/app-setup/index.css ./src/index.css
ENV APP=${APP}
RUN pnpm build


FROM node:alpine
WORKDIR  /app
COPY ./scripts ./scripts
COPY ./remix.config.js ./
COPY ./tailwind.config.js ./
ARG APP
ENV APP=${APP}
RUN npm i -g @remix-run/serve@1.17.0 @remix-run/node@1.17.0
COPY --from=build /app/public/${APP} ./public/${APP}
ENTRYPOINT [ "sh", "scripts/run.sh" ]