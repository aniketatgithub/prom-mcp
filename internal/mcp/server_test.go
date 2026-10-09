package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func rpc(t *testing.T, method string, id int, params string) response {
	t.Helper()
	var req request
	req.JSONRPC = "2.0"
	req.Method = method
	idJSON, err := json.Marshal(id)
	if err != nil {
		t.Fatal(err)
	}
	req.ID = idJSON
	if params != "" {
		req.Params = json.RawMessage(params)
	}
	resp, reply := handle(req, nil)
	if !reply {
		t.Fatalf("expected reply for %s", method)
	}
	return resp
}

func TestInitialize(t *testing.T) {
	resp := rpc(t, "initialize", 1, `{}`)
	if resp.Error != nil {
		t.Fatalf("initialize error: %v", resp.Error.Message)
	}
	b, _ := json.Marshal(resp.Result)
	if !strings.Contains(string(b), "prom-mcp") || !strings.Contains(string(b), "2024-11-05") {
		t.Fatalf("unexpected initialize result: %s", b)
	}
}

func TestToolsList(t *testing.T) {
	resp := rpc(t, "tools/list", 2, `{}`)
	b, _ := json.Marshal(resp.Result)
	for _, name := range []string{"prom_query", "prom_alerts_explain", "prom_series_discover"} {
		if !strings.Contains(string(b), name) {
			t.Fatalf("tools/list missing %s: %s", name, b)
		}
	}
}

func TestUnknownMethod(t *testing.T) {
	resp := rpc(t, "nope/method", 3, `{}`)
	if resp.Error == nil || resp.Error.Code != -32601 {
		t.Fatalf("expected -32601, got %+v", resp.Error)
	}
}

func TestPing(t *testing.T) {
	resp := rpc(t, "ping", 4, `{}`)
	if resp.Error != nil {
		t.Fatalf("ping error: %s", resp.Error.Message)
	}
}

func TestServeHandshake(t *testing.T) {
	in := strings.NewReader("{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"initialize\"}\n{\"jsonrpc\":\"2.0\",\"method\":\"notifications/initialized\"}\n{\"jsonrpc\":\"2.0\",\"id\":2,\"method\":\"tools/list\"}\n")
	var out strings.Builder
	if err := Serve(in, &out, nil); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 replies (notifications produce none), got %d: %s", len(lines), out.String())
	}
}
