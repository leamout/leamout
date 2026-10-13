#!/bin/sh
set -eu
ROOT=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
cd "$ROOT"
export APP_ENV_FILE=../server/.env
python3 deploy/cloud/preflight.py
compose() {
  docker compose --env-file .env -f deploy/compose.yaml -f deploy/cloud/compose.yaml "$@"
}
for name in postgres redis nats minio server worker media freeswitch rtpengine opensips caddy; do
  id=$(compose ps -q "$name")
  if [ -z "$id" ]; then
    echo "FAIL: $name is not running" >&2
    exit 1
  fi
  state=$(docker inspect --format '{{.State.Status}}' "$id")
  health=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$id")
  if [ "$state" != running ] || { [ "$health" != none ] && [ "$health" != healthy ]; }; then
    echo "FAIL: $name state=$state health=$health" >&2
    exit 1
  fi
  echo "PASS: $name"
done
domain=$(sed -n 's/^DOMAIN=//p' server/.env | tr -d "'\"")
if [ -z "$domain" ]; then
  echo "FAIL: DOMAIN is missing" >&2
  exit 1
fi
for path in healthz readyz; do
  curl --fail --silent --show-error --max-time 15 "https://api.$domain/$path" > /dev/null
  echo "PASS: api.$domain/$path"
done
echo "PASS: Cloud readiness smoke completed (live carrier and AI calls not tested)"
