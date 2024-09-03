FROM node:20.8.1-alpine as base
RUN npm i -g pnpm
WORKDIR /webinar
COPY package.json ./package.json
RUN pnpm i
COPY . .
RUN pnpm build

ENTRYPOINT npm run start