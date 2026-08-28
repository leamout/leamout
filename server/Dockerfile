# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.6

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG LEAMOUT_COMMAND=server
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/leamout \
    "./cmd/${LEAMOUT_COMMAND}"

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S leamout \
    && adduser -S -G leamout leamout

COPY --from=builder /out/leamout /usr/local/bin/leamout

USER leamout

ENTRYPOINT ["/usr/local/bin/leamout"]
