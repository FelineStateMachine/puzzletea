package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/netwalk"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
	"github.com/FelineStateMachine/puzzletea/sudoku"
	"github.com/FelineStateMachine/puzzletea/theme"
	"github.com/spf13/cobra"
)

func TestRunTestRendersANSIToStdout(t *testing.T) {
	reset := snapshotTestFlags()
	defer reset()

	input := filepath.Join(t.TempDir(), "visual.jsonl")
	data, err := netwalk.VisualFixtureJSONL()
	if err != nil {
		t.Fatalf("VisualFixtureJSONL() error = %v", err)
	}
	if err := os.WriteFile(input, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd, out := newTestCmd()
	if err := runTest(cmd, []string{input}); err != nil {
		t.Fatalf("runTest() error = %v", err)
	}

	rendered := out.String()
	if !strings.Contains(rendered, "=== Netwalk | Visual Fixture | cursor-root-horizontal | #1 ===") {
		t.Fatalf("expected first section header, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "\x1b[") {
		t.Fatal("expected ANSI escape sequences in review output")
	}
}

func TestRunTestWritesOutputFile(t *testing.T) {
	reset := snapshotTestFlags()
	defer reset()

	input := filepath.Join(t.TempDir(), "visual.jsonl")
	data, err := netwalk.VisualFixtureJSONL()
	if err != nil {
		t.Fatalf("VisualFixtureJSONL() error = %v", err)
	}
	if err := os.WriteFile(input, data, 0o644); err != nil {
		t.Fatal(err)
	}

	testOutput = filepath.Join(t.TempDir(), "review.txt")
	cmd, out := newTestCmd()
	if err := runTest(cmd, []string{input}); err != nil {
		t.Fatalf("runTest() error = %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected stdout to stay empty when --output is set, got %q", out.String())
	}

	written, err := os.ReadFile(testOutput)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(written), "solved-with-empty-cells") {
		t.Fatalf("expected written output to contain fixture case names, got:\n%s", string(written))
	}
}

func TestRunTestRejectsBadSchema(t *testing.T) {
	reset := snapshotTestFlags()
	defer reset()

	input := filepath.Join(t.TempDir(), "bad.jsonl")
	record := pdfexport.JSONLRecord{Schema: "bad.schema"}
	line, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(input, append(line, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd, _ := newTestCmd()
	err = runTest(cmd, []string{input})
	if err == nil {
		t.Fatal("expected unsupported schema error")
	}
	if !strings.Contains(err.Error(), "unsupported schema") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunTestRendersMixedGamesInInputOrder(t *testing.T) {
	reset := snapshotTestFlags()
	defer reset()

	input := filepath.Join(t.TempDir(), "mixed.jsonl")
	records := []pdfexport.JSONLRecord{
		testJSONLRecord(t, "Netwalk", "Visual Fixture", "netwalk-first", 1, spawnSeededSave(t, netwalk.Modes[0].(game.SeededSpawner))),
		testJSONLRecord(t, "Sudoku", "Easy", "sudoku-second", 2, spawnSeededSave(t, sudoku.Modes[0].(game.SeededSpawner))),
	}
	writeTestJSONL(t, input, records)

	cmd, out := newTestCmd()
	if err := runTest(cmd, []string{input}); err != nil {
		t.Fatalf("runTest() error = %v", err)
	}

	rendered := out.String()
	first := strings.Index(rendered, "netwalk-first")
	second := strings.Index(rendered, "sudoku-second")
	if first == -1 || second == -1 {
		t.Fatalf("expected both records in output, got:\n%s", rendered)
	}
	if first > second {
		t.Fatalf("expected input order to be preserved, got:\n%s", rendered)
	}
}

func TestRunTestIgnoresPersistedThemeWithoutFlag(t *testing.T) {
	reset := snapshotTestFlags()
	defer reset()
	t.Cleanup(func() { _ = theme.Apply("") })

	configPath, _ := writeCommandConfig(t)
	flagConfigPath = configPath
	if err := theme.Apply(""); err != nil {
		t.Fatalf("theme.Apply() error = %v", err)
	}
	defaultFG := fmt.Sprint(theme.Current().FG)
	if err := theme.Apply("Dracula"); err != nil {
		t.Fatalf("theme.Apply() error = %v", err)
	}

	input := filepath.Join(t.TempDir(), "visual.jsonl")
	data, err := netwalk.VisualFixtureJSONL()
	if err != nil {
		t.Fatalf("VisualFixtureJSONL() error = %v", err)
	}
	if err := os.WriteFile(input, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd, _ := newTestCmd()
	if err := runTest(cmd, []string{input}); err != nil {
		t.Fatalf("runTest() error = %v", err)
	}

	if got, want := fmt.Sprint(theme.Current().FG), defaultFG; got != want {
		t.Fatalf("test command theme FG = %v, want default %v", got, want)
	}
}

func TestRunTestUsesExplicitThemeFlag(t *testing.T) {
	reset := snapshotTestFlags()
	defer reset()
	t.Cleanup(func() { _ = theme.Apply("") })

	configPath, _ := writeCommandConfig(t)
	flagConfigPath = configPath
	flagTheme = "Dracula"

	input := filepath.Join(t.TempDir(), "visual.jsonl")
	data, err := netwalk.VisualFixtureJSONL()
	if err != nil {
		t.Fatalf("VisualFixtureJSONL() error = %v", err)
	}
	if err := os.WriteFile(input, data, 0o644); err != nil {
		t.Fatal(err)
	}

	cmd, _ := newTestCmd()
	if err := runTest(cmd, []string{input}); err != nil {
		t.Fatalf("runTest() error = %v", err)
	}

	want := theme.LookupTheme(flagTheme).Palette()
	if got := fmt.Sprint(theme.Current().FG); got != fmt.Sprint(want.FG) {
		t.Fatalf("theme override FG = %v, want %v", theme.Current().FG, want.FG)
	}
}

func snapshotTestFlags() func() {
	prevOutput := testOutput
	prevTheme := flagTheme
	prevConfigPath := flagConfigPath
	testOutput = ""
	return func() {
		testOutput = prevOutput
		flagTheme = prevTheme
		flagConfigPath = prevConfigPath
	}
}

func newTestCmd() (*cobra.Command, *bytes.Buffer) {
	cmd := &cobra.Command{}
	var out bytes.Buffer
	cmd.SetOut(&out)
	return cmd, &out
}

func writeTestJSONL(t *testing.T, path string, records []pdfexport.JSONLRecord) {
	t.Helper()

	var data []byte
	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, line...)
		data = append(data, '\n')
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func testJSONLRecord(
	t *testing.T,
	gameName, modeName, name string,
	index int,
	save []byte,
) pdfexport.JSONLRecord {
	t.Helper()

	return pdfexport.JSONLRecord{
		Schema: pdfexport.ExportSchemaV1,
		Pack: pdfexport.JSONLPackMeta{
			Generated:     "2026-03-14T00:00:00Z",
			Version:       "test",
			Category:      gameName,
			ModeSelection: modeName,
			Count:         2,
		},
		Puzzle: pdfexport.JSONLPuzzle{
			Index: index,
			Name:  name,
			Game:  gameName,
			Mode:  modeName,
			Save:  json.RawMessage(save),
		},
	}
}

func spawnSeededSave(t *testing.T, spawner game.SeededSpawner) []byte {
	t.Helper()

	rng := rand.New(rand.NewPCG(1, 2))
	g, err := spawner.SpawnSeeded(rng)
	if err != nil {
		t.Fatalf("SpawnSeeded() error = %v", err)
	}
	save, err := g.GetSave()
	if err != nil {
		t.Fatalf("GetSave() error = %v", err)
	}
	return save
}
