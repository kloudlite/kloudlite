FROM oven/bun:alpine

WORKDIR /app

COPY package.json bun.lock* ./
RUN bun install --frozen-lockfile --production

COPY src ./src
COPY tsconfig.json ./

RUN addgroup --system --gid 1001 cronapp && \
    adduser --system --uid 1001 cronapp

USER cronapp

ENV NODE_ENV=production

CMD ["bun", "run", "src/index.ts"]
