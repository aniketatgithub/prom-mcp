# Changelog

## v0.2.0 (2026-10-09)

Ten-revision production hardening pass after the initial launch.

- Client: bounded retries on transient failures (network/5xx, no retry on 4xx), 15s default timeout, 16 MiB response cap, richer error messages with Prometheus errorType.
- New tools: `prom_query_range` (per-series points/first/last/min/max summary) and `prom_label_values`; series output capped at 100 with a refine hint.
- Alert explanations now include rule group, `for` duration, and all annotations.
- Configuration: `PROM_URL`, `PROM_TOKEN` (bearer auth), `PROM_TIMEOUT`, and JSON config file (`PROM_CONFIG`, `./prom-mcp.json`, `~/.prom-mcp.json`); env overrides file.
- MCP prompts/resources: `triage` prompt template; `prometheus://alerts` and `prometheus://config` resources.
- Tests: range/label fixtures, retry semantics, config resolution, MCP handler tests.
- Release workflow: cross-platform binaries (linux/darwin, amd64/arm64) with checksums on tags.
- Examples: Claude Code, Cursor, VS Code configs; Docker Compose Prometheus demo stack.

## v0.1.0 (2026-10-09)

- Initial release: `prom_query`, `prom_alerts_explain`, `prom_series_discover` over MCP stdio; fixture demo mode; one-shot CLI query.
