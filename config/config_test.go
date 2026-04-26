package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FelineStateMachine/puzzletea/puzzle"
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

func TestSaveAndLoadExportConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		Theme: "Dracula",
		Export: ExportConfig{
			Title:           "Sampler",
			Volume:          2,
			SheetLayout:     "duplex-booklet",
			PDFOutputPath:   filepath.Join(dir, "sample.pdf"),
			JSONLEnabled:    true,
			JSONLOutputPath: filepath.Join(dir, "sample.jsonl"),
			Counts: map[puzzle.GameID]map[puzzle.ModeID]int{
				puzzle.CanonicalGameID("Sudoku"): {
					puzzle.CanonicalModeID("Easy"): 2,
				},
			},
		},
	}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Export.Title != cfg.Export.Title {
		t.Fatalf("loaded export title = %q, want %q", loaded.Export.Title, cfg.Export.Title)
	}
	if !loaded.Export.JSONLEnabled {
		t.Fatal("expected jsonl toggle to round-trip")
	}
	if got := loaded.Export.Counts[puzzle.CanonicalGameID("Sudoku")][puzzle.CanonicalModeID("Easy")]; got != 2 {
		t.Fatalf("loaded export count = %d, want 2", got)
	}
}

func TestSaveAndLoadCreateConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &Config{
		Create: CreateConfig{
			SelectedLeafIDs: []string{"nonogram/5x5", "takuzu/takuzu/6x6"},
			Elo:             1400,
		},
	}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Create.Elo != 1400 {
		t.Fatalf("loaded create Elo = %d, want 1400", loaded.Create.Elo)
	}
	if len(loaded.Create.SelectedLeafIDs) != 2 || loaded.Create.SelectedLeafIDs[1] != "takuzu/takuzu/6x6" {
		t.Fatalf("loaded create leaves = %#v", loaded.Create.SelectedLeafIDs)
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
