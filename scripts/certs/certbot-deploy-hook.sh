#!/bin/sh
set -eu

CERT_DIR="${CERT_DIR:-deploy/certs}"
TLS_DOMAIN="${TLS_DOMAIN:-}"
ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"
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

SOURCE_CERT_DIR="$lineage" \
CERT_DIR="$CERT_DIR" \
ENV_FILE="$ENV_FILE" \
COMPOSE_FILE="$COMPOSE_FILE" \
  sh scripts/certs/activate-runtime.sh
