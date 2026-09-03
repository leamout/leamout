#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib.sh"

sh "$SCRIPT_DIR/preflight.sh"
sh "$SCRIPT_DIR/build.sh"

if [ -n "${DEPLOY_SERVICES:-}" ]; then
  # shellcheck disable=SC2086
  compose up -d --no-build $DEPLOY_SERVICES
else
  compose up -d --no-build
fi
