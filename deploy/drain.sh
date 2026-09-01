#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
timeout_seconds=${LEAMOUT_DRAIN_TIMEOUT_SECONDS:-300}
poll_seconds=${LEAMOUT_DRAIN_POLL_SECONDS:-2}

case "$timeout_seconds" in
  ''|*[!0-9]*) echo "LEAMOUT_DRAIN_TIMEOUT_SECONDS must be a non-negative integer" >&2; exit 2 ;;
esac
case "$poll_seconds" in
  ''|*[!0-9]*) echo "LEAMOUT_DRAIN_POLL_SECONDS must be a non-negative integer" >&2; exit 2 ;;
esac
[ "$poll_seconds" -gt 0 ] || { echo "LEAMOUT_DRAIN_POLL_SECONDS must be greater than zero" >&2; exit 2; }

run_compose() {
  (cd "$script_dir" && docker compose "$@")
}

opensips_control() {
  run_compose exec -T opensips /usr/local/bin/leamout-opensips-drain "$1"
}

freeswitch_control() {
  run_compose exec -T freeswitch /usr/local/bin/leamout-freeswitch-drain "$1"
}

echo "Enabling OpenSIPS admission drain..."
opensips_control enable >/dev/null

echo "Pausing FreeSWITCH call admission..."
if ! freeswitch_control enable >/dev/null; then
  echo "FreeSWITCH pause failed; restoring OpenSIPS admission." >&2
  opensips_control disable >/dev/null 2>&1 || true
  exit 1
fi

started_at=$(date +%s)
while :; do
  dialogs=$(opensips_control dialogs)
  channels=$(freeswitch_control channels)

  echo "Drain progress: OpenSIPS dialogs=$dialogs FreeSWITCH channels=$channels"

  if [ "$dialogs" -eq 0 ] && [ "$channels" -eq 0 ]; then
    break
  fi

  now=$(date +%s)
  elapsed=$((now - started_at))
  if [ "$elapsed" -ge "$timeout_seconds" ]; then
    echo "Drain deadline reached with active calls remaining." >&2
    echo "The node remains drained. Run deploy/resume.sh to restore admission, or retry drain after calls finish." >&2
    exit 1
  fi

  sleep "$poll_seconds"
done

echo "No active signaling or FreeSWITCH sessions remain. Stopping telecom services..."
run_compose stop opensips
run_compose stop freeswitch
run_compose stop rtpengine

echo "Telecom node drained and stopped cleanly."
