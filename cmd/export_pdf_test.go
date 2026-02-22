package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/pdfexport"
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

func snapshotExportPDFFlags() func() {
	oldTitle := flagPDFTitle
	oldVolume := flagPDFVolume
	oldAdvert := flagPDFAdvert
	oldCoverColor := flagPDFCoverColor
	oldOutput := flagPDFOutput
	oldShuffle := flagPDFShuffleSeed

	return func() {
		flagPDFTitle = oldTitle
		flagPDFVolume = oldVolume
		flagPDFAdvert = oldAdvert
		flagPDFCoverColor = oldCoverColor
		flagPDFOutput = oldOutput
		flagPDFShuffleSeed = oldShuffle
	}
}
