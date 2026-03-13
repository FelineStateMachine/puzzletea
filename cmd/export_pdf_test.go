package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/pdfexport"

	"github.com/spf13/cobra"
)

func TestExportPDFVolumeFlagDefault(t *testing.T) {
	flag := exportPDFCmd.Flags().Lookup("volume")
	if flag == nil {
		t.Fatal("expected --volume flag")
	}
	if flag.DefValue != "1" {
		t.Fatalf("default volume = %q, want %q", flag.DefValue, "1")
	}
}

func TestValidatePDFVolume(t *testing.T) {
	tests := []struct {
		name   string
		volume int
		wantOK bool
	}{
		{name: "default volume", volume: 1, wantOK: true},
		{name: "positive volume", volume: 7, wantOK: true},
		{name: "zero volume rejected", volume: 0, wantOK: false},
		{name: "negative volume rejected", volume: -3, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePDFVolume(tt.volume)
			if tt.wantOK && err != nil {
				t.Fatalf("validatePDFVolume(%d) error = %v, want nil", tt.volume, err)
			}
			if !tt.wantOK {
				if err == nil {
					t.Fatalf("validatePDFVolume(%d) expected error", tt.volume)
				}
				if !strings.Contains(err.Error(), "--volume") {
					t.Fatalf("validatePDFVolume(%d) error = %q, want mention of --volume", tt.volume, err.Error())
				}
			}
		})
	}
}

func TestBuildRenderConfigForPDFUsesTitleAsCoverSubtitle(t *testing.T) {
	reset := snapshotExportPDFFlags()
	defer reset()

	flagPDFTitle = "Catacombs & Pines"
	flagPDFHeader = "Custom heading paragraph"
	flagPDFVolume = 7
	flagPDFAdvert = "Custom advert"
	flagPDFCoverColor = ""

	now := time.Date(2026, 2, 22, 11, 0, 0, 0, time.UTC)
	docs := []pdfexport.PackDocument{{Metadata: pdfexport.PackMetadata{Category: "Nonogram"}}}
	cfg, err := buildRenderConfigForPDF(docs, "seed-1", now)
	if err != nil {
		t.Fatalf("buildRenderConfigForPDF error = %v", err)
	}
	if cfg.CoverSubtitle != "Catacombs & Pines" {
		t.Fatalf("CoverSubtitle = %q, want %q", cfg.CoverSubtitle, "Catacombs & Pines")
	}
	if cfg.VolumeNumber != 7 {
		t.Fatalf("VolumeNumber = %d, want %d", cfg.VolumeNumber, 7)
	}
	if cfg.HeaderText != "Custom heading paragraph" {
		t.Fatalf("HeaderText = %q, want %q", cfg.HeaderText, "Custom heading paragraph")
	}
	if cfg.AdvertText != "Custom advert" {
		t.Fatalf("AdvertText = %q, want %q", cfg.AdvertText, "Custom advert")
	}
	if !cfg.GeneratedAt.Equal(now) {
		t.Fatalf("GeneratedAt = %v, want %v", cfg.GeneratedAt, now)
	}
}

func TestBuildRenderConfigForPDFDefaultsSubtitleFromDocs(t *testing.T) {
	reset := snapshotExportPDFFlags()
	defer reset()

	flagPDFTitle = ""
	flagPDFVolume = 1
	flagPDFAdvert = "Find more puzzles"
	flagPDFCoverColor = ""

	docs := []pdfexport.PackDocument{{Metadata: pdfexport.PackMetadata{Category: "Sudoku"}}}
	cfg, err := buildRenderConfigForPDF(docs, "seed-2", time.Now())
	if err != nil {
		t.Fatalf("buildRenderConfigForPDF error = %v", err)
	}
	if cfg.CoverSubtitle != "Sudoku Puzzle Pack" {
		t.Fatalf("CoverSubtitle = %q, want %q", cfg.CoverSubtitle, "Sudoku Puzzle Pack")
	}
}

