// Package mcp implements a minimal Model Context Protocol server over
// stdio (newline-delimited JSON-RPC 2.0) exposing Prometheus tools.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aniketatgithub/prom-mcp/internal/prom"
)

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var tools = []map[string]any{
	{
		"name":        "prom_query",
		"description": "Run an instant PromQL query against Prometheus and return the series and values.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{
			"query": map[string]any{"type": "string", "description": "PromQL expression, e.g. up or rate(http_requests_total[5m])"},
		}, "required": []string{"query"}},
	},
	{
		"name":        "prom_alerts_explain",
		"description": "List active Prometheus alerts joined with their rule expressions and annotations, explained for triage.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
	},
	{
		"name":        "prom_series_discover",
		"description": "Discover which series exist for a label matcher, e.g. up{job=\"node\"}, before writing queries against them.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{
			"match": map[string]any{"type": "string", "description": "Series selector, e.g. up or http_requests_total{job=\"api\"}"},
		}, "required": []string{"match"}},
	},
}

// Serve runs the MCP server until EOF on in.
func Serve(in io.Reader, out io.Writer, client *prom.Client) error {
	sc := bufio.NewScanner(in)
	sc.Buffer(make([]byte, 0, 1024*1024), 64*1024*1024)
	enc := json.NewEncoder(out)
	for sc.Scan() {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var req request
		if json.Unmarshal(sc.Bytes(), &req) != nil {
			continue
		}
		if resp, reply := handle(req, client); reply {
			if err := enc.Encode(resp); err != nil {
				return err
			}
		}
	}
	return sc.Err()
}

func handle(req request, client *prom.Client) (response, bool) {
	base := response{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		base.Result = map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "prom-mcp", "version": "0.1.0"},
		}
		return base, true
	case "notifications/initialized", "notifications/cancelled":
		return base, false
	case "ping":
		base.Result = map[string]any{}
		return base, true
	case "tools/list":
		base.Result = map[string]any{"tools": tools}
		return base, true
	case "tools/call":
		var p struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			base.Error = &rpcError{Code: -32602, Message: "invalid params"}
			return base, true
		}
		text, err := Call(context.Background(), client, p.Name, p.Arguments)
		if err != nil {
			base.Error = &rpcError{Code: -32602, Message: err.Error()}
			return base, true
		}
		base.Result = map[string]any{"content": []map[string]any{{"type": "text", "text": text}}}
		return base, true
	default:
		base.Error = &rpcError{Code: -32601, Message: "method not found: " + req.Method}
		return base, true
	}
}

// Call dispatches one tool invocation. Exported for the demo mode and tests.
func Call(ctx context.Context, client *prom.Client, name string, args map[string]any) (string, error) {
	switch name {
	case "prom_query":
		q, _ := args["query"].(string)
		if q == "" {
			return "", fmt.Errorf("query is required")
		}
		return client.Query(ctx, q)
	case "prom_alerts_explain":
		return client.AlertsExplain(ctx)
	case "prom_series_discover":
		m, _ := args["match"].(string)
		if m == "" {
			return "", fmt.Errorf("match is required")
		}
		return client.Series(ctx, m)
	default:
		return "", fmt.Errorf("unknown tool %q", name)
	}
}
