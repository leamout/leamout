#!/bin/sh
set -eu

CERT_DIR="${CERT_DIR:-deploy/certs}"
CERT_CN="${CERT_CN:-localhost}"
CERT_DAYS="${CERT_DAYS:-365}"
CERT_FORCE="${CERT_FORCE:-0}"

fullchain="$CERT_DIR/fullchain.pem"
private_key="$CERT_DIR/privkey.pem"
carrier_ca="$CERT_DIR/carrier-ca.pem"

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required to generate TLS certificates" >&2
  exit 1
fi

case "$CERT_DAYS" in
  ''|*[!0-9]*)
    echo "CERT_DAYS must be a positive integer" >&2
    exit 1
    ;;
esac

if [ "$CERT_DAYS" -le 0 ]; then
  echo "CERT_DAYS must be greater than zero" >&2
  exit 1
fi

if [ "$CERT_FORCE" != "1" ]; then
  for file in "$fullchain" "$private_key" "$carrier_ca"; do
    if [ -e "$file" ]; then
      echo "Refusing to overwrite existing certificate file: $file" >&2
      echo "Set CERT_FORCE=1 only if you intentionally want to replace the self-signed certificates." >&2
      exit 1
    fi
  done
fi

mkdir -p "$CERT_DIR"
umask 077

openssl req \
  -x509 \
  -newkey rsa:2048 \
  -sha256 \
  -nodes \
  -keyout "$private_key" \
  -out "$fullchain" \
  -subj "/CN=$CERT_CN" \
  -days "$CERT_DAYS"

cp "$fullchain" "$carrier_ca"
chmod 600 "$private_key"
chmod 644 "$fullchain" "$carrier_ca"

echo "Generated self-signed TLS certificates in $CERT_DIR"
echo "These certificates are intended for local development and CI, not public production TLS."
