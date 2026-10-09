# prom-mcp with Claude Code

```bash
go install github.com/aniketatgithub/prom-mcp@latest
export PROM_URL=http://localhost:9090
claude mcp add prom -- prom-mcp
```

With a token-protected Prometheus:

```bash
claude mcp add prom --env PROM_URL=https://prom.example.com --env PROM_TOKEN=xxx -- prom-mcp
```

Then try: "Use the triage prompt on the node job" or "Why is NodeDown firing?"
