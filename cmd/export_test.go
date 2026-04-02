package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/FelineStateMachine/puzzletea/export/pack"
	"github.com/FelineStateMachine/puzzletea/puzzle"
	"github.com/spf13/cobra"
)

func TestExportCmdRunsSpecFile(t *testing.T) {
	restore := snapshotCommandGlobals(t)
	defer restore()

	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.json")
	pdfPath := filepath.Join(dir, "sample.pdf")
	jsonlPath := filepath.Join(dir, "sample.jsonl")

	spec := packexport.Spec{
		Title:           "CLI Export",
		Volume:          1,
		SheetLayout:     "half-letter",
		PDFOutputPath:   pdfPath,
		JSONLOutputPath: jsonlPath,
		Counts: map[puzzle.GameID]map[puzzle.ModeID]int{
			puzzle.CanonicalGameID("Sudoku"): {
				puzzle.CanonicalModeID("Easy"): 1,
			},
		},
	}
	data, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(specPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	exportSpecPath = specPath
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)

	if err := exportCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("exportCmd.RunE returned error: %v", err)
	}
	if out.String() == "" {
		t.Fatal("expected command output")
	}
	if _, err := os.Stat(pdfPath); err != nil {
		t.Fatalf("expected pdf output: %v", err)
	}
	if _, err := os.Stat(jsonlPath); err != nil {
		t.Fatalf("expected jsonl output: %v", err)
	}
}
