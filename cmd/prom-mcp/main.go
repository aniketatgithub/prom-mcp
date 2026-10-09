// Command prom-mcp is an MCP server that gives AI agents production-grade
// access to Prometheus: instant queries, alert explanations, series discovery.
//
// Usage:
//
//	prom-mcp                 # MCP server over stdio (PROM_URL env, default http://localhost:9090)
//	prom-mcp demo            # self-test against built-in fixtures, no server needed
//	prom-mcp query 'up'      # one-shot CLI query
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/aniketatgithub/prom-mcp/internal/mcp"
	"github.com/aniketatgithub/prom-mcp/internal/prom"
)

func main() {
	base := os.Getenv("PROM_URL")
	client := prom.New(base)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "demo":
			if err := demo(); err != nil {
				fmt.Fprintln(os.Stderr, "demo error:", err)
				os.Exit(1)
			}
			return
		case "query":
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "usage: prom-mcp query '<promql>'")
				os.Exit(2)
			}
			out, err := client.Query(context.Background(), os.Args[2])
			fmt.Print(out)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			return
		}
	}

	if err := mcp.Serve(os.Stdin, os.Stdout, client); err != nil {
		fmt.Fprintln(os.Stderr, "mcp error:", err)
		os.Exit(1)
	}
}

// demo spins up a fake Prometheus serving testdata fixtures and runs each
// tool once, so the whole stack is verifiable with no live server.
func demo() error {
	mux := http.NewServeMux()
	serve := func(file string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			data, err := os.ReadFile(filepath.Join("testdata", file))
			if err != nil {
				// Installed-binary fallback: fixtures next to source only;
				// in that case report and skip gracefully.
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
		}
	}
	mux.HandleFunc("/api/v1/query", serve("query_up.json"))
	mux.HandleFunc("/api/v1/alerts", serve("alerts.json"))
	mux.HandleFunc("/api/v1/rules", serve("rules.json"))
	mux.HandleFunc("/api/v1/series", serve("series.json"))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := prom.New(srv.URL)
	ctx := context.Background()
	calls := []struct {
		name string
		args map[string]any
	}{
		{"prom_query", map[string]any{"query": "up"}},
		{"prom_alerts_explain", map[string]any{}},
		{"prom_series_discover", map[string]any{"match": "up"}},
	}
	fmt.Println("prom-mcp demo against fixture Prometheus at", srv.URL)
	for _, call := range calls {
		out, err := mcp.Call(ctx, c, call.name, call.args)
		if err != nil {
			return err
		}
		fmt.Printf("\n== %s ==\n%s", call.name, out)
	}
	fmt.Println("\ndemo OK: all 3 tools answered from fixtures")
	return nil
}
