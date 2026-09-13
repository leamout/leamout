ARG GO_VERSION=1.26.6

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

RUN apk add --no-cache \
    ca-certificates \
    git \
    nodejs \
    npm \
    tzdata

RUN go install github.com/a-h/templ/cmd/templ@v0.3.1020

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN templ generate

RUN npm ci --prefix ./internal/backoffice --no-audit --no-fund \
    && npm run --prefix ./internal/backoffice build

RUN set -eux; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/server \
        ./cmd/cloud; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/worker \
        ./cmd/cloud-worker; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/backoffice \
        ./cmd/backoffice; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/selfhosted-server \
        ./cmd/selfhosted; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/selfhosted-worker \
        ./cmd/selfhosted-worker; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/leamout \
        ./cmd/leamout

FROM alpine:3.22 AS runtime

WORKDIR /leamout

RUN apk add --no-cache \
        ca-certificates \
        tzdata \
    && addgroup -S leamout \
    && adduser -S -D -H -G leamout leamout

USER leamout

FROM runtime AS cloud

COPY --from=build /out/server /leamout/server
COPY --from=build /out/worker /leamout/worker
COPY --from=build /out/backoffice /leamout/backoffice

CMD ["/leamout/server"]

FROM runtime AS self-hosted

COPY --from=build /out/selfhosted-server /leamout/server
COPY --from=build /out/selfhosted-worker /leamout/worker
COPY --from=build /out/leamout /leamout/leamout

CMD ["/leamout/server"]
