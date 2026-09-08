ARG GO_VERSION=1.26.6

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

RUN apk add --no-cache \
    ca-certificates \
    git \
    nodejs \
    npm \
    tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN npm install --prefix ./internal/backoffice --no-audit --no-fund \
    && npm run --prefix ./internal/backoffice build

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
        -o /out/backoffice \
        ./cmd/backoffice

FROM alpine:3.22

WORKDIR /leamout

RUN apk add --no-cache \
        ca-certificates \
        tzdata \
    && addgroup -S leamout \
    && adduser -S -D -H -G leamout leamout

COPY --from=build /out/server /leamout/server
COPY --from=build /out/worker /leamout/worker
COPY --from=build /out/backoffice /leamout/backoffice
COPY --from=build /src/internal/backoffice/static /leamout/internal/backoffice/static

USER leamout

CMD ["/leamout/server"]
