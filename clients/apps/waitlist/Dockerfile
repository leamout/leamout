ARG NODE_VERSION=24
ARG BUN_VERSION=1.3.14

FROM oven/bun:${BUN_VERSION}-alpine AS bun

FROM node:${NODE_VERSION}-alpine AS base
COPY --from=bun /usr/local/bin/bun /usr/local/bin/bun
WORKDIR /app

FROM base AS prepare
COPY . .
RUN bun x turbo@2.10.12 prune @leamout/waitlist --docker

FROM base AS builder
COPY --from=prepare /app/out/json/ .
RUN bun install --frozen-lockfile
COPY --from=prepare /app/out/full/ .
WORKDIR /app/apps/waitlist
RUN node ../../node_modules/next/dist/bin/next build

FROM node:${NODE_VERSION}-alpine AS runner
WORKDIR /app
RUN addgroup --system --gid 1001 leamout
RUN adduser --system --uid 1001 waitlist
USER waitlist
COPY --from=builder --chown=waitlist:leamout /app/apps/waitlist/.next/standalone ./
COPY --from=builder --chown=waitlist:leamout /app/apps/waitlist/.next/static ./apps/waitlist/.next/static
COPY --from=builder --chown=waitlist:leamout /app/apps/waitlist/public ./apps/waitlist/public
CMD ["node", "apps/waitlist/server.js"]
