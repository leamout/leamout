#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib.sh"

sh "$SCRIPT_DIR/preflight.sh"
git pull --ff-only origin main
sh "$SCRIPT_DIR/build.sh"
compose up -d --no-build
sh "$SCRIPT_DIR/verify.sh"
