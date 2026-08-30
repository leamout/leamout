ARG NODE_VERSION=24
ARG BUN_VERSION=1.3.14

FROM oven/bun:${BUN_VERSION}-alpine AS base
WORKDIR /app

FROM base AS prepare
COPY . .
RUN bunx turbo@2.10.12 prune @leamout/web --docker

FROM base AS builder
COPY --from=prepare /app/out/json/ .
RUN bun install --frozen-lockfile
COPY --from=prepare /app/out/full/ .
RUN bunx turbo build --filter=@leamout/web

FROM node:${NODE_VERSION}-alpine AS runner
WORKDIR /app
RUN addgroup --system --gid 1001 leamout
RUN adduser --system --uid 1001 web
USER web
COPY --from=builder --chown=web:leamout /app/apps/web/.next/standalone ./
COPY --from=builder --chown=web:leamout /app/apps/web/.next/static ./apps/web/.next/static
COPY --from=builder --chown=web:leamout /app/apps/web/public ./apps/web/public
CMD ["node", "apps/web/server.js"]
