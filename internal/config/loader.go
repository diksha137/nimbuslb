package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Server.Port == 0 {
		return nil, fmt.Errorf("server port must be configured")
	}

	if len(cfg.Backends) == 0 {
		return nil, fmt.Errorf("at least one backend must be configured")
	}

	return &cfg, nil
}
