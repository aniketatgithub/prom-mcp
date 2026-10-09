#!/usr/bin/env bash
# Build .mcpb bundles for prom-mcp into dist/.
#
# Each bundle is a zip with:
#   manifest.json      (from mcpb/manifest.json)
#   server/prom-mcp    (static Go binary for one platform)
#
# Usage: mcpb/build.sh [version]
# Version defaults to the "version" field of mcpb/manifest.json.
# After building, SHA-256 hashes are printed: paste them into the
# fileSha256 fields of server.json before publishing to the MCP Registry.
set -euo pipefail

cd "$(dirname "$0")/.."

VERSION="${1:-$(python3 -c 'import json; print(json.load(open("mcpb/manifest.json"))["version"])')}"
TARGETS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64"

mkdir -p dist

for target in $TARGETS; do
  os="${target%/*}"
  arch="${target#*/}"
  stage="$(mktemp -d)"
  trap 'rm -rf "$stage"' EXIT

  mkdir -p "$stage/server"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" \
    go build -trimpath -ldflags "-s -w" -o "$stage/server/prom-mcp" ./cmd/prom-mcp
  cp mcpb/manifest.json "$stage/manifest.json"

  out="dist/prom-mcp_v${VERSION}_${os}_${arch}.mcpb"
  STAGE="$stage" OUT="$out" python3 - <<'PY'
import os, zipfile

stage = os.environ["STAGE"]
out = os.environ["OUT"]
entries = []
for root, dirs, files in os.walk(stage):
    dirs.sort()
    for name in sorted(files):
        full = os.path.join(root, name)
        entries.append((os.path.relpath(full, stage), full))
entries.sort()

# Fixed timestamps and permissions keep bundle bytes reproducible, so the
# SHA-256 stays stable across rebuilds with the same Go toolchain.
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    for arc, full in entries:
        info = zipfile.ZipInfo(arc, date_time=(1980, 1, 1, 0, 0, 0))
        info.create_system = 3
        info.compress_type = zipfile.ZIP_DEFLATED
        info.external_attr = (0o755 if os.access(full, os.X_OK) else 0o644) << 16
        with open(full, "rb") as fh:
            z.writestr(info, fh.read())
PY
  rm -rf "$stage"
  trap - EXIT
  echo "built $out"
done

echo
echo "SHA-256 (for server.json fileSha256):"
(cd dist && sha256sum prom-mcp_v"${VERSION}"_*.mcpb)
