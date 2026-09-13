ARG GO_VERSION=1.26.6

FROM golang:${GO_VERSION}-alpine AS source

WORKDIR /src

RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM source AS cloud-build

RUN apk add --no-cache nodejs npm \
    && go install github.com/a-h/templ/cmd/templ@v0.3.1020

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
        ./cmd/backoffice

FROM source AS self-hosted-build

RUN set -eux; \
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

COPY --from=cloud-build /out/server /leamout/server
COPY --from=cloud-build /out/worker /leamout/worker
COPY --from=cloud-build /out/backoffice /leamout/backoffice

CMD ["/leamout/server"]

FROM runtime AS self-hosted

COPY --from=self-hosted-build /out/selfhosted-server /leamout/server
COPY --from=self-hosted-build /out/selfhosted-worker /leamout/worker
COPY --from=self-hosted-build /out/leamout /leamout/leamout

CMD ["/leamout/server"]
