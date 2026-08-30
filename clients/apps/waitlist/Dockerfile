ARG NODE_VERSION=24.13.1
ARG BUN_VERSION=1.3.14

FROM oven/bun:${BUN_VERSION}-alpine AS base
WORKDIR /app

FROM base AS prepare
COPY . .
RUN bunx turbo@2.10.12 prune @leamout/waitlist --docker

FROM base AS builder
COPY --from=prepare /app/out/json/ .
RUN bun install --frozen-lockfile
COPY --from=prepare /app/out/full/ .
RUN bunx turbo build --filter=@leamout/waitlist

FROM node:${NODE_VERSION}-alpine AS runner
WORKDIR /app
RUN addgroup --system --gid 1001 leamout
RUN adduser --system --uid 1001 waitlist
USER waitlist
COPY --from=builder --chown=waitlist:leamout /app/apps/waitlist/.next/standalone ./
COPY --from=builder --chown=waitlist:leamout /app/apps/waitlist/.next/static ./apps/waitlist/.next/static
COPY --from=builder --chown=waitlist:leamout /app/apps/waitlist/public ./apps/waitlist/public
CMD ["node", "apps/waitlist/server.js"]
