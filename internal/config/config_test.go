package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()

	configPath := filepath.Join(dir, "config.yaml")

	configData := `
server:
  port: 8080

health:
  interval_seconds: 5

backends:
  - name: Backend A
    url: http://localhost:9001

  - name: Backend B
    url: http://localhost:9002
`

	err := os.WriteFile(
		configPath,
		[]byte(configData),
		0644,
	)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Server.Port != 8080 {
		t.Fatalf(
			"expected port 8080, got %d",
			cfg.Server.Port,
		)
	}

	if cfg.Health.IntervalSeconds != 5 {
		t.Fatalf(
			"expected health interval 5, got %d",
			cfg.Health.IntervalSeconds,
		)
	}

	if len(cfg.Backends) != 2 {
		t.Fatalf(
			"expected 2 backends, got %d",
			len(cfg.Backends),
		)
	}

	if cfg.Backends[0].Name != "Backend A" {
		t.Fatalf(
			"expected Backend A, got %s",
			cfg.Backends[0].Name,
		)
	}

	if cfg.Backends[1].Name != "Backend B" {
		t.Fatalf(
			"expected Backend B, got %s",
			cfg.Backends[1].Name,
		)
	}
}
