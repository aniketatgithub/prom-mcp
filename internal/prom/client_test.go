package prom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for path, file := range map[string]string{
		"/api/v1/query":            "query_up.json",
		"/api/v1/query_range":      "query_range.json",
		"/api/v1/alerts":           "alerts.json",
		"/api/v1/rules":            "rules.json",
		"/api/v1/series":           "series.json",
		"/api/v1/label/job/values": "label_job_values.json",
	} {
		file := file
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			data, err := os.ReadFile(filepath.Join("..", "..", "testdata", file))
			if err != nil {
				t.Fatal(err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		})
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestQuery(t *testing.T) {
	c := New(fixtureServer(t).URL)
	out, err := c.Query(context.Background(), "up")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2 series") || !strings.Contains(out, `instance="node1:9100"`) {
		t.Fatalf("unexpected query output:\n%s", out)
	}
}

func TestAlertsExplain(t *testing.T) {
	c := New(fixtureServer(t).URL)
	out, err := c.AlertsExplain(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "NodeDown") || !strings.Contains(out, `up{job="node"} == 0`) {
		t.Fatalf("explanation missing rule join:\n%s", out)
	}
}

func TestQueryRange(t *testing.T) {
	c := New(fixtureServer(t).URL)
	out, err := c.QueryRange(context.Background(), "up", 60, "60s")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "1 series") || !strings.Contains(out, "points=3") || !strings.Contains(out, "min=0 max=1") {
		t.Fatalf("unexpected range output:\n%s", out)
	}
}

func TestLabelValues(t *testing.T) {
	c := New(fixtureServer(t).URL)
	out, err := c.LabelValues(context.Background(), "job", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2 value(s)") || !strings.Contains(out, "node") {
		t.Fatalf("unexpected label output:\n%s", out)
	}
}

func TestRetryOn500(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		data, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "query_up.json"))
		w.Write(data)
	}))
	defer srv.Close()
	c := New(srv.URL)
	if _, err := c.Query(context.Background(), "up"); err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestNoRetryOn400(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad query"))
	}))
	defer srv.Close()
	c := New(srv.URL)
	if _, err := c.Query(context.Background(), "up{"); err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("4xx must not retry, got %d attempts", attempts)
	}
}

func TestLoadConfigEnvWins(t *testing.T) {
	t.Setenv("PROM_URL", "http://env.example:9999")
	t.Setenv("PROM_TOKEN", "sekret")
	cfg := LoadConfig()
	if cfg.BaseURL != "http://env.example:9999" || cfg.Token != "sekret" {
		t.Fatalf("env config not applied: %+v", cfg)
	}
	cl := cfg.Client()
	if cl.Token != "sekret" || cl.Base != "http://env.example:9999" {
		t.Fatalf("client not built from config: %+v", cl)
	}
}

func TestSeries(t *testing.T) {
	c := New(fixtureServer(t).URL)
	out, err := c.Series(context.Background(), "up")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "2 series match") {
		t.Fatalf("unexpected series output:\n%s", out)
	}
}
