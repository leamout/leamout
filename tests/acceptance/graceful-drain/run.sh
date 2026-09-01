#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)

export GRACEFUL_DRAIN_SUITE_DIR="$SCRIPT_DIR"
export FREESWITCH_ESL_PASSWORD="${FREESWITCH_ESL_PASSWORD:-graceful-drain-esl-secret}"
export CARRIER_CREDENTIAL_ENCRYPTION_KEY="${CARRIER_CREDENTIAL_ENCRYPTION_KEY:-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA}"
export RTPENGINE_PUBLIC_IP="${RTPENGINE_PUBLIC_IP:-172.30.0.10}"
export TURN_AUTH_SECRET="${TURN_AUTH_SECRET:-graceful-drain-turn-secret-0123456789abcdef}"
export TURN_EXTERNAL_IP="${TURN_EXTERNAL_IP:-127.0.0.1}"
export TURN_PUBLIC_URLS="${TURN_PUBLIC_URLS:-turn:127.0.0.1:3478}"
export TURN_REALM="${TURN_REALM:-graceful-drain.local}"

COMPOSE="docker compose -f deploy/compose.yaml -f tests/acceptance/graceful-drain/compose.yaml"
COMPOSE_CONFIG_TMP=""

cleanup() {
    status=$?
    trap - EXIT INT TERM
    if [ -n "$COMPOSE_CONFIG_TMP" ]; then
        rm -f "$COMPOSE_CONFIG_TMP"
    fi
    if [ "$status" -ne 0 ]; then
        (cd "$REPO_ROOT" && $COMPOSE ps -a) || true
        (cd "$REPO_ROOT" && $COMPOSE exec -T graceful-drain-carrier fs_cli -H 127.0.0.1 -P 8021 -p "$FREESWITCH_ESL_PASSWORD" -x 'sofia status profile internal') || true
        (cd "$REPO_ROOT" && $COMPOSE exec -T freeswitch fs_cli -H 127.0.0.1 -P 8021 -p "$FREESWITCH_ESL_PASSWORD" -x 'sofia status profile internal') || true
        (cd "$REPO_ROOT" && $COMPOSE exec -T opensips /usr/local/bin/leamout-opensips-drain status) || true
        (cd "$REPO_ROOT" && $COMPOSE logs --no-color --tail=250 server worker opensips freeswitch rtpengine graceful-drain-carrier postgres) || true
    fi
    if [ "${GRACEFUL_DRAIN_KEEP_STACK:-0}" != "1" ]; then
        (cd "$REPO_ROOT" && $COMPOSE down -v --remove-orphans) >/dev/null 2>&1 || true
    fi
    exit "$status"
}
trap cleanup EXIT INT TERM

command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 1; }
command -v openssl >/dev/null 2>&1 || { echo "openssl is required" >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "python3 is required" >&2; exit 1; }

mkdir -p "$SCRIPT_DIR/certs"
openssl req \
    -x509 \
    -newkey rsa:2048 \
    -nodes \
    -keyout "$SCRIPT_DIR/certs/privkey.pem" \
    -out "$SCRIPT_DIR/certs/fullchain.pem" \
    -subj '/CN=graceful-drain.local' \
    -days 1 \
    >/dev/null 2>&1
cp "$SCRIPT_DIR/certs/fullchain.pem" "$SCRIPT_DIR/certs/carrier-ca.pem"

for fixture in fullchain.pem privkey.pem carrier-ca.pem; do
    test -s "$SCRIPT_DIR/certs/$fixture" || {
        echo "missing generated TLS fixture: $SCRIPT_DIR/certs/$fixture" >&2
        exit 1
    }
done

cd "$REPO_ROOT"
$COMPOSE config --quiet
COMPOSE_CONFIG_TMP=$(mktemp)
$COMPOSE config --format json > "$COMPOSE_CONFIG_TMP"
python3 - "$COMPOSE_CONFIG_TMP" "$SCRIPT_DIR/certs" <<'PY'
import json
import os
import sys

