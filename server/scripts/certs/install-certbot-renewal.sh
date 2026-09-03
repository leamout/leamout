#!/bin/sh
set -eu

TLS_DOMAIN="${TLS_DOMAIN:-}"
REPO_DIR="${REPO_DIR:-$(pwd)}"
CERT_DIR="${CERT_DIR:-$REPO_DIR/deploy/certs}"
ENV_FILE="${ENV_FILE:-$REPO_DIR/.env}"
COMPOSE_FILE="${COMPOSE_FILE:-$REPO_DIR/deploy/compose.yaml}"
CERTBOT_HOOK_DIR="${CERTBOT_HOOK_DIR:-/etc/letsencrypt/renewal-hooks/deploy}"
HOOK_PATH="$CERTBOT_HOOK_DIR/leamout-opensips"

if [ -z "$TLS_DOMAIN" ]; then
  echo "TLS_DOMAIN is required, for example: TLS_DOMAIN=sip.example.com" >&2
  exit 1
fi

if [ "$(id -u)" -ne 0 ]; then
  echo "Automatic renewal setup must run as root." >&2
  echo "Run: sudo TLS_DOMAIN=$TLS_DOMAIN make certs-auto-renew" >&2
  exit 1
fi

if [ ! -f "$REPO_DIR/server/scripts/certs/certbot-deploy-hook.sh" ]; then
  echo "Leamout Certbot deploy hook is missing under: $REPO_DIR" >&2
  exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
  echo "Leamout environment file is missing: $ENV_FILE" >&2
  exit 1
fi

mkdir -p "$CERTBOT_HOOK_DIR"

cat > "$HOOK_PATH" <<EOF
#!/bin/sh
set -eu
cd '$REPO_DIR'
TLS_DOMAIN='$TLS_DOMAIN' \
CERT_DIR='$CERT_DIR' \
ENV_FILE='$ENV_FILE' \
COMPOSE_FILE='$COMPOSE_FILE' \
sh server/scripts/certs/certbot-deploy-hook.sh
EOF
chmod 0755 "$HOOK_PATH"

if command -v systemctl >/dev/null 2>&1; then
  if systemctl list-unit-files certbot.timer >/dev/null 2>&1; then
    systemctl enable --now certbot.timer
    echo "Enabled Certbot systemd renewal timer."
  else
    echo "certbot.timer is not installed; the deploy hook is installed, but renewal scheduling must be provided by the Certbot package or the operator." >&2
  fi
else
  echo "systemctl is not available; the deploy hook is installed, but renewal scheduling must be provided separately." >&2
fi

echo "Installed Leamout Certbot deploy hook: $HOOK_PATH"
echo "Renewed certificate: $TLS_DOMAIN"
echo "Repository: $REPO_DIR"
