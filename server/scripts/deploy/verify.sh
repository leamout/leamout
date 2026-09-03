#!/bin/sh
set -eu

SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
. "$SCRIPT_DIR/lib.sh"

ready=0
for _ in $(seq 1 "${HEALTH_RETRIES:-60}"); do
  if curl --fail --silent http://127.0.0.1:8080/healthz >/dev/null; then
    ready=1
    break
  fi
  sleep "${HEALTH_INTERVAL:-2}"
done

compose ps -a

test "$ready" -eq 1 || {
  echo "Server health check did not become ready" >&2
  exit 1
}

for service in server worker web console waitlist opensips coturn freeswitch rtpengine postgres redis nats; do
  container_id="$(compose ps -q "$service")"
  test -n "$container_id" || {
    echo "Missing container for service: $service" >&2
    exit 1
  }
  test "$(docker inspect -f '{{.State.Running}}' "$container_id")" = true || {
    echo "Service is not running: $service" >&2
    exit 1
  }
done

migrate_id="$(compose ps -aq migrate)"
test -n "$migrate_id" || {
  echo "Missing migrate container" >&2
  exit 1
}
test "$(docker inspect -f '{{.State.ExitCode}}' "$migrate_id")" -eq 0 || {
  echo "Migration service failed" >&2
  exit 1
}

echo "Deployment verification passed."
