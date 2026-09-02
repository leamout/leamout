#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib.sh"

sh "$SCRIPT_DIR/preflight.sh"
sh "$SCRIPT_DIR/build.sh"
compose up -d --no-build
