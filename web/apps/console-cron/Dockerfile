FROM oven/bun:alpine

WORKDIR /app

COPY package.json bun.lock* ./
RUN bun install --frozen-lockfile --production

COPY src ./src
COPY tsconfig.json ./

RUN addgroup --system --gid 1001 cron && \
    adduser --system --uid 1001 cron

USER cron

ENV NODE_ENV=production

CMD ["bun", "run", "src/index.ts"]