config_path, cert_dir = sys.argv[1:]
with open(config_path, encoding="utf-8") as handle:
    config = json.load(handle)

volumes = config["services"]["opensips"].get("volumes", [])
resolved = {
    item.get("target"): os.path.realpath(item.get("source", ""))
    for item in volumes
    if isinstance(item, dict)
}
expected = {
    "/etc/opensips/tls/fullchain.pem": os.path.realpath(
        os.path.join(cert_dir, "fullchain.pem")
    ),
    "/etc/opensips/tls/privkey.pem": os.path.realpath(
        os.path.join(cert_dir, "privkey.pem")
    ),
    "/etc/opensips/tls/carrier-ca.pem": os.path.realpath(
        os.path.join(cert_dir, "carrier-ca.pem")
    ),
}

missing = {
    target: {"expected": source, "resolved": resolved.get(target)}
    for target, source in expected.items()
    if resolved.get(target) != source
}
if missing:
    raise SystemExit(
        "resolved OpenSIPS TLS mounts do not use acceptance fixtures: "
        + json.dumps(missing, sort_keys=True)
    )
PY
rm -f "$COMPOSE_CONFIG_TMP"
COMPOSE_CONFIG_TMP=""

# TEMPORARY: regenerate and print the canonical Atlas checksum so the stale
# repository atlas.sum can be repaired from Atlas 1.3.0 itself. Remove this
# block after committing the generated sum.
docker run --rm \
    -v "$REPO_ROOT/server/migrations:/migrations" \
    arigaio/atlas:1.3.0-alpine \
    migrate hash --dir file:///migrations
echo '--- BEGIN GENERATED ATLAS SUM ---'
cat server/migrations/atlas.sum
echo '--- END GENERATED ATLAS SUM ---'

$COMPOSE up -d --build postgres redis nats rtpengine freeswitch graceful-drain-carrier
until $COMPOSE exec -T postgres pg_isready -U leamout -d leamout >/dev/null 2>&1; do
    sleep 1
done

$COMPOSE up --build migrate
$COMPOSE exec -T postgres \
    psql -v ON_ERROR_STOP=1 -U leamout -d leamout \
    < tests/acceptance/graceful-drain/bootstrap.sql \
    >/dev/null

$COMPOSE up -d --build opensips server worker

sip_ready=0
for _ in $(seq 1 60); do
    if $COMPOSE exec -T graceful-drain-carrier \
            fs_cli -H 127.0.0.1 -P 8021 -p "$FREESWITCH_ESL_PASSWORD" \
            -x 'sofia status profile internal' >/dev/null 2>&1 \
        && $COMPOSE exec -T freeswitch \
            fs_cli -H 127.0.0.1 -P 8021 -p "$FREESWITCH_ESL_PASSWORD" \
            -x 'sofia status profile internal' >/dev/null 2>&1 \
        && $COMPOSE exec -T opensips \
            /usr/local/bin/leamout-opensips-drain status >/dev/null 2>&1; then
        sip_ready=1
        break
    fi
    sleep 1
done

if [ "$sip_ready" -ne 1 ]; then
    echo "SIP services did not become ready within 60 seconds" >&2
    exit 1
fi

ready=0
for _ in $(seq 1 90); do
    if python3 - <<'PY'
import urllib.request
try:
    with urllib.request.urlopen('http://127.0.0.1:8080/readyz', timeout=2) as response:
        raise SystemExit(0 if response.status == 204 else 1)
except Exception:
    raise SystemExit(1)
PY
    then
        ready=1
        break
    fi
    sleep 1
done

if [ "$ready" -ne 1 ]; then
    echo "API did not become ready within 90 seconds" >&2
    exit 1
fi

python3 tests/acceptance/graceful-drain/acceptance.py
