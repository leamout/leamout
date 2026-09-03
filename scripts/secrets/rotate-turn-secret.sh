#!/bin/sh
set -eu

ENV_FILE="${ENV_FILE:-.env}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose.yaml}"
MIN_OVERLAP_SECONDS="${TURN_ROTATION_MIN_OVERLAP_SECONDS:-660}"
action="${1:-begin}"

if [ ! -f "$ENV_FILE" ]; then
  echo "environment file does not exist: $ENV_FILE" >&2
  exit 2
fi
command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 2; }
command -v python3 >/dev/null 2>&1 || { echo "python3 is required" >&2; exit 2; }

case "$MIN_OVERLAP_SECONDS" in
  ''|*[!0-9]*) echo "TURN_ROTATION_MIN_OVERLAP_SECONDS must be an integer" >&2; exit 2 ;;
esac
if [ "$MIN_OVERLAP_SECONDS" -lt 600 ]; then
  echo "TURN_ROTATION_MIN_OVERLAP_SECONDS must be at least 600 seconds" >&2
  exit 2
fi

read_env() {
  key=$1
  awk -F= -v key="$key" '
    $0 ~ "^[[:space:]]*" key "=" {
      sub(/^[^=]*=/, ""); print; found=1; exit
    }
    END { if (!found) exit 1 }
  ' "$ENV_FILE"
}

compose() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" "$@"
}

case "$action" in
  begin)
    NEW_SECRET="${NEW_TURN_AUTH_SECRET:-}"
    if [ ${#NEW_SECRET} -lt 32 ]; then
      echo "NEW_TURN_AUTH_SECRET must be at least 32 bytes" >&2
      exit 2
    fi
    current=$(read_env TURN_AUTH_SECRET) || {
      echo "TURN_AUTH_SECRET is missing from $ENV_FILE" >&2
      exit 2
    }
    previous=$(read_env TURN_AUTH_SECRET_PREVIOUS 2>/dev/null || true)
    if [ -n "$previous" ]; then
      echo "TURN_AUTH_SECRET_PREVIOUS is already set; finalize the existing rotation first" >&2
      exit 2
    fi
    if [ "$current" = "$NEW_SECRET" ]; then
      echo "new TURN secret matches the current secret" >&2
      exit 2
    fi

    echo "Draining active calls before changing the Coturn process..."
    ENV_FILE="$ENV_FILE" COMPOSE_FILE="$COMPOSE_FILE" sh deploy/drain.sh

    # Install the old secret as the overlap secret before switching issuance to
    # the new one. If the process stops between these writes, the still-running
    # services continue using their old environment and no credentials are lost.
    OLD_TURN_AUTH_SECRET="$current" \
      python3 scripts/secrets/set-env.py \
        "$ENV_FILE" TURN_AUTH_SECRET_PREVIOUS OLD_TURN_AUTH_SECRET
    NEW_TURN_AUTH_SECRET="$NEW_SECRET" \
      python3 scripts/secrets/set-env.py \
        "$ENV_FILE" TURN_AUTH_SECRET NEW_TURN_AUTH_SECRET
    ROTATED_AT=$(date +%s)
    TURN_AUTH_SECRET_ROTATED_AT="$ROTATED_AT" \
      python3 scripts/secrets/set-env.py \
        "$ENV_FILE" TURN_AUTH_SECRET_ROTATED_AT TURN_AUTH_SECRET_ROTATED_AT

    # Coturn must know the new secret before the API begins issuing credentials
    # signed by it. The entrypoint appends TURN_AUTH_SECRET_PREVIOUS only when it
    # is non-empty, giving us an explicit two-secret overlap window.
    compose up -d --force-recreate --no-deps coturn
    compose up -d --force-recreate --no-deps server

    ready=0
    for _ in $(seq 1 60); do
      if compose exec -T server wget --spider -q http://127.0.0.1:8080/readyz >/dev/null 2>&1; then
        ready=1
        break
      fi
      sleep 1
    done
    if [ "$ready" -ne 1 ]; then
      echo "API did not become ready after TURN secret rotation" >&2
      echo "The node remains drained; correct the configuration before resuming." >&2
      exit 1
    fi

    ENV_FILE="$ENV_FILE" COMPOSE_FILE="$COMPOSE_FILE" sh deploy/resume.sh
    echo "TURN secret overlap started."
    echo "Keep TURN_AUTH_SECRET_PREVIOUS for at least ${MIN_OVERLAP_SECONDS}s before finalizing."
    ;;

  finalize)
    previous=$(read_env TURN_AUTH_SECRET_PREVIOUS 2>/dev/null || true)
    if [ -z "$previous" ]; then
      echo "no TURN secret rotation is pending" >&2
      exit 2
    fi
    rotated_at=$(read_env TURN_AUTH_SECRET_ROTATED_AT 2>/dev/null || true)
    case "$rotated_at" in
      ''|*[!0-9]*) echo "TURN_AUTH_SECRET_ROTATED_AT is missing or invalid" >&2; exit 2 ;;
    esac
    now=$(date +%s)
    elapsed=$((now - rotated_at))
    if [ "$elapsed" -lt "$MIN_OVERLAP_SECONDS" ] && [ "${TURN_ROTATION_FORCE:-0}" != "1" ]; then
      remaining=$((MIN_OVERLAP_SECONDS - elapsed))
      echo "TURN secret overlap is still required for ${remaining}s" >&2
      exit 2
    fi

    echo "Draining active calls before removing the previous Coturn secret..."
    ENV_FILE="$ENV_FILE" COMPOSE_FILE="$COMPOSE_FILE" sh deploy/drain.sh

    EMPTY_VALUE="" ALLOW_EMPTY_ENV_VALUE=1 \
      python3 scripts/secrets/set-env.py \
        "$ENV_FILE" TURN_AUTH_SECRET_PREVIOUS EMPTY_VALUE
    EMPTY_VALUE="" ALLOW_EMPTY_ENV_VALUE=1 \
      python3 scripts/secrets/set-env.py \
        "$ENV_FILE" TURN_AUTH_SECRET_ROTATED_AT EMPTY_VALUE

    compose up -d --force-recreate --no-deps coturn
    ENV_FILE="$ENV_FILE" COMPOSE_FILE="$COMPOSE_FILE" sh deploy/resume.sh
    echo "TURN shared-secret rotation finalized; the previous secret is no longer accepted."
    ;;

  *)
    echo "usage: $0 [begin|finalize]" >&2
    exit 2
    ;;
esac
