FROM node:20.8.1-alpine as base
RUN npm i -g pnpm
WORKDIR /app
COPY package.json ./package.json
RUN pnpm i

# design system
COPY ../../src/design-system/ ./src/design-system/
WORKDIR /app/src/design-system
RUN pnpm i -p
RUN pnpm local:release

# webinar
WORKDIR /app
COPY . .
RUN pnpm i -p
RUN pnpm build

ENTRYPOINT npm run start
