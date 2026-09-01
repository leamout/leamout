#!/bin/sh
set -eu

config=${1:-/etc/opensips/opensips.cfg}

if ! grep -Fq 'include_file "drain.cfg"' "$config"; then
  sed -i '/include_file "tls.cfg"/a include_file "drain.cfg"' "$config"
fi

if ! grep -Fq 'route(LEAMOUT_DRAIN);' "$config"; then
  sed -i '/^    # REGISTER is authenticated against PostgreSQL/i\    route(LEAMOUT_DRAIN);\
' "$config"
fi

grep -Fq 'include_file "drain.cfg"' "$config"
grep -Fq 'route(LEAMOUT_DRAIN);' "$config"
