#!/bin/sh
set -eu

VERSION="${VERSION:?VERSION is required}"
OUT_DIR="${OUT_DIR:-dist}"
ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../../.." && pwd)"
RUNTIME_DIR="$ROOT_DIR/server/release/runtime"
MIGRATIONS_DIR="$ROOT_DIR/server/migrations"
ARCHIVE="leamout_runtime_${VERSION}_linux_amd64.tar.gz"

case "$VERSION" in
  ''|*[!0-9A-Za-z.+-]*)
    echo "invalid VERSION: $VERSION" >&2
    exit 1
    ;;
esac

for path in \
  "$RUNTIME_DIR/compose.yaml.tmpl" \
  "$RUNTIME_DIR/coturn/turnserver.conf" \
  "$MIGRATIONS_DIR/atlas.sum"; do
  [ -f "$path" ] || { echo "required runtime asset missing: $path" >&2; exit 1; }
done

mkdir -p "$OUT_DIR"
workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT HUP INT TERM
mkdir -p "$workdir/runtime/coturn" "$workdir/runtime/migrations"

cp "$RUNTIME_DIR/compose.yaml.tmpl" "$workdir/runtime/compose.yaml.tmpl"
cp "$RUNTIME_DIR/coturn/turnserver.conf" "$workdir/runtime/coturn/turnserver.conf"
cp "$MIGRATIONS_DIR/atlas.sum" "$workdir/runtime/migrations/atlas.sum"
cp "$MIGRATIONS_DIR"/*.sql "$workdir/runtime/migrations/"

TZ=UTC tar \
  --sort=name \
  --mtime='UTC 1970-01-01' \
  --owner=0 \
  --group=0 \
  --numeric-owner \
  -C "$workdir" \
  -cf - runtime \
  | gzip -n > "$OUT_DIR/$ARCHIVE"

(
  cd "$OUT_DIR"
  sha256sum "$ARCHIVE" > "$ARCHIVE.sha256"
)

printf '%s\n' "$OUT_DIR/$ARCHIVE"
