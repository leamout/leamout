#!/bin/sh
set -eu

: "${FREESWITCH_ESL_PASSWORD:?FREESWITCH_ESL_PASSWORD must be set}"

fs() {
  fs_cli \
    -H 127.0.0.1 \
    -P 8021 \
    -p "$FREESWITCH_ESL_PASSWORD" \
    -x "$1"
}

case "${1:-}" in
  enable)
    fs 'fsctl pause' >/dev/null
    echo "draining"
    ;;
  disable|resume)
    fs 'fsctl resume' >/dev/null
    echo "accepting"
    ;;
  status)
    state=$(fs 'fsctl pause_check' | tr -d '\r' | tail -n 1)
    case "$state" in
      true) echo "draining" ;;
      false) echo "accepting" ;;
      *)
        echo "Unexpected FreeSWITCH pause state: $state" >&2
        exit 1
        ;;
    esac
    ;;
  channels)
    response=$(fs 'show channels count' | tr -d '\r')
    count=$(printf '%s\n' "$response" | sed -n 's/^\([0-9][0-9]*\)[[:space:]]\+total.*/\1/p' | head -n 1)
    [ -n "$count" ] || {
      echo "Unable to parse FreeSWITCH active channel count" >&2
      printf '%s\n' "$response" >&2
      exit 1
    }
    echo "$count"
    ;;
  *)
    echo "usage: $0 {enable|disable|resume|status|channels}" >&2
    exit 2
    ;;
esac
