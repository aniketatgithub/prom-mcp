// Package prom is a small Prometheus HTTP API client.
package prom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Client talks to one Prometheus server.
type Client struct {
	Base string
	HTTP *http.Client
}

// New builds a Client. Empty base defaults to http://localhost:9090.
func New(base string) *Client {
	if base == "" {
		base = "http://localhost:9090"
	}
	return &Client{Base: strings.TrimRight(base, "/"), HTTP: &http.Client{Timeout: 10 * time.Second}}
}

type apiResponse struct {
	Status string          `json:"status"`
	Data   json.RawMessage `json:"data"`
	Error  string          `json:"error"`
}

func (c *Client) get(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	u := c.Base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("prometheus unreachable at %s: %w", c.Base, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ar apiResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, fmt.Errorf("bad response from %s (HTTP %d)", u, resp.StatusCode)
	}
	if ar.Status != "success" {
		return nil, fmt.Errorf("prometheus error: %s", ar.Error)
	}
	return ar.Data, nil
}

// Sample is one instant-query result.
type Sample struct {
	Metric map[string]string `json:"metric"`
	Value  [2]any            `json:"value"`
}

// Query runs an instant PromQL query and renders a compact text summary.
func (c *Client) Query(ctx context.Context, expr string) (string, error) {
	p := url.Values{"query": {expr}}
	data, err := c.get(ctx, "/api/v1/query", p)
	if err != nil {
		return "", err
	}
	var payload struct {
		ResultType string   `json:"resultType"`
		Result     []Sample `json:"result"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "query: %s\nresultType: %s, %d series\n", expr, payload.ResultType, len(payload.Result))
	for _, s := range payload.Result {
		fmt.Fprintf(&b, "%s => %v\n", formatLabels(s.Metric), s.Value[1])
	}
	return b.String(), nil
}

// Alert is one active alert.
type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    string            `json:"activeAt"`
	Value       string            `json:"value"`
}

type ruleGroup struct {
	Rules []struct {
		Name        string            `json:"name"`
		Query       string            `json:"query"`
		Type        string            `json:"type"`
		Annotations map[string]string `json:"annotations"`
	} `json:"rules"`
}

// AlertsExplain fetches active alerts and joins them with rule definitions,
// producing a human-readable explanation an agent (or human) can act on.
func (c *Client) AlertsExplain(ctx context.Context) (string, error) {
	data, err := c.get(ctx, "/api/v1/alerts", nil)
	if err != nil {
		return "", err
	}
	var payload struct {
		Alerts []Alert `json:"alerts"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	rules := map[string]string{}
	if rdata, err := c.get(ctx, "/api/v1/rules", nil); err == nil {
		var rp struct {
			Groups []ruleGroup `json:"groups"`
		}
		if json.Unmarshal(rdata, &rp) == nil {
			for _, g := range rp.Groups {
				for _, r := range g.Rules {
					if r.Type == "alerting" {
						rules[r.Name] = r.Query
					}
				}
			}
		}
	}
	if len(payload.Alerts) == 0 {
		return "no active alerts", nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d active alert(s)\n", len(payload.Alerts))
	for _, a := range payload.Alerts {
		name := a.Labels["alertname"]
		fmt.Fprintf(&b, "\n[%s] %s (since %s, value %s)\n", a.State, name, a.ActiveAt, a.Value)
		fmt.Fprintf(&b, "  labels: %s\n", formatLabels(a.Labels))
		if s := a.Annotations["summary"]; s != "" {
			fmt.Fprintf(&b, "  summary: %s\n", s)
		}
		if d := a.Annotations["description"]; d != "" {
			fmt.Fprintf(&b, "  description: %s\n", d)
		}
		if q, ok := rules[name]; ok {
			fmt.Fprintf(&b, "  rule: %s\n", q)
		}
	}
	return b.String(), nil
}

// Series discovers series matching a label matcher (e.g. up{job="node"}).
func (c *Client) Series(ctx context.Context, match string) (string, error) {
	p := url.Values{"match[]": {match}}
	data, err := c.get(ctx, "/api/v1/series", p)
	if err != nil {
		return "", err
	}
	var series []map[string]string
	if err := json.Unmarshal(data, &series); err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%d series match %s\n", len(series), match)
	for _, s := range series {
		fmt.Fprintf(&b, "%s\n", formatLabels(s))
	}
	return b.String(), nil
}

func formatLabels(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%s=%q", k, m[k])
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
