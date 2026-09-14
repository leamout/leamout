#!/bin/sh
set -eu

cat >&2 <<'MESSAGE'
Uninstall is owned by the leamout CLI but is not implemented in this release.
Use `leamout down` to stop the deployment. No data was removed.
MESSAGE
exit 1
