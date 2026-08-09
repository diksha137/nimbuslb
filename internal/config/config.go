package config

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
