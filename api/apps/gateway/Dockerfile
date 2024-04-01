FROM node:current-alpine
RUN npm i -g pnpm
WORKDIR /home/node
RUN --mount=type=bind,source=package.json,target=package.json \
  --mount=type=bind,source=pnpm-lock.yaml,target=pnpm-lock.yaml \
  --mount=type=cache,target=/root/.local/share/pnpm/store \
  pnpm i
USER node
COPY --chown=node . ./
COPY  . ./
ARG SUPERGRAPH_FILE
COPY ${SUPERGRAPH_FILE} ./
ENV SUPERGRAPH_SCHEMA_FILE=${SUPERGRAPH_FILE}
ENTRYPOINT ["npm","start"]
