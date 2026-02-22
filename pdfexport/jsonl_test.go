package pdfexport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseJSONLFile(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "pack.jsonl")

	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Nonogram",
			ModeSelection: "Mini",
			Count:         1,
			Seed:          "seed-1",
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "ember-newt",
			Game:  "Nonogram",
			Mode:  "Mini",
			Save:  json.RawMessage(`{"width":2}`),
			Snippet: "### Puzzle Grid with Integrated Hints\n\n" +
				"| R1 | C1 | C2 |\n" +
				"| --- | --- | --- |\n" +
				"| . | 1 | 2 |\n" +
				"| 1 | . | . |\n",
		},
		Print: JSONLPrintData{
			Kind:       "nonogram",
			Paper:      "A5",
			MarginMM:   10,
			EmptyGlyph: " ",
			HintTone:   "dim",
		},
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	if doc.Puzzles[0].Nonogram == nil {
		t.Fatal("expected nonogram print payload from snippet fallback")
	}
}

func TestParseJSONLFileRejectsNonJSONLExtension(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "pack.md")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ParseJSONLFile(path)
	if err == nil {
		t.Fatal("expected extension validation error")
	}
	if !strings.Contains(err.Error(), "expected .jsonl input") {
		t.Fatalf("unexpected error: %v", err)
	}
}
