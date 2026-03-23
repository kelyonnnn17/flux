package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	// Example placeholder field
	LogLevel string `yaml:"log_level"`
}

// Load reads the config file from the user's config directory (~/.config/flux/config.yaml).
// If the file does not exist, it returns an empty Config without error.
func Load() (*Config, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	cfgPath := filepath.Join(cfgDir, "flux", "config.yaml")
	data, err := os.ReadFile(cfgPath)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
