# syntax=docker/dockerfile:1

# Dependencies stage
FROM oven/bun:alpine AS deps

WORKDIR /app

# Copy package files (lockfile optional for standalone builds)
COPY package.json ./
COPY bun.lock* ./

# Install dependencies
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile 2>/dev/null || bun install

# Builder stage
FROM oven/bun:alpine AS builder

WORKDIR /app

# Copy dependencies from deps stage
COPY --from=deps /app/node_modules ./node_modules

# Copy source code
COPY . .

# Build with cache mount for Next.js build cache
RUN --mount=type=cache,target=/app/.next/cache \
    --mount=type=secret,id=env_file,target=/app/.env.local \
    bun run build

# Runtime stage
FROM oven/bun:alpine AS runner

WORKDIR /app

# Install OpenSSL for certificate generation
RUN apk add --no-cache --no-scripts openssl

# Create a non-root user
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Copy built artifacts from builder stage (monorepo structure)
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./apps/console/.next/static
COPY --from=builder --chown=nextjs:nodejs /app/public ./apps/console/public

USER nextjs

EXPOSE 3000

ENV PORT=3000 \
    HOSTNAME="0.0.0.0" \
    NODE_ENV=production

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD bun -e "fetch('http://localhost:3000').then(r => process.exit(r.ok ? 0 : 1)).catch(() => process.exit(1))"

CMD ["bun", "apps/console/server.js"]
