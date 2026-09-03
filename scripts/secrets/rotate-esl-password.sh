#!/bin/sh
set -eu

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"
NEW_PASSWORD="${NEW_FREESWITCH_ESL_PASSWORD:-}"

if [ -z "$NEW_PASSWORD" ]; then
  echo "NEW_FREESWITCH_ESL_PASSWORD is required" >&2
  exit 2
fi
if [ ! -f "$ENV_FILE" ]; then
  echo "environment file does not exist: $ENV_FILE" >&2
  exit 2
fi
command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 2; }
command -v python3 >/dev/null 2>&1 || { echo "python3 is required" >&2; exit 2; }

current=$(awk -F= '
  /^[[:space:]]*FREESWITCH_ESL_PASSWORD=/ {
    sub(/^[^=]*=/, ""); print; found=1; exit
  }
  END { if (!found) exit 1 }
' "$ENV_FILE") || {
  echo "FREESWITCH_ESL_PASSWORD is missing from $ENV_FILE" >&2
  exit 2
}
if [ "$current" = "$NEW_PASSWORD" ]; then
  echo "new FreeSWITCH ESL password matches the current password" >&2
  exit 2
fi

compose() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

# FreeSWITCH reads the event-socket password when its configuration is loaded,
# while API and worker read it from their process environments. Drain calls so
# those three processes can be recreated as one coordinated maintenance change.
echo "Draining active calls before ESL password rotation..."
ENV_FILE="$ENV_FILE" COMPOSE_FILE="$COMPOSE_FILE" sh deploy/drain.sh

NEW_FREESWITCH_ESL_PASSWORD="$NEW_PASSWORD" \
  python3 scripts/secrets/set-env.py \
    "$ENV_FILE" FREESWITCH_ESL_PASSWORD NEW_FREESWITCH_ESL_PASSWORD

echo "Recreating FreeSWITCH, API, and worker with the new ESL password..."
compose up -d --force-recreate --no-deps freeswitch server worker

ready=0
for _ in $(seq 1 60); do
  if compose exec -T freeswitch \
      fs_cli -H 127.0.0.1 -P 8021 -p "$NEW_PASSWORD" -x status >/dev/null 2>&1 \
    && compose exec -T server wget --spider -q http://127.0.0.1:8080/readyz >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done
if [ "$ready" -ne 1 ]; then
  echo "FreeSWITCH/API did not become ready with the new ESL password" >&2
  echo "The node remains drained; fix the configuration and run deploy/resume.sh." >&2
  exit 1
fi

ENV_FILE="$ENV_FILE" COMPOSE_FILE="$COMPOSE_FILE" sh deploy/resume.sh

echo "FreeSWITCH ESL password rotation completed successfully."
