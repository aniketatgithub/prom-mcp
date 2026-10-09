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

const (
	defaultTimeout  = 15 * time.Second
	maxResponseBody = 16 << 20 // 16 MiB cap on Prometheus responses
	maxAttempts     = 3
)

// Client talks to one Prometheus server.
type Client struct {
	Base  string
	HTTP  *http.Client
	Token string // optional bearer token, sent as Authorization header
}

// New builds a Client. Empty base defaults to http://localhost:9090.
func New(base string) *Client {
	if base == "" {
		base = "http://localhost:9090"
	}
	return &Client{Base: strings.TrimRight(base, "/"), HTTP: &http.Client{Timeout: defaultTimeout}}
}

type apiResponse struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
	Error     string          `json:"error"`
	ErrorType string          `json:"errorType"`
}

// get performs one API call with bounded retries on transient failures
// (network errors and HTTP 5xx). Client errors (4xx) fail immediately.
func (c *Client) get(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		data, retryable, err := c.getOnce(ctx, path, params)
		if err == nil {
			return data, nil
		}
		lastErr = err
		if !retryable || attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt) * 200 * time.Millisecond):
		}
	}
	return nil, lastErr
}

func (c *Client) getOnce(ctx context.Context, path string, params url.Values) (json.RawMessage, bool, error) {
	u := c.Base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, false, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, false, ctx.Err()
		}
		return nil, true, fmt.Errorf("prometheus unreachable at %s: %w", c.Base, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, true, fmt.Errorf("reading response from %s: %w", u, err)
	}
	if resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("prometheus at %s returned HTTP %d", c.Base, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("prometheus at %s returned HTTP %d: %s", c.Base, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var ar apiResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, false, fmt.Errorf("bad response from %s (HTTP %d)", u, resp.StatusCode)
	}
	if ar.Status != "success" {
		msg := ar.Error
		if ar.ErrorType != "" {
			msg = ar.ErrorType + ": " + msg
		}
		return nil, false, fmt.Errorf("prometheus error: %s", msg)
	}
	return ar.Data, false, nil
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

// QueryRange runs a range PromQL query. Step defaults to a sensible value
// for the window when empty.
func (c *Client) QueryRange(ctx context.Context, expr string, minutes int, step string) (string, error) {
	if minutes <= 0 {
		minutes = 60
	}
	end := time.Now()
	start := end.Add(-time.Duration(minutes) * time.Minute)
	if step == "" {
		step = "60s"
		if minutes > 360 {
			step = "300s"
		}
	}
	p := url.Values{
		"query": {expr},
		"start": {start.Format(time.RFC3339)},
		"end":   {end.Format(time.RFC3339)},
		"step":  {step},
	}
	data, err := c.get(ctx, "/api/v1/query_range", p)
	if err != nil {
		return "", err
	}
	var payload struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][2]any          `json:"values"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "range query: %s (last %dm, step %s)\n%d series\n", expr, minutes, step, len(payload.Result))
	for _, s := range payload.Result {
		var first, last any
		var min, max float64
		for i, v := range s.Values {
			if i == 0 {
				first = v[1]
			}
			last = v[1]
			var fv float64
			fmt.Sscanf(fmt.Sprint(v[1]), "%g", &fv)
			if i == 0 || fv < min {
				min = fv
			}
			if i == 0 || fv > max {
				max = fv
			}
		}
		fmt.Fprintf(&b, "%s points=%d first=%v last=%v min=%g max=%g\n", formatLabels(s.Metric), len(s.Values), first, last, min, max)
	}
	return b.String(), nil
}

// LabelValues lists values of one label, optionally scoped by a series matcher.
func (c *Client) LabelValues(ctx context.Context, label, match string) (string, error) {
	p := url.Values{}
	if match != "" {
		p.Set("match[]", match)
	}
	data, err := c.get(ctx, "/api/v1/label/"+url.PathEscape(label)+"/values", p)
	if err != nil {
		return "", err
	}
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return "", err
	}
	scope := ""
	if match != "" {
		scope = " for " + match
	}
	return fmt.Sprintf("label %s%s: %d value(s)\n%s\n", label, scope, len(values), strings.Join(values, ", ")), nil
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
	const cap = 100
	for i, s := range series {
		if i >= cap {
			fmt.Fprintf(&b, "... and %d more (refine the matcher)\n", len(series)-cap)
			break
		}
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
