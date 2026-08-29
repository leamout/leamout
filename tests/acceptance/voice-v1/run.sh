#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../../.." && pwd)
CERT_DIR=$(mktemp -d "${TMPDIR:-/tmp}/leamout-voice-v1.XXXXXX")

export VOICE_V1_SUITE_DIR="$SCRIPT_DIR"
export VOICE_V1_CERT_DIR="$CERT_DIR"
export FREESWITCH_ESL_PASSWORD="${FREESWITCH_ESL_PASSWORD:-voice-v1-esl-secret}"
export TURN_REALM="${TURN_REALM:-voice-v1.local}"
export TURN_AUTH_SECRET="${TURN_AUTH_SECRET:-voice-v1-turn-secret}"
export TURN_EXTERNAL_IP="${TURN_EXTERNAL_IP:-127.0.0.1}"
export RTPENGINE_PUBLIC_IP="${RTPENGINE_PUBLIC_IP:-172.30.0.10}"

COMPOSE="docker compose -f deploy/compose.yaml -f tests/acceptance/voice-v1/compose.yaml"

cleanup() {
    status=$?
    trap - EXIT INT TERM

    if [ "$status" -ne 0 ]; then
        printf '\n%s\n' "=== Voice v1 diagnostics: compose ps ==="
        (cd "$REPO_ROOT" && $COMPOSE ps -a) || true

        printf '\n%s\n' "=== Voice v1 diagnostics: service logs ==="
        (
            cd "$REPO_ROOT" &&
                $COMPOSE logs --no-color --tail=200 \
                    api \
                    worker \
                    freeswitch \
                    opensips \
                    rtpengine \
                    postgres \
                    nats \
                    redis \
                    voice-v1-carrier \
                    voice-v1-webhook
        ) || true
    fi

    if [ "${VOICE_V1_KEEP_STACK:-0}" != "1" ]; then
        (cd "$REPO_ROOT" && $COMPOSE down -v --remove-orphans) >/dev/null 2>&1 || true
        rm -rf "$CERT_DIR"
    else
        printf '%s\n' "Voice v1 stack retained; certificates: $CERT_DIR"
    fi
    exit "$status"
}
trap cleanup EXIT INT TERM

command -v docker >/dev/null 2>&1 || {
    echo "docker is required" >&2
    exit 1
}
command -v openssl >/dev/null 2>&1 || {
    echo "openssl is required" >&2
    exit 1
}
command -v python3 >/dev/null 2>&1 || {
    echo "python3 is required" >&2
    exit 1
}

cat >"$CERT_DIR/webhook.ext" <<'EOF'
subjectAltName=DNS:voice-v1-webhook,IP:127.0.0.1
extendedKeyUsage=serverAuth
EOF

openssl req \
    -x509 \
    -newkey rsa:2048 \
    -nodes \
    -days 1 \
    -keyout "$CERT_DIR/ca.key" \
    -out "$CERT_DIR/ca.crt" \
    -subj "/CN=Leamout Voice v1 Acceptance CA" \
    >/dev/null 2>&1

openssl req \
    -newkey rsa:2048 \
    -nodes \
    -keyout "$CERT_DIR/webhook.key" \
    -out "$CERT_DIR/webhook.csr" \
    -subj "/CN=voice-v1-webhook" \
    >/dev/null 2>&1

openssl x509 \
    -req \
    -in "$CERT_DIR/webhook.csr" \
    -CA "$CERT_DIR/ca.crt" \
    -CAkey "$CERT_DIR/ca.key" \
    -CAcreateserial \
    -days 1 \
    -out "$CERT_DIR/webhook.crt" \
    -extfile "$CERT_DIR/webhook.ext" \
    >/dev/null 2>&1
chmod 0644 "$CERT_DIR/ca.crt" "$CERT_DIR/webhook.crt"
chmod 0600 "$CERT_DIR/webhook.key"

cd "$REPO_ROOT"

$COMPOSE config --quiet
$COMPOSE up -d --build \
    postgres \
    redis \
    nats \
    rtpengine \
    freeswitch \
    opensips \
    voice-v1-carrier \
    voice-v1-webhook

printf '%s\n' "Waiting for PostgreSQL..."
i=0
until $COMPOSE exec -T postgres pg_isready -U leamout -d leamout >/dev/null 2>&1; do
    i=$((i + 1))
    if [ "$i" -ge 60 ]; then
        echo "PostgreSQL did not become ready" >&2
        exit 1
    fi
    sleep 1
done

printf '%s\n' "Applying migrations..."
for migration in server/migrations/*.sql; do
    $COMPOSE exec -T postgres \
        psql -v ON_ERROR_STOP=1 -U leamout -d leamout \
        <"$migration" \
        >/dev/null
done

$COMPOSE exec -T postgres \
    psql -v ON_ERROR_STOP=1 -U leamout -d leamout \
    <tests/acceptance/voice-v1/bootstrap.sql \
    >/dev/null

$COMPOSE up -d --build api worker

python3 tests/acceptance/voice-v1/acceptance.py
