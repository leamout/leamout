#!/bin/sh
set -eu

CERT_DIR="${CERT_DIR:-deploy/certs}"
TLS_DOMAIN="${TLS_DOMAIN:-}"
TLS_EMAIL="${TLS_EMAIL:-}"
CERTBOT_BIN="${CERTBOT_BIN:-certbot}"
CERTBOT_CONFIG_DIR="${CERTBOT_CONFIG_DIR:-/etc/letsencrypt}"

if [ -z "$TLS_DOMAIN" ]; then
  echo "TLS_DOMAIN is required, for example: TLS_DOMAIN=sip.example.com" >&2
  exit 1
fi

if [ -z "$TLS_EMAIL" ]; then
  echo "TLS_EMAIL is required for Let's Encrypt registration and expiry notices" >&2
  exit 1
fi

if ! command -v "$CERTBOT_BIN" >/dev/null 2>&1; then
  echo "certbot is required for production TLS provisioning" >&2
  echo "Install Certbot on the VPS, then run this command again." >&2
  exit 1
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "Let's Encrypt provisioning must run as root so Certbot can bind port 80 and read its certificate store." >&2
  echo "Run: sudo TLS_DOMAIN=$TLS_DOMAIN TLS_EMAIL=$TLS_EMAIL make certs-production" >&2
  exit 1
fi

live_dir="$CERTBOT_CONFIG_DIR/live/$TLS_DOMAIN"

sync_certificates() {
  mkdir -p "$CERT_DIR"
  install -m 0644 "$live_dir/fullchain.pem" "$CERT_DIR/fullchain.pem"
  install -m 0600 "$live_dir/privkey.pem" "$CERT_DIR/privkey.pem"
}

if [ ! -f "$live_dir/fullchain.pem" ] || [ ! -f "$live_dir/privkey.pem" ]; then
  echo "Requesting a Let's Encrypt certificate for $TLS_DOMAIN"
  echo "DNS for $TLS_DOMAIN must point to this VPS and inbound TCP port 80 must be reachable."

  "$CERTBOT_BIN" certonly \
    --standalone \
    --non-interactive \
    --agree-tos \
    --email "$TLS_EMAIL" \
    --domain "$TLS_DOMAIN"
else
  echo "Existing Certbot certificate found for $TLS_DOMAIN; reusing it."
fi

sync_certificates

CERT_DIR="$CERT_DIR" sh scripts/certs/install-system-ca.sh
CERT_DIR="$CERT_DIR" sh scripts/certs/check-certs.sh

cat <<EOF
Production TLS certificate installed in $CERT_DIR.

Certbot manages the source certificate under:
  $live_dir

Renewal should run 'certbot renew' and then copy the renewed certificate into
$CERT_DIR before restarting or reloading OpenSIPS.
EOF
