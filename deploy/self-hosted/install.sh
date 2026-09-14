#!/bin/sh
set -eu

command -v leamout >/dev/null 2>&1 || {
  echo "leamout CLI is not installed; install a verified release artifact first" >&2
  exit 1
}
exec leamout init "$@"
