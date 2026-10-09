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
		"/api/v1/query":  "query_up.json",
		"/api/v1/alerts": "alerts.json",
		"/api/v1/rules":  "rules.json",
		"/api/v1/series": "series.json",
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
