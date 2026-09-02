#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib.sh"

services="${BUILD_SERVICES:-server worker web console waitlist opensips freeswitch rtpengine}"

for service in $services; do
  echo "==> Building $service"
  compose build "$service"
done
