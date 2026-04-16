package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultFileName = "iconkit.json"

type Config struct {
	Input   string  `json:"input,omitempty"`
	Sizes   []int   `json:"sizes,omitempty"`
	Radius  int     `json:"radius,omitempty"`
	Preset  string  `json:"preset,omitempty"`
	Out     string  `json:"out,omitempty"`
	Force   bool    `json:"force,omitempty"`
	Padding float64 `json:"pad,omitempty"`
	Bg      string  `json:"bg,omitempty"`
	Ico     bool    `json:"ico,omitempty"`
}

// Load reads an iconkit.json config from the given directory.
// Returns nil (no error) if the file does not exist.
func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, DefaultFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}
	return &cfg, nil
}

// LoadFromFile reads config from a specific file path.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}
	return &cfg, nil
}
