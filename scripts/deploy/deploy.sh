#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib.sh"

sh "$SCRIPT_DIR/preflight.sh"
git pull --ff-only origin main
compose up -d --build
sh "$SCRIPT_DIR/verify.sh"
