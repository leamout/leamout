#!/bin/sh
set -eu

config=${OPENSIPS_CONFIG:-/etc/opensips/opensips.cfg}
advertised_address=${OPENSIPS_ADVERTISED_ADDRESS:-}

if [ -n "$advertised_address" ]; then
  case "$advertised_address" in
    *[!A-Za-z0-9._-]*)
      echo "OPENSIPS_ADVERTISED_ADDRESS must be a hostname or IPv4 address" >&2
      exit 1
      ;;
  esac

  tmp=$(mktemp)
  awk -v address="$advertised_address" '
    BEGIN { replaced = 0 }

    /^[[:space:]]*#?[[:space:]]*advertised_address[[:space:]]*=/ {
      if (!replaced) {
        print "advertised_address = \"" address "\""
        print "alias = udp:" address ":5060"
        print "alias = tcp:" address ":5060"
        print "alias = tls:" address ":5061"
        print "alias = wss:" address ":5062"
        replaced = 1
      }
      next
    }

    { print }

    END {
      if (!replaced) {
        exit 42
      }
    }
  ' "$config" > "$tmp" || {
    rc=$?
    rm -f "$tmp"
    if [ "$rc" -eq 42 ]; then
      echo "advertised_address anchor is missing from $config" >&2
    fi
    exit "$rc"
  }
  cat "$tmp" > "$config"
  rm -f "$tmp"

  # Fail before the daemon starts if the runtime-advertised route set is not a
  # valid OpenSIPS configuration. This address is what peers use for ACK/BYE.
  opensips -C -f "$config"
fi

exec "$@"
