# How prom-mcp compares

There are three Prometheus MCP servers worth knowing about. They aim at
different jobs, and this page says which is which, based on each project's
own README. Checked 2026-10-09 against:

- prom-mcp: this repo
- [pab1it0/prometheus-mcp-server](https://github.com/pab1it0/prometheus-mcp-server/blob/main/README.md)
- [prometheus/prometheus-mcp](https://github.com/prometheus/prometheus-mcp/blob/main/README.md)
  (the Prometheus org's server)

## The short version

- **prom-mcp** is the focused one: 5 tools built around alert triage. A
  firing alert comes back joined with the rule that fired it, discovery
  happens before queries, and output reads like a triage note.
- **pab1it0/prometheus-mcp-server** is a general PromQL query server with
  the richest container story: Docker, GHCR, Helm, and the Docker Desktop
  MCP Catalog.
- **prometheus/prometheus-mcp** is the broadest: most of the Prometheus
  HTTP API as tools, plus embedded runbooks and Prometheus docs search,
  from the Prometheus org itself.

## Side by side

| | prom-mcp | pab1it0/prometheus-mcp-server | prometheus/prometheus-mcp |
| --- | --- | --- | --- |
| Focus | Alert triage | General querying and metric discovery | Broad Prometheus API coverage |
| Tools | 5: `prom_query`, `prom_query_range`, `prom_alerts_explain`, `prom_label_values`, `prom_series_discover` | 6 in its README table (`execute_query`, `execute_range_query`, `list_metrics`, `get_metric_metadata`, `get_targets`, `health_check`); the enabled set is configurable | 27 in its full tool list, plus 3 TSDB admin tools behind a flag |
| Firing alerts | One call returns active alerts joined with rule expression, group, `for`, labels, and annotations | No alert tool in its README tool list | `list_alerts` and `list_rules` are separate tools |
| Discovery before query | `prom_series_discover` and `prom_label_values` are core tools | `list_metrics`, `get_metric_metadata`, `get_targets` | `series`, `label_names`, `label_values`, `metric_metadata` |
| Output shape | Compact, triage-shaped text | Prometheus API JSON (with an option to drop UI links to save tokens) | Prometheus API JSON, optional TOON encoding and truncation |
| Runtime | Single Go binary, zero dependencies | Python, usually run via Docker | Go binary, also Docker, Helm, and system packages |
| Transports | stdio | stdio, HTTP, SSE | stdio, Streamable HTTP |
| Install routes | MCPB bundles for Claude Desktop, official MCP Registry, release binaries, `go install` | Docker/GHCR, Helm chart, Docker Desktop MCP Catalog | Release binaries, Docker/GHCR, Helm chart, system packages |
| Auth | Bearer token | Basic auth, bearer, mTLS client certs, custom headers, multi-tenant org ID | Prometheus HTTP config files; forwards client credentials over HTTP |
| Extras | 2 resources (`prometheus://alerts`, `prometheus://config`), 1 `triage` prompt | Per-deployment tool selection via `TOOL_PREFIX` and config | Embedded runbooks (Agent Skills) exposed as tools, resources, and prompts; Prometheus docs search; Thanos backend |

## Which one should you pick?

Pick **prom-mcp** when the agent's job is "something is firing, figure out
why": the alert explanation is the center of the design, the tool surface
is small enough to stay out of the model's way, and it runs as one static
binary next to any Prometheus, with no container or runtime to manage.

Pick **pab1it0/prometheus-mcp-server** when you want a query server you can
deploy like infrastructure: container image, Helm chart, several auth
options including mTLS, and HTTP/SSE transports for shared access.

Pick **prometheus/prometheus-mcp** when you want the whole Prometheus API
available to the model: config and flags inspection, TSDB stats, exemplar
queries, guided runbooks, and docs search, maintained under the Prometheus
org.

prom-mcp deliberately does less. It will not inspect your Prometheus
config or walk you through a runbook. What it does instead is keep the
agent honest during an incident: discover what exists, explain what is
firing with the rule attached, then check the trend, in output shaped for
triage rather than raw API dumps.
