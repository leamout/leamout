#!/bin/sh
set -eu

for script in /docker-entrypoint.d/*.sh; do
  [ -f "$script" ] || continue
  "$script"
done

exec /usr/local/freeswitch/bin/freeswitch -nonat -nf
