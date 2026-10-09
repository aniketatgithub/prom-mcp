# modelcontextprotocol/servers submission (HELD, do not submit yet)

That repo lists official + community servers and accepts PRs adding a community server entry to its README.

Entry text (adjust to repo format when submitting):

### prom-mcp
- **Repository:** https://github.com/aniketatgithub/prom-mcp
- **Description:** Prometheus MCP server for AI agents. Tools: `prom_query` (instant PromQL), `prom_alerts_explain` (active alerts joined with rule expressions and annotations), `prom_series_discover` (series existence checks). Single Go binary, zero dependencies, fixture-backed demo mode.
- **Install:** `go install github.com/aniketatgithub/prom-mcp@latest`, then `claude mcp add prom -- prom-mcp` with `PROM_URL` set.

Status: prepared 2026-10-09. Hold submission until README demo GIF exists.
