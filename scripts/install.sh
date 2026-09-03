#!/bin/sh
set -eu

VERSION="${LEAMOUT_VERSION:-}"
BASE_URL="${LEAMOUT_RELEASE_BASE_URL:-https://get.leamout.com/releases}"
INSTALL_DIR="${LEAMOUT_INSTALL_DIR:-/usr/local/bin}"
MINISIGN_PUBLIC_KEY="${LEAMOUT_MINISIGN_PUBLIC_KEY:-}"
CONFIG_DIR="${LEAMOUT_CONFIG_DIR:-/etc/leamout}"
STATE_DIR="${LEAMOUT_STATE_DIR:-/var/lib/leamout}"
LOG_DIR="${LEAMOUT_LOG_DIR:-/var/log/leamout}"

usage() {
  cat <<'EOF'
Install the Leamout self-hosted operator CLI.

Usage:
  install.sh [--version <version>] [--base-url <url>] [--install-dir <dir>]

Environment:
  LEAMOUT_VERSION              exact CLI version to install
  LEAMOUT_RELEASE_BASE_URL     release root (default: https://get.leamout.com/releases)
  LEAMOUT_INSTALL_DIR          binary destination directory (default: /usr/local/bin)
  LEAMOUT_MINISIGN_PUBLIC_KEY  trusted Minisign public key for release verification
  LEAMOUT_CONFIG_DIR           base configuration directory (default: /etc/leamout)
  LEAMOUT_STATE_DIR            base state directory (default: /var/lib/leamout)
  LEAMOUT_LOG_DIR              base log directory (default: /var/log/leamout)

If no version is supplied, the installer resolves BASE_URL/stable.txt.
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || { echo "--version requires a value" >&2; exit 2; }
      VERSION="$2"
      shift 2
      ;;
    --base-url)
      [ "$#" -ge 2 ] || { echo "--base-url requires a value" >&2; exit 2; }
      BASE_URL="$2"
      shift 2
      ;;
    --install-dir)
      [ "$#" -ge 2 ] || { echo "--install-dir requires a value" >&2; exit 2; }
      INSTALL_DIR="$2"
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
    echo "Required command not found: $1" >&2
    exit 1
  }
}

version_ge() {
  [ "$1" = "$2" ] || [ "$(printf '%s\n%s\n' "$2" "$1" | sort -V | tail -n 1)" = "$1" ]
}

for command in uname sed grep curl tar mktemp install sort tail docker systemctl; do
  require_command "$command"
done

if command -v sha256sum >/dev/null 2>&1; then
  SHA256="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA256="shasum -a 256"
else
  echo "Required SHA-256 tool not found (sha256sum or shasum)" >&2
  exit 1
fi

[ "$(uname -s)" = "Linux" ] || {
  echo "Unsupported operating system: $(uname -s). Leamout Phase 2 supports Linux only." >&2
  exit 1
}

case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64 ;;
  *)
    echo "Unsupported architecture: $(uname -m). Leamout Phase 2 supports amd64 only." >&2
    exit 1
    ;;
esac

[ -r /etc/os-release ] || {
  echo "Cannot determine Linux distribution: /etc/os-release is missing" >&2
  exit 1
}

# shellcheck disable=SC1091
. /etc/os-release
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

systemctl --version >/dev/null 2>&1 || {
  echo "systemd is required by the supported host contract" >&2
  exit 1
}

docker_version="$(docker version --format '{{.Server.Version}}' 2>/dev/null || true)"
[ -n "$docker_version" ] || {
  echo "Docker Engine is required and its daemon must be reachable" >&2
  exit 1
}
version_ge "$docker_version" "27.0" || {
  echo "Docker Engine >= 27.0 is required; found $docker_version" >&2
  exit 1
}

compose_version="$(docker compose version --short 2>/dev/null | sed 's/^v//' || true)"
[ -n "$compose_version" ] || {
  echo "Docker Compose plugin >= 2.30 is required" >&2
  exit 1
}
version_ge "$compose_version" "2.30" || {
  echo "Docker Compose plugin >= 2.30 is required; found $compose_version" >&2
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

artifact="leamout_${VERSION}_linux_${ARCH}.tar.gz"
release_url="$BASE_URL/$VERSION"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT HUP INT TERM

printf '✓ Supported %s %s linux/%s host\n' "$ID" "$VERSION_ID" "$ARCH"
printf '✓ Docker Engine %s and Compose %s\n' "$docker_version" "$compose_version"
printf 'Installing Leamout CLI %s\n' "$VERSION"

curl -fsSL "$release_url/$artifact" -o "$workdir/$artifact"
curl -fsSL "$release_url/checksums.txt" -o "$workdir/checksums.txt"

if [ -n "$MINISIGN_PUBLIC_KEY" ]; then
  require_command minisign
  curl -fsSL "$release_url/checksums.txt.minisig" -o "$workdir/checksums.txt.minisig"
  minisign -Vm "$workdir/checksums.txt" -P "$MINISIGN_PUBLIC_KEY" -x "$workdir/checksums.txt.minisig" >/dev/null
else
  echo "Refusing unsigned production installation: LEAMOUT_MINISIGN_PUBLIC_KEY is not configured." >&2
  echo "A Leamout release trust root must be pinned before get.leamout.com is published." >&2
  exit 1
fi

expected="$(grep "  $artifact\$" "$workdir/checksums.txt" | sed -n '1{s/[[:space:]].*$//;p;}')"
[ -n "$expected" ] || {
  echo "Artifact checksum is missing from checksums.txt: $artifact" >&2
  exit 1
}
actual="$(cd "$workdir" && $SHA256 "$artifact" | sed 's/[[:space:]].*$//')"
[ "$actual" = "$expected" ] || {
  echo "SHA-256 verification failed for $artifact" >&2
  exit 1
}

mkdir -p "$workdir/extract"
tar -xzf "$workdir/$artifact" -C "$workdir/extract"
[ -x "$workdir/extract/leamout" ] || {
  echo "Release archive does not contain an executable leamout binary" >&2
  exit 1
}

install -d -m 0755 "$INSTALL_DIR"
install -m 0755 "$workdir/extract/leamout" "$INSTALL_DIR/leamout"
install -d -m 0755 "$CONFIG_DIR" "$STATE_DIR" "$LOG_DIR"

printf '✓ Leamout CLI %s installed at %s/leamout\n' "$VERSION" "$INSTALL_DIR"
printf '✓ Base directories initialized\n'
printf 'Run: sudo leamout init\n'
