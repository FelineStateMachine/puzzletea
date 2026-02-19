// Package config provides persistent application configuration for PuzzleTea.
// Settings are stored as JSON in ~/.puzzletea/config.json alongside the game
// history database.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all user-configurable settings.
type Config struct {
	Theme string `json:"theme,omitempty"` // theme name; empty = default
}

// Default returns a Config with all settings at their zero values (the
// built-in defaults).
func Default() *Config {
	return &Config{}
}

// DefaultPath returns ~/.puzzletea/config.json.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".puzzletea", "config.json")
}

// Load reads and parses a config file. If the file does not exist, a default
// Config is returned with no error. Parse errors are returned so callers can
// log a warning, but the application should still start with defaults.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return Default(), fmt.Errorf("reading config: %w", err)
	}

	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return Default(), fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// Save writes the config to disk, creating the parent directory if needed.
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}
