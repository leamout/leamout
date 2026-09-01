#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
compose="docker compose -f $script_dir/compose.yaml"

run_compose() {
  # shellcheck disable=SC2086
  $compose "$@"
}

echo "Resuming FreeSWITCH call admission..."
run_compose exec -T freeswitch /usr/local/bin/leamout-freeswitch-drain resume >/dev/null

echo "Clearing OpenSIPS admission drain..."
run_compose exec -T opensips /usr/local/bin/leamout-opensips-drain resume >/dev/null

echo "Telecom node is accepting new calls."
