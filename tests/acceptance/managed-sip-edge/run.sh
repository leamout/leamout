#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
CERT_DIR=$(mktemp -d "${TMPDIR:-/tmp}/leamout-managed-sip-edge.XXXXXX")
export MANAGED_SIP_EDGE_SUITE_DIR="$SCRIPT_DIR"
export MANAGED_SIP_EDGE_CERT_DIR="$CERT_DIR"
export MANAGED_SIP_ADMISSION_SECRET="${MANAGED_SIP_ADMISSION_SECRET:-$(openssl rand -hex 32)}"
export FREESWITCH_ESL_PASSWORD="${FREESWITCH_ESL_PASSWORD:-managed-edge-esl-secret}"
export CARRIER_CREDENTIAL_ENCRYPTION_KEY="${CARRIER_CREDENTIAL_ENCRYPTION_KEY:-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA}"
export TURN_AUTH_SECRET="${TURN_AUTH_SECRET:-managed-edge-turn-secret-0123456789abcdef}"
export TURN_PUBLIC_URLS="${TURN_PUBLIC_URLS:-turn:127.0.0.1:3478}"
export TURN_REALM="${TURN_REALM:-managed-edge.local}"
export TURN_EXTERNAL_IP="${TURN_EXTERNAL_IP:-127.0.0.1}"
export RTPENGINE_PUBLIC_IP="${RTPENGINE_PUBLIC_IP:-172.31.0.10}"
COMPOSE="docker compose -f deploy/compose.yaml -f tests/acceptance/managed-sip-edge/compose.yaml"

cleanup() {
    status=$?; trap - EXIT INT TERM
    if [ "$status" -ne 0 ]; then
        (cd "$REPO_ROOT" && $COMPOSE ps -a) || true
        (cd "$REPO_ROOT" && $COMPOSE logs --no-color --tail=400 server opensips managed-sip-edge-wholesale postgres) || true
    fi
    if [ "${MANAGED_SIP_EDGE_KEEP_STACK:-0}" != "1" ]; then
        (cd "$REPO_ROOT" && $COMPOSE down -v --remove-orphans) >/dev/null 2>&1 || true
        rm -rf "$CERT_DIR"
    fi
    exit "$status"
}
trap cleanup EXIT INT TERM

command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 1; }
command -v openssl >/dev/null 2>&1 || { echo "openssl is required" >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "python3 is required" >&2; exit 1; }

openssl req -x509 -newkey rsa:2048 -nodes -keyout "$CERT_DIR/privkey.pem" \
    -out "$CERT_DIR/fullchain.pem" -subj '/CN=sip.leamout.com' -days 1 >/dev/null 2>&1
cp "$CERT_DIR/fullchain.pem" "$CERT_DIR/carrier-ca.pem"

cd "$REPO_ROOT"
$COMPOSE config --quiet
$COMPOSE up -d --build postgres redis nats rtpengine freeswitch managed-sip-edge-wholesale
until $COMPOSE exec -T postgres pg_isready -U leamout -d leamout >/dev/null 2>&1; do sleep 1; done
$COMPOSE up --build migrate
$COMPOSE exec -T postgres psql -v ON_ERROR_STOP=1 -U leamout -d leamout <tests/acceptance/managed-sip-edge/bootstrap.sql >/dev/null
$COMPOSE up -d --build opensips server

ready=0
for _ in $(seq 1 90); do
    if python3 -c 'import urllib.request; assert urllib.request.urlopen("http://127.0.0.1:8080/readyz", timeout=2).status == 204' >/dev/null 2>&1 \
        && python3 -c 'import urllib.request; assert urllib.request.urlopen("http://127.0.0.1:18088", timeout=2).status == 200' >/dev/null 2>&1; then
        ready=1; break
    fi
    sleep 1
done
[ "$ready" -eq 1 ] || { echo "managed SIP edge stack did not become ready" >&2; exit 1; }
python3 tests/acceptance/managed-sip-edge/acceptance.py
