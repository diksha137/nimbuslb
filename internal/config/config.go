package config

import (
	"fmt"
	"net/url"
	"strings"
)

type Config struct {
	Server   ServerConfig    `yaml:"server"`
	Health   HealthConfig    `yaml:"health"`
	Backends []BackendConfig `yaml:"backends"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type HealthConfig struct {
	IntervalSeconds int `yaml:"interval_seconds"`
}

type BackendConfig struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf(
			"server port must be between 1 and 65535, got %d",
			c.Server.Port,
		)
	}

	if c.Health.IntervalSeconds <= 0 {
		return fmt.Errorf(
			"health interval must be greater than 0",
		)
	}

	if len(c.Backends) == 0 {
		return fmt.Errorf(
			"at least one backend must be configured",
		)
	}

	seenNames := make(map[string]bool)

	for i, backend := range c.Backends {
		name := strings.TrimSpace(backend.Name)

		if name == "" {
			return fmt.Errorf(
				"backend %d has an empty name",
				i,
			)
		}

		if seenNames[name] {
			return fmt.Errorf(
				"duplicate backend name: %s",
				name,
			)
		}

		seenNames[name] = true

		parsedURL, err := url.Parse(backend.URL)
		if err != nil {
			return fmt.Errorf(
				"backend %s has invalid URL: %w",
				name,
				err,
			)
		}

		if parsedURL.Scheme != "http" &&
			parsedURL.Scheme != "https" {
			return fmt.Errorf(
				"backend %s must use http or https",
				name,
			)
		}

		if parsedURL.Host == "" {
			return fmt.Errorf(
				"backend %s has no host",
				name,
			)
		}
	}

	return nil
}
