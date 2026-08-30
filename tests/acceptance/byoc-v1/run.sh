#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
export FREESWITCH_ESL_PASSWORD="${FREESWITCH_ESL_PASSWORD:-byoc-v1-esl-secret}"
export BYOC_V1_SUITE_DIR="$SCRIPT_DIR"
export CARRIER_CREDENTIAL_ENCRYPTION_KEY="${CARRIER_CREDENTIAL_ENCRYPTION_KEY:-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA}"
export TURN_REALM="${TURN_REALM:-byoc-v1.local}"
export TURN_AUTH_SECRET="${TURN_AUTH_SECRET:-byoc-v1-turn-secret}"
export TURN_EXTERNAL_IP="${TURN_EXTERNAL_IP:-127.0.0.1}"
export RTPENGINE_PUBLIC_IP="${RTPENGINE_PUBLIC_IP:-172.30.0.10}"
COMPOSE="docker compose -f deploy/compose.yaml -f tests/acceptance/byoc-v1/compose.yaml"

cleanup() {
    status=$?
    trap - EXIT INT TERM
    if [ "$status" -ne 0 ]; then
        (cd "$REPO_ROOT" && $COMPOSE ps -a) || true
        (cd "$REPO_ROOT" && $COMPOSE logs --no-color --tail=200 api opensips freeswitch byoc-v1-carrier postgres) || true
    fi
    if [ "${BYOC_V1_KEEP_STACK:-0}" != "1" ]; then
        (cd "$REPO_ROOT" && $COMPOSE down -v --remove-orphans) >/dev/null 2>&1 || true
    fi
    exit "$status"
}
trap cleanup EXIT INT TERM
command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "python3 is required" >&2; exit 1; }
cd "$REPO_ROOT"
$COMPOSE config --quiet
$COMPOSE up -d --build postgres redis nats rtpengine freeswitch byoc-v1-carrier
until $COMPOSE exec -T postgres pg_isready -U leamout -d leamout >/dev/null 2>&1; do sleep 1; done
for migration in server/migrations/*.sql; do
    $COMPOSE exec -T postgres psql -v ON_ERROR_STOP=1 -U leamout -d leamout <"$migration" >/dev/null
done
$COMPOSE exec -T postgres psql -v ON_ERROR_STOP=1 -U leamout -d leamout <tests/acceptance/byoc-v1/bootstrap.sql >/dev/null
$COMPOSE up -d --build opensips api worker
python3 tests/acceptance/byoc-v1/acceptance.py
