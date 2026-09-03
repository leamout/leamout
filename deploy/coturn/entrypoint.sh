#!/bin/sh
set -eu

# The upstream coturn image normally prepends turnserver through its own
# entrypoint. Leamout overrides the entrypoint so it can conditionally append a
# previous TURN REST secret during rotation without ever rendering an empty
# static-auth-secret argument.
set -- turnserver "$@"

if [ -n "${TURN_AUTH_SECRET_PREVIOUS:-}" ]; then
  set -- "$@" "--static-auth-secret=${TURN_AUTH_SECRET_PREVIOUS}"
fi

exec "$@"
