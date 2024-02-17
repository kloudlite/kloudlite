FROM node:current-alpine
RUN npm i -g pnpm
WORKDIR /home/node
USER node
COPY --chown=node package.json pnpm-lock.yaml ./
RUN pnpm i
COPY --chown=node . ./
ENTRYPOINT ["npm","start"]
