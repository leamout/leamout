#!/bin/sh

SCRIPT_DIR="${SCRIPT_DIR:-$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)}"
REPO_ROOT="$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)"

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"
COMPOSE_OVERRIDE_FILE="${COMPOSE_OVERRIDE_FILE:-}"
CERT_DIR="${CERT_DIR:-deploy/certs}"

export ENV_FILE COMPOSE_FILE COMPOSE_OVERRIDE_FILE CERT_DIR

cd "$REPO_ROOT"

compose() {
  if [ -n "$COMPOSE_OVERRIDE_FILE" ]; then
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" -f "$COMPOSE_OVERRIDE_FILE" "$@"
  else
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
  fi
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Required command not found: $1" >&2
    exit 1
  fi
}
