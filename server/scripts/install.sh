#!/bin/sh
set -eu

VERSION="${LEAMOUT_VERSION:-}"
BASE_URL="${LEAMOUT_RELEASE_BASE_URL:-https://get.leamout.com/releases}"
MINISIGN_PUBLIC_KEY="${LEAMOUT_MINISIGN_PUBLIC_KEY:-}"
OS_RELEASE_FILE="${LEAMOUT_OS_RELEASE_FILE:-/etc/os-release}"

usage() {
  cat <<'EOF'
Install Leamout Self-Hosted.

Usage:
  install.sh [--version <version>]

If no version is supplied, the current stable release is installed.
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || { echo "--version requires a value" >&2; exit 2; }
      VERSION="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

require_command() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Required installation tool not found: $1" >&2
    exit 1
  }
}

version_ge() {
  [ "$1" = "$2" ] || [ "$(printf '%s\n%s\n' "$2" "$1" | sort -V | tail -n 1)" = "$1" ]
}

for command in uname sed grep curl tar mktemp install sort tail; do
  require_command "$command"
done

if command -v sha256sum >/dev/null 2>&1; then
  SHA256="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA256="shasum -a 256"
else
  echo "Required installation tool for release verification is unavailable" >&2
  exit 1
fi

[ "$(uname -s)" = "Linux" ] || {
  echo "Unsupported operating system: $(uname -s). Leamout Self-Hosted supports Linux only." >&2
  exit 1
}

case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  *)
    echo "Unsupported architecture: $(uname -m). Leamout Self-Hosted supports amd64 only." >&2
    exit 1
    ;;
esac

[ -r "$OS_RELEASE_FILE" ] || {
  echo "Cannot determine Linux distribution" >&2
  exit 1
}

# shellcheck disable=SC1090
. "$OS_RELEASE_FILE"
case "${ID:-}:${VERSION_ID:-}" in
  ubuntu:24.04) MIN_KERNEL=6.8 ;;
  debian:13) MIN_KERNEL=6.12 ;;
  *)
    echo "Unsupported host: ${ID:-unknown} ${VERSION_ID:-unknown}. Supported: Ubuntu 24.04, Debian 13." >&2
    exit 1
    ;;
esac

kernel="$(uname -r | sed 's/-.*//')"
version_ge "$kernel" "$MIN_KERNEL" || {
  echo "Unsupported kernel: $kernel; ${ID} ${VERSION_ID} requires >= $MIN_KERNEL" >&2
  exit 1
}

if ! command -v systemctl >/dev/null 2>&1 || ! systemctl --version >/dev/null 2>&1; then
  echo "This host does not satisfy Leamout runtime requirements" >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "This host does not satisfy Leamout runtime requirements" >&2
  exit 1
fi

docker_version="$(docker version --format '{{.Server.Version}}' 2>/dev/null || true)"
[ -n "$docker_version" ] || {
  echo "This host does not satisfy Leamout runtime requirements" >&2
  exit 1
}
version_ge "$docker_version" "27.0" || {
  echo "This host does not satisfy Leamout runtime requirements" >&2
  exit 1
}

compose_version="$(docker compose version --short 2>/dev/null | sed 's/^v//' || true)"
[ -n "$compose_version" ] || {
  echo "This host does not satisfy Leamout runtime requirements" >&2
  exit 1
}
version_ge "$compose_version" "2.30" || {
  echo "This host does not satisfy Leamout runtime requirements" >&2
  exit 1
}

if [ -z "$VERSION" ]; then
  VERSION="$(curl -fsSL "$BASE_URL/stable.txt" | sed -n '1p')"
fi

case "$VERSION" in
  ''|*[!0-9A-Za-z.+-]*)
    echo "Invalid Leamout version: $VERSION" >&2
    exit 1
    ;;
esac

cli_artifact="leamout_${VERSION}_linux_${ARCH}.tar.gz"
runtime_artifact="leamout_runtime_${VERSION}_linux_${ARCH}.tar.gz"
release_url="$BASE_URL/$VERSION"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT HUP INT TERM

printf '✓ Host supported\n'
printf '✓ Host requirements satisfied\n'
printf 'Installing Leamout %s\n' "$VERSION"

for file in \
  "$cli_artifact" \
  "$runtime_artifact" \
  checksums.txt \
  checksums.txt.minisig \
  release-manifest.json \
  release-manifest.json.minisig; do
  curl -fsSL "$release_url/$file" -o "$workdir/$file"
done

if [ -z "$MINISIGN_PUBLIC_KEY" ]; then
  echo "Leamout release trust is not configured; refusing installation" >&2
  exit 1
fi
if ! command -v minisign >/dev/null 2>&1; then
  echo "Leamout release verification support is unavailable" >&2
  exit 1
fi
minisign -Vm "$workdir/checksums.txt" -P "$MINISIGN_PUBLIC_KEY" -x "$workdir/checksums.txt.minisig" >/dev/null
minisign -Vm "$workdir/release-manifest.json" -P "$MINISIGN_PUBLIC_KEY" -x "$workdir/release-manifest.json.minisig" >/dev/null

verify_artifact() {
  artifact="$1"
  expected="$(grep "  $artifact\$" "$workdir/checksums.txt" | sed -n '1{s/[[:space:]].*$//;p;}')"
  [ -n "$expected" ] || {
    echo "Leamout release verification failed" >&2
    exit 1
  }
  actual="$(cd "$workdir" && $SHA256 "$artifact" | sed 's/[[:space:]].*$//')"
  [ "$actual" = "$expected" ] || {
    echo "Leamout release verification failed" >&2
    exit 1
  }
}

verify_artifact "$cli_artifact"
verify_artifact "$runtime_artifact"

manifest_version="$(sed -n 's/^[[:space:]]*"release_version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$workdir/release-manifest.json" | sed -n '1p')"
[ "$manifest_version" = "$VERSION" ] || {
  echo "Leamout release manifest does not match requested version" >&2
  exit 1
}

mkdir -p "$workdir/extract"
tar -xzf "$workdir/$cli_artifact" -C "$workdir/extract"
[ -x "$workdir/extract/leamout" ] || {
  echo "Leamout release is incomplete" >&2
  exit 1
}

printf '✓ Leamout release verified\n'

install -d -m 0755 /usr/local/bin
install -m 0755 "$workdir/extract/leamout" /usr/local/bin/leamout
install -d -m 0750 /etc/leamout /var/lib/leamout /var/lib/leamout/releases /var/log/leamout
install -d -m 0750 "/var/lib/leamout/releases/$VERSION"
install -m 0640 "$workdir/release-manifest.json" "/var/lib/leamout/releases/$VERSION/release-manifest.json"
install -m 0640 "$workdir/checksums.txt" "/var/lib/leamout/releases/$VERSION/checksums.txt"
install -m 0640 "$workdir/$runtime_artifact" "/var/lib/leamout/releases/$VERSION/$runtime_artifact"

printf '✓ Leamout %s installed\n' "$VERSION"
printf '✓ Production runtime release staged\n'
printf 'Run: sudo leamout init\n'
