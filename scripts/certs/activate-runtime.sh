#!/bin/sh
set -eu

SOURCE_CERT_DIR="${SOURCE_CERT_DIR:?SOURCE_CERT_DIR is required}"
CERT_DIR="${CERT_DIR:-deploy/certs}"
ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"

command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 2; }
command -v mktemp >/dev/null 2>&1 || { echo "mktemp is required" >&2; exit 2; }

for name in fullchain.pem privkey.pem; do
  if [ ! -s "$SOURCE_CERT_DIR/$name" ]; then
    echo "source certificate file is missing or empty: $SOURCE_CERT_DIR/$name" >&2
    exit 2
  fi
done
if [ ! -s "$CERT_DIR/carrier-ca.pem" ] && [ ! -s "$SOURCE_CERT_DIR/carrier-ca.pem" ]; then
  echo "carrier CA bundle is required in either $SOURCE_CERT_DIR or $CERT_DIR" >&2
  exit 2
fi

stage=$(mktemp -d "${TMPDIR:-/tmp}/leamout-cert-rotation.XXXXXX")
backup=$(mktemp -d "${TMPDIR:-/tmp}/leamout-cert-backup.XXXXXX")
activated=0

cleanup() {
  status=$?
  trap - EXIT INT TERM
  if [ "$status" -ne 0 ] && [ "$activated" -eq 1 ]; then
    echo "Certificate activation failed; restoring the previous runtime certificate set." >&2
    for name in fullchain.pem privkey.pem carrier-ca.pem; do
      if [ -f "$backup/$name" ] && [ -f "$CERT_DIR/$name" ]; then
        cat "$backup/$name" > "$CERT_DIR/$name" || true
      fi
    done
    CERT_DIR="$CERT_DIR" sh scripts/certs/check-certs.sh >/dev/null 2>&1 || true
  fi
  rm -rf "$stage" "$backup"
  exit "$status"
}
trap cleanup EXIT INT TERM

cp "$SOURCE_CERT_DIR/fullchain.pem" "$stage/fullchain.pem"
cp "$SOURCE_CERT_DIR/privkey.pem" "$stage/privkey.pem"
if [ -s "$SOURCE_CERT_DIR/carrier-ca.pem" ]; then
  cp "$SOURCE_CERT_DIR/carrier-ca.pem" "$stage/carrier-ca.pem"
else
  cp "$CERT_DIR/carrier-ca.pem" "$stage/carrier-ca.pem"
fi
chmod 0644 "$stage/fullchain.pem" "$stage/carrier-ca.pem"
chmod 0600 "$stage/privkey.pem"
CERT_DIR="$stage" sh scripts/certs/check-certs.sh

mkdir -p "$CERT_DIR"
for name in fullchain.pem privkey.pem carrier-ca.pem; do
  if [ -f "$CERT_DIR/$name" ]; then
    cp "$CERT_DIR/$name" "$backup/$name"
  fi
done

# Draining before activation makes the OpenSIPS restart deterministic and also
# gives Coturn a no-active-call window for its TLS context reload.
echo "Draining active calls before certificate activation..."
ENV_FILE="$ENV_FILE" COMPOSE_FILE="$COMPOSE_FILE" sh deploy/drain.sh

# These certificate files are individual Docker bind mounts. Overwrite the
# existing inode instead of renaming a new file over it, otherwise a running
# container may keep seeing the inode that was mounted at container creation.
for name in fullchain.pem privkey.pem carrier-ca.pem; do
  if [ ! -f "$CERT_DIR/$name" ]; then
    : > "$CERT_DIR/$name"
  fi
  cat "$stage/$name" > "$CERT_DIR/$name"
done
chmod 0644 "$CERT_DIR/fullchain.pem" "$CERT_DIR/carrier-ca.pem"
chmod 0600 "$CERT_DIR/privkey.pem"
activated=1

CERT_DIR="$CERT_DIR" sh scripts/certs/check-certs.sh

# Coturn supports TLS certificate reload on SIGUSR2. It is not part of the
# telecom drain stop set, so reload its TLS context after the runtime files are
# updated. If Coturn is not currently running, normal Compose startup will read
# the new files instead.
if docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps --status running --services | grep -qx coturn; then
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" kill -s SIGUSR2 coturn >/dev/null
fi

ENV_FILE="$ENV_FILE" COMPOSE_FILE="$COMPOSE_FILE" sh deploy/resume.sh
activated=0

echo "TLS certificate rotation completed successfully."
