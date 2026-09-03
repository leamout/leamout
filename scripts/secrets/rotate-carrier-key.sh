#!/bin/sh
set -eu

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"
NEW_KEY="${NEW_CARRIER_CREDENTIAL_ENCRYPTION_KEY:-}"
migration_done=0

if [ -z "$NEW_KEY" ]; then
  echo "NEW_CARRIER_CREDENTIAL_ENCRYPTION_KEY is required" >&2
  exit 2
fi
if [ ! -f "$ENV_FILE" ]; then
  echo "environment file does not exist: $ENV_FILE" >&2
  exit 2
fi
command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 2; }
command -v python3 >/dev/null 2>&1 || { echo "python3 is required" >&2; exit 2; }

old_key=$(awk -F= '
  /^[[:space:]]*CARRIER_CREDENTIAL_ENCRYPTION_KEY=/ {
    sub(/^[^=]*=/, ""); print; found=1; exit
  }
  END { if (!found) exit 1 }
' "$ENV_FILE") || {
  echo "CARRIER_CREDENTIAL_ENCRYPTION_KEY is missing from $ENV_FILE" >&2
  exit 2
}

if [ -z "$old_key" ]; then
  echo "current carrier credential encryption key is empty" >&2
  exit 2
fi
if [ "$old_key" = "$NEW_KEY" ]; then
  echo "new carrier credential encryption key matches the current key" >&2
  exit 2
fi

compose() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

cleanup() {
  status=$?
  trap - EXIT INT TERM
  if [ "$status" -ne 0 ] && [ "$migration_done" -eq 0 ]; then
    echo "Rotation failed before database commit; restarting API and worker with the current key." >&2
    compose up -d --no-deps server worker >/dev/null 2>&1 || true
  elif [ "$status" -ne 0 ]; then
    echo "Database rotation completed but activation did not finish." >&2
    echo "Do not start the API with the old key. Re-run this command with the same NEW_CARRIER_CREDENTIAL_ENCRYPTION_KEY." >&2
  fi
  exit "$status"
}
trap cleanup EXIT INT TERM

# Build the image containing the rotation utility before taking the API down.
compose build server

echo "Stopping API and worker so no carrier credential writes can race the key migration..."
compose stop server worker

echo "Re-encrypting stored carrier credentials transactionally..."
CARRIER_CREDENTIAL_ENCRYPTION_KEY_OLD="$old_key" \
CARRIER_CREDENTIAL_ENCRYPTION_KEY_NEW="$NEW_KEY" \
  compose run --rm --no-deps \
    -e CARRIER_CREDENTIAL_ENCRYPTION_KEY_OLD \
    -e CARRIER_CREDENTIAL_ENCRYPTION_KEY_NEW \
    server /leamout/rotate-carrier-key
migration_done=1

NEW_CARRIER_CREDENTIAL_ENCRYPTION_KEY="$NEW_KEY" \
  python3 scripts/secrets/set-env.py \
    "$ENV_FILE" CARRIER_CREDENTIAL_ENCRYPTION_KEY NEW_CARRIER_CREDENTIAL_ENCRYPTION_KEY

echo "Starting API and worker with the new key..."
compose up -d --no-deps server worker

ready=0
for _ in $(seq 1 60); do
  if compose exec -T server wget --spider -q http://127.0.0.1:8080/readyz >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done
if [ "$ready" -ne 1 ]; then
  echo "API did not become ready after carrier credential key rotation" >&2
  exit 1
fi

echo "Carrier credential encryption key rotation completed successfully."