func TestBuildRenderConfigForPDFCoverColorControlsCoverPages(t *testing.T) {
	reset := snapshotExportPDFFlags()
	defer reset()

	flagPDFTitle = "Issue 01"
	flagPDFVolume = 1
	flagPDFAdvert = "Find more puzzles"
	docs := []pdfexport.PackDocument{{Metadata: pdfexport.PackMetadata{Category: "Sudoku"}}}

	flagPDFCoverColor = ""
	cfgNoCover, err := buildRenderConfigForPDF(docs, "seed-3", time.Now())
	if err != nil {
		t.Fatalf("buildRenderConfigForPDF (no cover color) error = %v", err)
	}
	if cfgNoCover.CoverColor != nil {
		t.Fatalf("CoverColor = %+v, want nil when --cover-color is omitted", cfgNoCover.CoverColor)
	}

	flagPDFCoverColor = "#112233"
	cfgWithCover, err := buildRenderConfigForPDF(docs, "seed-4", time.Now())
	if err != nil {
		t.Fatalf("buildRenderConfigForPDF (with cover color) error = %v", err)
	}
	if cfgWithCover.CoverColor == nil {
		t.Fatal("CoverColor = nil, want parsed color when --cover-color is set")
	}
	if *cfgWithCover.CoverColor != (pdfexport.RGB{R: 0x11, G: 0x22, B: 0x33}) {
		t.Fatalf("CoverColor = %+v, want {R:17 G:34 B:51}", *cfgWithCover.CoverColor)
	}
}

func TestRunExportPDFRejectsInputsWhenAllPuzzlesUnsupported(t *testing.T) {
	reset := snapshotExportPDFFlags()
	defer reset()

	dir := t.TempDir()
	input := filepath.Join(dir, "lights.jsonl")
	output := filepath.Join(dir, "lights.pdf")

	record := pdfexport.JSONLRecord{
		Schema: pdfexport.ExportSchemaV1,
		Pack: pdfexport.JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Lights Out",
			ModeSelection: "Standard",
			Count:         1,
		},
		Puzzle: pdfexport.JSONLPuzzle{
			Index: 1,
			Name:  "glow-shore",
			Game:  "Lights Out",
			Mode:  "Standard",
			Save:  json.RawMessage(`{"size":5}`),
		},
	}
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	flagPDFOutput = output
	flagPDFVolume = 1

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)

	err = runExportPDF(cmd, []string{input})
	if err == nil {
		t.Fatal("expected unsupported export error")
	}
	if !strings.Contains(err.Error(), "no printable puzzles found") {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != "" {
		t.Fatalf("expected no stdout output, got %q", out.String())
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("expected no output file, stat err = %v", err)
	}
}

func TestRunExportPDFSkipsUnsupportedRecordsWhenMixed(t *testing.T) {
	reset := snapshotExportPDFFlags()
	defer reset()

	dir := t.TempDir()
	input := filepath.Join(dir, "mixed.jsonl")
	output := filepath.Join(dir, "mixed.pdf")

	records := []pdfexport.JSONLRecord{
		{
			Schema: pdfexport.ExportSchemaV1,
			Pack: pdfexport.JSONLPackMeta{
				Generated:     "2026-02-22T10:00:00Z",
				Version:       "v-test",
				Category:      "Lights Out",
				ModeSelection: "Standard",
				Count:         2,
			},
			Puzzle: pdfexport.JSONLPuzzle{
				Index: 1,
				Name:  "glow-shore",
				Game:  "Lights Out",
				Mode:  "Standard",
				Save:  json.RawMessage(`{"size":5}`),
			},
		},
		{
			Schema: pdfexport.ExportSchemaV1,
			Pack: pdfexport.JSONLPackMeta{
				Generated:     "2026-02-22T10:00:00Z",
				Version:       "v-test",
				Category:      "Sudoku",
				ModeSelection: "Easy",
				Count:         2,
			},
			Puzzle: pdfexport.JSONLPuzzle{
				Index: 2,
				Name:  "moss-pine",
				Game:  "Sudoku",
				Mode:  "Easy",
				Save:  json.RawMessage(`{"provided":[{"x":0,"y":0,"v":5}]}`),
			},
		},
	}
	var lines []byte
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line...)
		lines = append(lines, '\n')
	}
	if err := os.WriteFile(input, lines, 0o644); err != nil {
		t.Fatal(err)
	}

	flagPDFOutput = output
	flagPDFVolume = 1

	cmd := &cobra.Command{}
	cmd.SetOut(&bytes.Buffer{})
	if err := runExportPDF(cmd, []string{input}); err != nil {
		t.Fatalf("expected mixed export success, got %v", err)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatalf("expected output file, got stat error: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty output PDF")
	}
}

func snapshotExportPDFFlags() func() {
	oldTitle := flagPDFTitle
	oldHeader := flagPDFHeader
	oldVolume := flagPDFVolume
	oldAdvert := flagPDFAdvert
	oldCoverColor := flagPDFCoverColor
	oldOutput := flagPDFOutput
	oldShuffle := flagPDFShuffleSeed

	return func() {
		flagPDFTitle = oldTitle
		flagPDFHeader = oldHeader
		flagPDFVolume = oldVolume
		flagPDFAdvert = oldAdvert
		flagPDFCoverColor = oldCoverColor
		flagPDFOutput = oldOutput
		flagPDFShuffleSeed = oldShuffle
	}
}
