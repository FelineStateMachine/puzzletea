package config

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Load / Save round-trip (P0) ---

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/tmp/puzzletea-test-nonexistent/config.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg.Theme != "" {
		t.Errorf("expected empty theme for default config, got %q", cfg.Theme)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{Theme: "Dracula"}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Theme != "Dracula" {
		t.Errorf("expected theme Dracula, got %q", loaded.Theme)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "config.json")

	cfg := &Config{Theme: "Monokai Pro"}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to exist: %v", err)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	// Should still return a usable default.
	if cfg.Theme != "" {
		t.Errorf("expected default theme on error, got %q", cfg.Theme)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.Theme != "" {
		t.Errorf("expected empty theme, got %q", cfg.Theme)
	}
}
