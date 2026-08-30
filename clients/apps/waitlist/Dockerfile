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
ENV NODE_ENV=production
ENV HOSTNAME=0.0.0.0
ENV PORT=3002
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs
USER nextjs
COPY --from=builder --chown=nextjs:nodejs /app/apps/waitlist/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/apps/waitlist/.next/static ./apps/waitlist/.next/static
COPY --from=builder --chown=nextjs:nodejs /app/apps/waitlist/public ./apps/waitlist/public
EXPOSE 3002
CMD ["node", "apps/waitlist/server.js"]
