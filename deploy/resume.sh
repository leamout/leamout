#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

run_compose() {
  (cd "$script_dir" && docker compose "$@")
}

echo "Resuming FreeSWITCH call admission..."
run_compose exec -T freeswitch /usr/local/bin/leamout-freeswitch-drain resume >/dev/null

echo "Clearing OpenSIPS admission drain..."
run_compose exec -T opensips /usr/local/bin/leamout-opensips-drain resume >/dev/null

echo "Telecom node is accepting new calls."
