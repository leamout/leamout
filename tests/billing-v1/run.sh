#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
CONTAINER_NAME=""

cleanup() {
    status=$?
    trap - EXIT INT TERM
    if [ -n "$CONTAINER_NAME" ]; then
        if [ "$status" -ne 0 ]; then
            docker logs --tail=100 "$CONTAINER_NAME" >&2 || true
        fi
        docker rm -f -v "$CONTAINER_NAME" >/dev/null 2>&1 || true
    fi
    exit "$status"
}
trap cleanup EXIT INT TERM

command -v go >/dev/null 2>&1 || {
    echo "Go matching server/go.mod is required" >&2
    exit 1
}

if [ -z "${BILLING_TEST_DATABASE_URL:-}" ]; then
    command -v docker >/dev/null 2>&1 || {
        echo "Docker or BILLING_TEST_DATABASE_URL is required" >&2
        exit 1
    }
    CONTAINER_NAME="leamout-billing-acceptance-$$"
    docker run --detach --rm \
        --name "$CONTAINER_NAME" \
        --publish 127.0.0.1::5432 \
        --env POSTGRES_DB=leamout_billing_acceptance \
        --env POSTGRES_USER=leamout \
        --env POSTGRES_PASSWORD=billing-acceptance-only \
        postgres:17-alpine >/dev/null

    ready=0
    for _ in $(seq 1 60); do
        if docker exec "$CONTAINER_NAME" \
            psql -U leamout -d leamout_billing_acceptance \
            --tuples-only --command 'SELECT 1' >/dev/null 2>&1; then
            ready=1
            break
        fi
        sleep 1
    done
    if [ "$ready" -ne 1 ]; then
        echo "Disposable PostgreSQL did not become ready" >&2
        exit 1
    fi

    POSTGRES_PORT=$(docker port "$CONTAINER_NAME" 5432/tcp | sed -n 's/^127\.0\.0\.1://p')
    case "$POSTGRES_PORT" in
        ''|*[!0-9]*)
            echo "Could not determine disposable PostgreSQL port" >&2
            exit 1
            ;;
    esac
    BILLING_TEST_DATABASE_URL="postgres://leamout:billing-acceptance-only@127.0.0.1:$POSTGRES_PORT/leamout_billing_acceptance?sslmode=disable"
    export BILLING_TEST_DATABASE_URL
fi

cd "$REPO_ROOT/server"
if ! go test -list '^TestBillingAcceptance' ./internal/commercial/billing | grep -q '^TestBillingAcceptance'; then
    echo "No billing acceptance tests were discovered" >&2
    exit 1
fi
go test -race -count=1 -v -run '^TestBillingAcceptance' ./internal/commercial/billing

case "${BILLING_RUN_POLAR_SANDBOX_CHECKS:-0}" in
    1)
        command -v python3 >/dev/null 2>&1 || {
            echo "Python 3 is required for the optional sandbox checks" >&2
            exit 1
        }
        python3 "$SCRIPT_DIR/sandbox.py"
        ;;
    0)
        printf '%s\n' "Sandbox API checks skipped; no live Polar purchases were tested."
        ;;
    *)
        echo "BILLING_RUN_POLAR_SANDBOX_CHECKS must be 0 or 1" >&2
        exit 1
        ;;
esac
