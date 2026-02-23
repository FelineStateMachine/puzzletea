package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/pdfexport"

	"github.com/spf13/cobra"
)

func TestRunNewExportRejectsUnsupportedGame(t *testing.T) {
	withExportFlagReset(t)
	output := filepath.Join(t.TempDir(), "lights.jsonl")
	flagOutput = output

	cmd, _ := newExportTestCmd(t, false)
	err := runNewExport(cmd, []string{"lights-out"})
	if err != nil {
		t.Fatalf("expected silent no-op for unsupported game, got: %v", err)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Fatalf("expected no output file, stat err = %v", statErr)
	}
}

func TestRunNewExportValidation(t *testing.T) {
	t.Run("writes jsonl to stdout when output omitted", func(t *testing.T) {
		withExportFlagReset(t)
		flagExport = 2

		cmd, out := newExportTestCmd(t, true)
		err := runNewExport(cmd, []string{"nonogram", "mini"})
		if err != nil {
			t.Fatalf("expected stdout export success, got error: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(out.String()), "\n")
		if got, want := len(lines), 2; got != want {
			t.Fatalf("jsonl lines = %d, want %d", got, want)
		}
		for i, line := range lines {
			var record pdfexport.JSONLRecord
			if err := json.Unmarshal([]byte(line), &record); err != nil {
				t.Fatalf("line %d is not valid jsonl: %v", i+1, err)
			}
			if record.Schema != pdfexport.ExportSchemaV1 {
				t.Fatalf("line %d schema = %q, want %q", i+1, record.Schema, pdfexport.ExportSchemaV1)
			}
		}
	})

	t.Run("output extension must be jsonl", func(t *testing.T) {
		withExportFlagReset(t)
		flagOutput = filepath.Join(t.TempDir(), "out.txt")

		cmd, _ := newExportTestCmd(t, false)
		err := runNewExport(cmd, []string{"nonogram", "mini"})
		if err == nil {
			t.Fatal("expected extension validation error")
		}
		if !strings.Contains(err.Error(), ".jsonl extension") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("set-seed cannot be combined with output", func(t *testing.T) {
		withExportFlagReset(t)
		flagSetSeed = "abc"
		flagOutput = filepath.Join(t.TempDir(), "out.jsonl")

		cmd, _ := newExportTestCmd(t, false)
		err := runNewExport(cmd, []string{"nonogram", "mini"})
		if err == nil {
			t.Fatal("expected set-seed validation error")
		}
		if !strings.Contains(err.Error(), "--set-seed cannot be combined with export (--export/--output)") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRunNewExportReproducibleWithSeed(t *testing.T) {
	withExportFlagReset(t)

	fixedNow := time.Date(2026, 2, 22, 11, 0, 0, 0, time.UTC)
	previousNow := exportNow
	exportNow = func() time.Time { return fixedNow }
	t.Cleanup(func() { exportNow = previousNow })

	fileA := filepath.Join(t.TempDir(), "a.jsonl")
	fileB := filepath.Join(t.TempDir(), "b.jsonl")

	flagExport = 3
	flagWithSeed = "zine-seed-01"
	flagOutput = fileA
	cmdA, _ := newExportTestCmd(t, false)
	if err := runNewExport(cmdA, []string{"nonogram", "mini"}); err != nil {
		t.Fatalf("first export failed: %v", err)
	}

	flagOutput = fileB
	cmdB, _ := newExportTestCmd(t, false)
	if err := runNewExport(cmdB, []string{"nonogram", "mini"}); err != nil {
		t.Fatalf("second export failed: %v", err)
	}

	a, err := os.ReadFile(fileA)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(fileB)
	if err != nil {
		t.Fatal(err)
	}
	if string(a) != string(b) {
		t.Fatal("expected deterministic jsonl output for identical seed and args")
	}
}

func TestRunNewExportOverwritesOutputFile(t *testing.T) {
	withExportFlagReset(t)

	file := filepath.Join(t.TempDir(), "out.jsonl")
	if err := os.WriteFile(file, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	flagExport = 1
	flagWithSeed = "overwrite-seed"
	flagOutput = file

	cmd, _ := newExportTestCmd(t, false)
	if err := runNewExport(cmd, []string{"nonogram", "mini"}); err != nil {
		t.Fatalf("export failed: %v", err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "old" {
		t.Fatal("expected output file to be overwritten")
	}
	if !strings.Contains(string(data), pdfexport.ExportSchemaV1) {
		t.Fatal("expected jsonl export schema marker")
	}
}

func TestRunNewExportOmitsPrintPayload(t *testing.T) {
	withExportFlagReset(t)
	flagExport = 1

	cmd, out := newExportTestCmd(t, true)
	if err := runNewExport(cmd, []string{"sudoku", "easy"}); err != nil {
		t.Fatalf("expected sudoku export success, got error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("jsonl lines = %d, want 1", len(lines))
	}

	var record map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &record); err != nil {
		t.Fatal(err)
	}
	if _, ok := record["print"]; ok {
		t.Fatal("did not expect print payload in export record")
	}

	puzzle, ok := record["puzzle"].(map[string]any)
	if !ok {
		t.Fatalf("expected puzzle object, got %T", record["puzzle"])
	}
	if _, ok := puzzle["snippet"]; ok {
		t.Fatal("did not expect markdown snippet in export record")
	}
}

func withExportFlagReset(t *testing.T) {
	t.Helper()

	prevSetSeed := flagSetSeed
	prevWithSeed := flagWithSeed
	prevExport := flagExport
	prevOutput := flagOutput

	flagSetSeed = ""
	flagWithSeed = ""
	flagExport = 1
	flagOutput = ""

	t.Cleanup(func() {
		flagSetSeed = prevSetSeed
		flagWithSeed = prevWithSeed
		flagExport = prevExport
		flagOutput = prevOutput
	})
}

func newExportTestCmd(t *testing.T, exportChanged bool) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.Flags().Int("export", 1, "")
	if exportChanged {
		if err := cmd.Flags().Set("export", strconv.Itoa(flagExport)); err != nil {
			t.Fatal(err)
		}
	}
	return cmd, &out
}
