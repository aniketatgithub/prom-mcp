# Launch plan (local notes)

## LIVE 2026-10-09
- Repo: https://github.com/aniketatgithub/prom-mcp (public, topics set)
- Release: https://github.com/aniketatgithub/prom-mcp/releases/tag/v0.1.0
- Distribution drafts held in `distribution/` (awesome-mcp-servers + modelcontextprotocol/servers entries); submit after demo GIF lands in README.

## Concept and why it wins
- Agent infrastructure is the 2026 breakout lane (GitHub search: ~79k mcp-topic repos and ~83k claude-code repos created this year).
- Prometheus is the #1 observability backend and Aniket has domain depth + live Prometheus OSS PRs, which the MCP lane rewards (codebase-memory-mcp reached 46k stars on domain depth).
- No dominant Prometheus MCP server exists yet. White space confirmed by the breakout analysis.

## Repo name options (pick at launch)
1. prom-mcp (clear, searchable) - recommended
2. promcp (shorter to type)
3. prometheus-mcp-agent (descriptive, longer)

## Description at creation
"Give your AI agent production-grade eyes on Prometheus. MCP server for PromQL queries, alert explanations, and series discovery. Single Go binary."

## Topics
mcp, model-context-protocol, prometheus, observability, promql, ai-agents, claude-code, go, sre, devops

## Distribution (in order)
1. PR to modelcontextprotocol/servers + punkpeye/awesome-mcp-servers once public.
2. Submit to MCP registries: mcp.so, PulseMCP, Smithery.
3. Posts: r/golang, r/Prometheus, r/ClaudeAI, Show HN, Prometheus community Slack/Discord, MCP Discord.
4. Cross-link from his Prometheus OSS activity (profile README line).

## First-post draft (r/golang / HN, adapt per venue)
Title: prom-mcp: an MCP server that lets AI agents query Prometheus and explain alerts
Body: "I work on production systems on-call and kept watching agents hallucinate metric names against raw Prometheus APIs. prom-mcp gives them three tools instead: instant PromQL with compact output, alert explanations joined with the rule that fired them, and series discovery so they check what exists before querying. Single Go binary, no dependencies, `claude mcp add prom -- prom-mcp`. Demo mode runs against fixtures, no server needed: `prom-mcp demo`. Feedback on the tool shapes welcome."

## Remaining before public launch
- Record the 15s demo GIF (agent triage question against a real or fixture-backed Prometheus).
- Tag v0.1.0; optional goreleaser binaries.
- Natural v0.2 scope: range queries, target health, label values. Not before launch feedback.
