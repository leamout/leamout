#!/bin/sh
set -eu

: "${VERSION:?VERSION is required}"
: "${CLI_BINARY:?CLI_BINARY is required}"

OUT_DIR="${OUT_DIR:-dist}"
REPO_ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
LICENSE_FILE="${LICENSE_FILE:-$REPO_ROOT/LICENSE}"
ARCH="${ARCH:-amd64}"
OS="${OS:-linux}"

case "$VERSION" in
  *[!0-9A-Za-z.+-]*|'')
    echo "VERSION contains unsupported characters: $VERSION" >&2
    exit 1
    ;;
esac

if [ "$OS" != "linux" ] || [ "$ARCH" != "amd64" ]; then
  echo "Phase 1 supports only linux/amd64 CLI artifacts" >&2
  exit 1
fi

for command in tar chmod mkdir cp mktemp; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "Required command not found: $command" >&2
    exit 1
  fi
done

if command -v sha256sum >/dev/null 2>&1; then
  checksum_command="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  checksum_command="shasum -a 256"
else
  echo "Required SHA-256 tool not found (sha256sum or shasum)" >&2
  exit 1
fi

[ -f "$CLI_BINARY" ] || {
  echo "CLI binary not found: $CLI_BINARY" >&2
  exit 1
}

[ -f "$LICENSE_FILE" ] || {
  echo "LICENSE file not found: $LICENSE_FILE" >&2
  exit 1
}

artifact="leamout_${VERSION}_${OS}_${ARCH}.tar.gz"
mkdir -p "$OUT_DIR"
OUT_DIR="$(CDPATH= cd -- "$OUT_DIR" && pwd)"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT HUP INT TERM

cp "$CLI_BINARY" "$workdir/leamout"
chmod 0755 "$workdir/leamout"
cp "$LICENSE_FILE" "$workdir/LICENSE"
chmod 0644 "$workdir/LICENSE"

# GNU tar options make the archive stable across repeated release builds.
# Release CI runs on Linux/GNU tar; Phase 2 installers only consume the archive.
tar \
  --sort=name \
  --mtime='UTC 1970-01-01' \
  --owner=0 \
  --group=0 \
  --numeric-owner \
  -C "$workdir" \
  -czf "$OUT_DIR/$artifact" \
  LICENSE leamout

(
  cd "$OUT_DIR"
  $checksum_command "$artifact" > checksums.txt
)

printf '%s\n' "$OUT_DIR/$artifact"
printf '%s\n' "$OUT_DIR/checksums.txt"
