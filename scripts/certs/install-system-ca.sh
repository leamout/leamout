#!/bin/sh
set -eu

CERT_DIR="${CERT_DIR:-deploy/certs}"
CARRIER_CA_FILE="${CARRIER_CA_FILE:-}"
carrier_ca="$CERT_DIR/carrier-ca.pem"

mkdir -p "$CERT_DIR"

if [ -f "$carrier_ca" ]; then
  echo "Carrier CA bundle already exists: $carrier_ca"
  exit 0
fi

if [ -n "$CARRIER_CA_FILE" ]; then
  if [ ! -f "$CARRIER_CA_FILE" ]; then
    echo "Configured carrier CA file does not exist: $CARRIER_CA_FILE" >&2
    exit 1
  fi
  cp "$CARRIER_CA_FILE" "$carrier_ca"
  chmod 644 "$carrier_ca"
  echo "Installed carrier CA bundle from $CARRIER_CA_FILE"
  exit 0
fi

for candidate in \
  /etc/ssl/certs/ca-certificates.crt \
  /etc/pki/tls/certs/ca-bundle.crt \
  /etc/ssl/ca-bundle.pem \
  /etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem
do
  if [ -f "$candidate" ]; then
    cp "$candidate" "$carrier_ca"
    chmod 644 "$carrier_ca"
    echo "Installed system CA bundle for outbound carrier TLS: $candidate"
    echo "Replace $carrier_ca with a carrier-specific CA bundle when required by the carrier."
    exit 0
  fi
done

echo "Could not find a system CA bundle." >&2
echo "Set CARRIER_CA_FILE=/path/to/carrier-ca.pem and run this command again." >&2
exit 1
