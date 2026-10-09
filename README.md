# prom-mcp

Give your AI agent production-grade eyes on Prometheus.

`prom-mcp` is an MCP server that lets Claude Code, Cursor, and other agents query your metrics, explain firing alerts with their rule definitions, and discover what series exist before writing PromQL. Built by a production engineer who works on-call, so the output reads like a triage note, not a JSON dump.

<!-- TODO before launch: 15-second demo GIF: agent asks "why is node down?", prom_alerts_explain answers with rule + summary. -->

## Install

```bash
go install github.com/aniketatgithub/prom-mcp@latest
```

## Use with your agent

```bash
export PROM_URL=http://localhost:9090   # your Prometheus
claude mcp add prom -- prom-mcp
```

No `PROM_URL`? It defaults to `http://localhost:9090`.

## Tools

- `prom_query` — instant PromQL query, rendered as compact series and values.
- `prom_alerts_explain` — active alerts joined with their alerting rule expressions and annotations, ready for triage.
- `prom_series_discover` — which series exist for a selector, so the agent stops guessing metric names.

## Try it with no server

```bash
prom-mcp demo        # fixture-backed self-test of all three tools
prom-mcp query 'up'  # one-shot CLI query against $PROM_URL
```

## Why not raw API access?

Agents drown in raw Prometheus JSON and hallucinate metric names. `prom-mcp` returns terse, labelled, triage-shaped text and forces discovery before querying. That is the difference between an agent that guesses and an agent that investigates.

Single Go binary, no dependencies, works fully offline against your own Prometheus.

## License

MIT. See [LICENSE](LICENSE).
