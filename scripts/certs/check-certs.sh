#!/bin/sh
set -eu

CERT_DIR="${CERT_DIR:-deploy/certs}"

fullchain="$CERT_DIR/fullchain.pem"
private_key="$CERT_DIR/privkey.pem"
carrier_ca="$CERT_DIR/carrier-ca.pem"

if ! command -v openssl >/dev/null 2>&1; then
  echo "openssl is required to validate TLS certificates" >&2
  exit 1
fi

for file in "$fullchain" "$private_key" "$carrier_ca"; do
  if [ ! -f "$file" ]; then
    echo "Missing required certificate file: $file" >&2
    echo "For local development/CI, run: make certs" >&2
    exit 1
  fi
done

openssl x509 -in "$fullchain" -noout >/dev/null
openssl pkey -in "$private_key" -noout >/dev/null
openssl x509 -in "$carrier_ca" -noout >/dev/null

cert_pub="$(openssl x509 -in "$fullchain" -pubkey -noout | openssl pkey -pubin -outform pem 2>/dev/null)"
key_pub="$(openssl pkey -in "$private_key" -pubout -outform pem 2>/dev/null)"

if [ "$cert_pub" != "$key_pub" ]; then
  echo "TLS certificate and private key do not match" >&2
  exit 1
fi

if ! openssl x509 -checkend 0 -noout -in "$fullchain" >/dev/null; then
  echo "TLS certificate has expired" >&2
  exit 1
fi

echo "TLS certificates are present and valid."
