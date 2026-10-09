# prom-mcp — Prometheus MCP Server for AI Agents (Claude Code, Cursor)

[![CI](https://github.com/aniketatgithub/prom-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/aniketatgithub/prom-mcp/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/go-1.23%2B-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](LICENSE)

**prom-mcp is a Prometheus MCP server (Model Context Protocol) that gives AI agents production-grade eyes on your metrics.** Connect Claude Code, Cursor, or any MCP client to Prometheus and let the agent run PromQL queries, explain firing alerts with their rule definitions, and discover which series actually exist before querying — so it investigates instead of guessing.

Built by a production engineer who works on-call, so the output reads like a triage note, not a raw JSON dump.

<!-- Demo GIF: agent asks "why is node down?" -> prom_alerts_explain answers with the firing rule, labels, and annotations. -->

## Install

```bash
go install github.com/aniketatgithub/prom-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/aniketatgithub/prom-mcp && cd prom-mcp
go build -o prom-mcp ./cmd/prom-mcp
```

## Quick start

```bash
export PROM_URL=http://localhost:9090   # your Prometheus (default if unset)
```

### Claude Code

```bash
claude mcp add prom -- prom-mcp
```

### Cursor (and other MCP clients)

Add to your MCP config (`~/.cursor/mcp.json` or equivalent):

```json
{
  "mcpServers": {
    "prom": {
      "command": "prom-mcp",
      "env": {
        "PROM_URL": "http://localhost:9090"
      }
    }
  }
}
```

Then ask your agent things like: *"What is firing right now?"* · *"Why is NodeDown alerting?"* · *"Show me `up` for the node job."*

## Tools

| Tool | What it does |
| --- | --- |
| `prom_query` | Runs an instant PromQL query and renders compact, labelled series and values. |
| `prom_alerts_explain` | Lists active alerts joined with their alerting-rule expressions and annotations, ready for triage. |
| `prom_series_discover` | Checks which series exist for a label matcher, so the agent stops hallucinating metric names. |

## Try it with no server

```bash
prom-mcp demo        # fixture-backed self-test of all three tools
prom-mcp query 'up'  # one-shot CLI query against $PROM_URL
```

## Configuration

| Setting | Env var | Config file key |
| --- | --- | --- |
| Prometheus base URL | `PROM_URL` | `base_url` |
| Bearer token (auth-protected Prometheus) | `PROM_TOKEN` | `token` |
| Request timeout | `PROM_TIMEOUT` (e.g. `30s`) | `timeout_seconds` |
| Config file path | `PROM_CONFIG` | |

Config file is JSON, looked up at `$PROM_CONFIG`, then `./prom-mcp.json`, then `~/.prom-mcp.json`:

```json
{ "base_url": "https://prometheus.example.com", "token": "…", "timeout_seconds": 30 }
```

Env vars override the file. No config at all defaults to `http://localhost:9090`.

## Why not just give the agent raw API access?

Agents drown in raw Prometheus JSON and invent metric names that do not exist. prom-mcp returns terse, labelled, triage-shaped text and makes the agent discover before querying. That is the difference between an agent that guesses and an agent that investigates.

Single Go binary, zero dependencies, works fully offline against your own Prometheus.

## FAQ

**Which AI clients work with prom-mcp?**
Any MCP client: Claude Code, Cursor, Windsurf, Zed, and custom agents. It speaks newline-delimited JSON-RPC 2.0 over stdio.

**How do I connect prom-mcp to a remote Prometheus?**
Set `PROM_URL` to your Prometheus base URL (for example `http://localhost:9090` or your internal endpoint) when adding the server.

**Can an AI agent explain why a Prometheus alert is firing?**
Yes. `prom_alerts_explain` joins active alerts with the rule that fired them (expression, summary, description), so the agent can reason about the cause instead of reporting raw JSON.

**Does prom-mcp need any API keys or cloud services?**
No. It is a single local binary talking only to your Prometheus.

## License

MIT. See [LICENSE](LICENSE).
