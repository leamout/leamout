#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
COMPOSE_FILE=${COMPOSE_FILE:-deploy/self-hosted/compose.yaml}
ENV_FILE=${ENV_FILE:-.env}

run_compose() {
  (cd "$REPO_ROOT" && docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@")
}

echo "Resuming FreeSWITCH call admission..."
run_compose exec -T freeswitch /usr/local/bin/leamout-freeswitch-drain resume >/dev/null

echo "Clearing OpenSIPS admission drain..."
run_compose exec -T opensips /usr/local/bin/leamout-opensips-drain resume >/dev/null

echo "Telecom node is accepting new calls."
