#!/bin/sh
set -eu

CERT_DIR="${CERT_DIR:-deploy/certs}"
TLS_DOMAIN="${TLS_DOMAIN:-}"
ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"
RESTART_OPENSIPS="${RESTART_OPENSIPS:-1}"
lineage="${RENEWED_LINEAGE:-}"
renewed_domains="${RENEWED_DOMAINS:-}"

if [ -z "$lineage" ]; then
  echo "RENEWED_LINEAGE is required by the Certbot deploy hook" >&2
  exit 1
fi

if [ -n "$TLS_DOMAIN" ]; then
  matched=0
  for domain in $renewed_domains; do
    if [ "$domain" = "$TLS_DOMAIN" ]; then
      matched=1
      break
    fi
  done

  if [ "$matched" -ne 1 ]; then
    echo "Ignoring renewed certificate for: ${renewed_domains:-unknown domains}"
    exit 0
  fi
fi

if [ ! -f "$lineage/fullchain.pem" ] || [ ! -f "$lineage/privkey.pem" ]; then
  echo "Certbot lineage is missing fullchain.pem or privkey.pem: $lineage" >&2
  exit 1
fi

mkdir -p "$CERT_DIR"
install -m 0644 "$lineage/fullchain.pem" "$CERT_DIR/fullchain.pem"
install -m 0600 "$lineage/privkey.pem" "$CERT_DIR/privkey.pem"

CERT_DIR="$CERT_DIR" sh scripts/certs/check-certs.sh

case "$RESTART_OPENSIPS" in
  0)
    echo "Let's Encrypt certificate synchronized; OpenSIPS restart skipped."
    ;;
  1)
    if [ ! -f "$ENV_FILE" ]; then
      echo "Cannot restart OpenSIPS because environment file is missing: $ENV_FILE" >&2
      exit 1
    fi
    if ! command -v docker >/dev/null 2>&1; then
      echo "docker is required to restart OpenSIPS after certificate renewal" >&2
      exit 1
    fi

    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" restart opensips
    echo "Let's Encrypt certificate synchronized and OpenSIPS restarted."
    ;;
  *)
    echo "RESTART_OPENSIPS must be 0 or 1" >&2
    exit 1
    ;;
esac
