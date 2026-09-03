#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib.sh"

require_command docker
require_command git

test -f "$ENV_FILE" || {
  echo "Missing environment file: $ENV_FILE" >&2
  echo "Copy .env.example to .env and configure production values first." >&2
  exit 1
}

CERT_DIR="$CERT_DIR" sh server/scripts/certs/check-certs.sh
compose config --quiet

echo "Deployment preflight passed."
