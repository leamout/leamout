# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.6

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/server \
        ./cmd/server; \
    CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/worker \
        ./cmd/worker

FROM alpine:3.22

ARG LEAMOUT_COMMAND=server

RUN apk add --no-cache ca-certificates \
    && addgroup -S leamout \
    && adduser -S -G leamout leamout

COPY --from=builder /out/${LEAMOUT_COMMAND} /usr/local/bin/leamout

USER leamout

ENTRYPOINT ["/usr/local/bin/leamout"]
