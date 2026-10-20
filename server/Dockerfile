# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.9

FROM oven/bun:1.3.14-alpine AS backoffice-assets

WORKDIR /assets
RUN apk add --no-cache nodejs
COPY internal/backoffice/package.json internal/backoffice/bun.lock ./
RUN bun install --frozen-lockfile
COPY internal/backoffice ./
RUN bun run build

FROM golang:${GO_VERSION}-alpine AS build


WORKDIR /src

RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=backoffice-assets /assets/assets/static ./internal/backoffice/assets/static
RUN go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate

RUN set -eux; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/server \
        ./cmd/server; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/worker \
        ./cmd/worker; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/media \
        ./cmd/media; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/backoffice \
        ./cmd/backoffice

FROM alpine:3.22 AS runtime

WORKDIR /leamout

RUN apk add --no-cache \
        ca-certificates \
        tzdata \
    && addgroup -S leamout \
    && adduser -S -D -H -G leamout leamout

COPY --from=build /out/server /leamout/server
COPY --from=build /out/worker /leamout/worker
COPY --from=build /out/media /leamout/media
COPY --from=build /out/backoffice /leamout/backoffice

USER leamout

EXPOSE 8080 8082

CMD ["/leamout/server"]
# Atlas and the immutable migration set are built as a separate target so
# Docker Compose can run schema changes before rolling out application workloads.
FROM arigaio/atlas:1.3.0-alpine AS migration-build

COPY migrations /migrations
RUN /atlas migrate hash --dir file:///migrations

FROM arigaio/atlas:1.3.0-alpine AS migrate

COPY --from=migration-build /migrations /migrations
ENTRYPOINT ["/atlas"]
