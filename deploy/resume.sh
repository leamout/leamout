#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"

run_compose() {
  (cd "$repo_root" && docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@")
}

echo "Starting telecom services..."
run_compose up -d rtpengine freeswitch opensips

echo "Resuming FreeSWITCH call admission..."
run_compose exec -T freeswitch /usr/local/bin/leamout-freeswitch-drain resume >/dev/null

echo "Clearing OpenSIPS admission drain..."
run_compose exec -T opensips /usr/local/bin/leamout-opensips-drain resume >/dev/null

echo "Telecom node is accepting new calls."
