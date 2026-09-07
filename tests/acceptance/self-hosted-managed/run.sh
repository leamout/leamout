#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
CERT_DIR=$(mktemp -d "${TMPDIR:-/tmp}/leamout-self-hosted-managed.XXXXXX")
export SELF_HOSTED_MANAGED_CERT_DIR="$CERT_DIR"
export MANAGED_SIP_ADMISSION_SECRET="${MANAGED_SIP_ADMISSION_SECRET:-$(openssl rand -hex 32)}"
export MANAGED_INBOUND_EDGE_ENABLED=true
export FREESWITCH_ESL_PASSWORD="${FREESWITCH_ESL_PASSWORD:-self-hosted-managed-esl}"
export CARRIER_CREDENTIAL_ENCRYPTION_KEY="${CARRIER_CREDENTIAL_ENCRYPTION_KEY:-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA}"
export TURN_AUTH_SECRET="${TURN_AUTH_SECRET:-$(openssl rand -hex 32)}"
export TURN_PUBLIC_URLS="${TURN_PUBLIC_URLS:-turn:127.0.0.1:3478}"
export TURN_REALM="${TURN_REALM:-self-hosted-managed.local}"
export TURN_EXTERNAL_IP="${TURN_EXTERNAL_IP:-127.0.0.1}"
export RTPENGINE_PUBLIC_IP="${RTPENGINE_PUBLIC_IP:-172.31.0.10}"
COMPOSE="docker compose -f deploy/compose.yaml -f tests/acceptance/self-hosted-managed/compose.yaml"

cleanup() {
    status=$?; trap - EXIT INT TERM
    if [ "$status" -ne 0 ]; then
        (cd "$REPO_ROOT" && $COMPOSE ps -a) || true
        (cd "$REPO_ROOT" && $COMPOSE logs --no-color --tail=400 server opensips self-hosted-opensips freeswitch postgres) || true
    fi
    if [ "${SELF_HOSTED_MANAGED_KEEP_STACK:-0}" != "1" ]; then
        (cd "$REPO_ROOT" && $COMPOSE down -v --remove-orphans) >/dev/null 2>&1 || true
        rm -rf "$CERT_DIR"
    fi
    exit "$status"
}
trap cleanup EXIT INT TERM

openssl req -x509 -newkey rsa:2048 -nodes -keyout "$CERT_DIR/privkey.pem" \
    -out "$CERT_DIR/fullchain.pem" -subj '/CN=managed-edge.test' -days 1 >/dev/null 2>&1
cp "$CERT_DIR/fullchain.pem" "$CERT_DIR/carrier-ca.pem"

cd "$REPO_ROOT"
$COMPOSE config --quiet
$COMPOSE up -d --build postgres redis nats rtpengine freeswitch
until $COMPOSE exec -T postgres pg_isready -U leamout -d leamout >/dev/null 2>&1; do sleep 1; done
$COMPOSE up --build migrate
$COMPOSE exec -T postgres psql -v ON_ERROR_STOP=1 -U leamout -d leamout <tests/acceptance/self-hosted-managed/bootstrap.sql >/dev/null
$COMPOSE up -d --build server opensips self-hosted-opensips

for _ in $(seq 1 90); do
    if python3 -c 'import urllib.request; assert urllib.request.urlopen("http://127.0.0.1:8080/readyz", timeout=2).status == 204' >/dev/null 2>&1; then
        python3 tests/acceptance/self-hosted-managed/acceptance.py
        exit
    fi
    sleep 1
done
echo "self-hosted managed stack did not become ready" >&2
exit 1
