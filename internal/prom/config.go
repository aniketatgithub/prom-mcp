package prom

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config controls how prom-mcp reaches Prometheus. Environment variables
// override the config file, which overrides defaults.
//
// Env: PROM_URL, PROM_TOKEN, PROM_TIMEOUT (e.g. 30s), PROM_CONFIG (path).
// File (JSON): {"base_url": "...", "token": "...", "timeout_seconds": 30}
// Default file locations: $PROM_CONFIG, ./prom-mcp.json, ~/.prom-mcp.json.
type Config struct {
	BaseURL        string `json:"base_url"`
	Token          string `json:"token"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// LoadConfig resolves configuration from env and the first config file found.
func LoadConfig() Config {
	var cfg Config
	for _, path := range candidatePaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if json.Unmarshal(data, &cfg) == nil {
			break
		}
	}
	if v := os.Getenv("PROM_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("PROM_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("PROM_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.TimeoutSeconds = int(d.Seconds())
		}
	}
	return cfg
}

func candidatePaths() []string {
	paths := []string{}
	if v := os.Getenv("PROM_CONFIG"); v != "" {
		paths = append(paths, v)
	}
	paths = append(paths, "prom-mcp.json")
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".prom-mcp.json"))
	}
	return paths
}

// Client builds a Client from this config.
func (cfg Config) Client() *Client {
	c := New(cfg.BaseURL)
	if cfg.Token != "" {
		c.Token = cfg.Token
	}
	if cfg.TimeoutSeconds > 0 {
		c.HTTP.Timeout = time.Duration(cfg.TimeoutSeconds) * time.Second
	}
	return c
}
