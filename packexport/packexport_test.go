package packexport

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

func TestRunDeterministicWithFixedSpecAndSeed(t *testing.T) {
	restore := snapshotNowFn(t, time.Date(2026, 3, 22, 18, 0, 0, 0, time.UTC))
	defer restore()

	dir := t.TempDir()
	specA := smallTestSpec(dir, "a")
	specB := smallTestSpec(dir, "b")

	if _, err := Run(t.Context(), specA); err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if _, err := Run(t.Context(), specB); err != nil {
		t.Fatalf("second run failed: %v", err)
	}

	jsonlA, err := os.ReadFile(specA.JSONLOutputPath)
	if err != nil {
		t.Fatal(err)
	}
	jsonlB, err := os.ReadFile(specB.JSONLOutputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(jsonlA) != string(jsonlB) {
		t.Fatal("expected identical jsonl output for identical spec and seed")
	}

	info, err := os.Stat(specA.PDFOutputPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty pdf output")
	}
}

func TestRunSkipsZeroCountModes(t *testing.T) {
	restore := snapshotNowFn(t, time.Date(2026, 3, 22, 18, 0, 0, 0, time.UTC))
	defer restore()

	dir := t.TempDir()
	spec := smallTestSpec(dir, "single")
	spec.Counts[puzzle.CanonicalGameID("Word Search")][puzzle.CanonicalModeID("Easy 10x10")] = 0

	result, err := Run(t.Context(), spec)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if result.TotalCount != 1 {
		t.Fatalf("TotalCount = %d, want 1", result.TotalCount)
	}

	data, err := os.ReadFile(spec.JSONLOutputPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if got := len(lines); got != 1 {
		t.Fatalf("jsonl line count = %d, want 1", got)
	}
	if strings.Contains(string(data), "Word Search") {
		t.Fatal("did not expect zero-count mode to appear in output")
	}
}

func TestValidateSpecRejectsUnknownTargets(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name string
		spec Spec
		want string
	}{
		{
			name: "unknown game",
			spec: func() Spec {
				spec := smallTestSpec(dir, "unknown-game")
				spec.Counts = zeroCounts()
				spec.Counts[puzzle.GameID("missing-game")] = map[puzzle.ModeID]int{
					puzzle.ModeID("missing-mode"): 1,
				}
				return spec
			}(),
			want: "not a known export target",
		},
		{
			name: "unknown mode",
			spec: func() Spec {
				spec := smallTestSpec(dir, "unknown-mode")
				spec.Counts = zeroCounts()
				spec.Counts[puzzle.CanonicalGameID("Sudoku")][puzzle.ModeID("missing-mode")] = 1
				return spec
			}(),
			want: "not a known export mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpec(tt.spec)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}

func TestRunDoesNotPublishOutputsAfterCancellation(t *testing.T) {
	restoreNow := snapshotNowFn(t, time.Date(2026, 3, 22, 18, 0, 0, 0, time.UTC))
	defer restoreNow()

	prevWritePDF := writePDFFn
	t.Cleanup(func() {
		writePDFFn = prevWritePDF
	})

	dir := t.TempDir()
	spec := smallTestSpec(dir, "canceled")
	ctx, cancel := context.WithCancel(t.Context())
	writePDFFn = func(outputPath string, _ []pdfexport.PackDocument, _ []pdfexport.Puzzle, _ pdfexport.RenderConfig) error {
		if err := os.WriteFile(outputPath, []byte("%PDF-test"), 0o644); err != nil {
			return err
		}
		cancel()
		return nil
	}

	_, err := Run(ctx, spec)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run error = %v, want context.Canceled", err)
	}
	if _, err := os.Stat(spec.PDFOutputPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no published pdf output, stat err = %v", err)
	}
	if _, err := os.Stat(spec.JSONLOutputPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no published jsonl output, stat err = %v", err)
	}
}

func snapshotNowFn(t *testing.T, now time.Time) func() {
	t.Helper()

	prev := nowFn
	nowFn = func() time.Time { return now }
	return func() { nowFn = prev }
}

func smallTestSpec(dir, suffix string) Spec {
	spec := DefaultSpec(dir)
	spec.Counts = zeroCounts()
	spec.Title = "Deterministic Sampler"
	spec.Seed = "packexport-test-seed"
	spec.PDFOutputPath = filepath.Join(dir, "pack-"+suffix+".pdf")
	spec.JSONLOutputPath = filepath.Join(dir, "pack-"+suffix+".jsonl")
	spec.Counts[puzzle.CanonicalGameID("Sudoku")][puzzle.CanonicalModeID("Easy")] = 1
	spec.Counts[puzzle.CanonicalGameID("Word Search")][puzzle.CanonicalModeID("Easy 10x10")] = 1
	return spec
}
