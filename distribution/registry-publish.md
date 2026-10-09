# Official MCP Registry publishing

Status: packaging ready, publish is one manual step (interactive login).

## What is in place

- `mcpb/manifest.json` - MCPB manifest (spec v0.3, binary server type),
  validated with `mcpb validate` from `@anthropic-ai/mcpb`.
- `mcpb/build.sh` - builds `dist/prom-mcp_v<version>_<os>_<arch>.mcpb`
  for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64. Bundles are
  plain zips: `manifest.json` plus `server/prom-mcp`. Zip timestamps and
  permissions are fixed, so rebuilds with the same Go toolchain produce
  identical bytes and stable SHA-256 hashes.
- `server.json` - Registry metadata for
  `io.github.aniketatgithub/prom-mcp`, validated with
  `mcp-publisher validate`. Its four `packages` entries point at the
  `.mcpb` release assets and carry their `fileSha256`.
- The release workflow builds the `.mcpb` bundles on every `v*` tag and
  uploads them with the binaries; `checksums.txt` covers both.

## Before publishing

The `fileSha256` values in `server.json` must match the `.mcpb` files
that end up on the GitHub release. After a release is cut:

1. Download the four `.mcpb` assets from the release (or read
   `checksums.txt` attached to it).
2. Compare hashes with `server.json`. If they differ (for example a
   newer Go toolchain changed the binary), update `fileSha256` and
   commit before publishing.

## Publish (manual, ~2 minutes)

Publishing needs an interactive GitHub login for namespace ownership,
so it cannot run from an unattended script:

```bash
# one-time: install the publisher CLI
curl -L "https://github.com/modelcontextprotocol/registry/releases/latest/download/mcp-publisher_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz" | tar xz mcp-publisher

./mcp-publisher login github   # browser sign-in as aniketatgithub
./mcp-publisher publish        # run from the repo root (reads ./server.json)
```

## Next release checklist

1. Bump `version` in `mcpb/manifest.json` and `server.json`
   (top level and each package) to the new version.
2. Tag `vX.Y.Z` and push the tag. The workflow attaches binaries,
   `.mcpb` bundles, and `checksums.txt`.
3. Verify `fileSha256` against the release `checksums.txt`; fix
   `server.json` if needed.
4. `./mcp-publisher publish`.
