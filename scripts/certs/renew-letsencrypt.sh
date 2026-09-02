#!/bin/sh
set -eu

CERT_DIR="${CERT_DIR:-deploy/certs}"
TLS_DOMAIN="${TLS_DOMAIN:-}"
CERTBOT_BIN="${CERTBOT_BIN:-certbot}"
CERTBOT_CONFIG_DIR="${CERTBOT_CONFIG_DIR:-/etc/letsencrypt}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"
ENV_FILE="${ENV_FILE:-.env}"

if [ -z "$TLS_DOMAIN" ]; then
  echo "TLS_DOMAIN is required, for example: TLS_DOMAIN=sip.example.com" >&2
  exit 1
fi

if ! command -v "$CERTBOT_BIN" >/dev/null 2>&1; then
  echo "certbot is required for production TLS renewal" >&2
  exit 1
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "Let's Encrypt renewal must run as root." >&2
  echo "Run: sudo TLS_DOMAIN=$TLS_DOMAIN make certs-renew" >&2
  exit 1
fi

live_dir="$CERTBOT_CONFIG_DIR/live/$TLS_DOMAIN"

"$CERTBOT_BIN" renew --cert-name "$TLS_DOMAIN" --non-interactive

if [ ! -f "$live_dir/fullchain.pem" ] || [ ! -f "$live_dir/privkey.pem" ]; then
  echo "Certbot certificate files are missing under $live_dir" >&2
  exit 1
fi

RENEWED_LINEAGE="$live_dir" \
RENEWED_DOMAINS="$TLS_DOMAIN" \
TLS_DOMAIN="$TLS_DOMAIN" \
CERT_DIR="$CERT_DIR" \
ENV_FILE="$ENV_FILE" \
COMPOSE_FILE="$COMPOSE_FILE" \
sh scripts/certs/certbot-deploy-hook.sh
