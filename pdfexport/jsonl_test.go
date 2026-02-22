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

func TestParseJSONLFileHydratesTakuzuFromSave(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "takuzu-pack.jsonl")

	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Takuzu",
			ModeSelection: "Beginner",
			Count:         1,
			Seed:          "seed-2",
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "binary-wave",
			Game:  "Takuzu",
			Mode:  "Beginner",
			Save:  json.RawMessage(`{"size":2,"state":"01\n10","provided":"##\n#."}`),
		},
		Print: JSONLPrintData{
			Kind: "grid-table",
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
	if doc.Puzzles[0].Takuzu == nil {
		t.Fatal("expected takuzu print payload from save hydration")
	}
	if got, want := doc.Puzzles[0].Takuzu.Givens[1][1], ""; got != want {
		t.Fatalf("takuzu row 1 col 1 = %q, want empty", got)
	}
}

func TestParseJSONLFileHydratesSudokuFromSave(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "sudoku.jsonl")

	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Sudoku",
			ModeSelection: "Easy",
			Count:         1,
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "moss-pine",
			Game:  "Sudoku",
			Mode:  "Easy",
			Save:  json.RawMessage(`{"grid":"500000000\n000000000\n000000000\n000000000\n000000000\n000000000\n000000000\n000000000\n000000000","provided":[{"x":0,"y":0,"v":5}]}`),
		},
		Print: JSONLPrintData{
			Kind: "sudoku",
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
	if doc.Puzzles[0].Sudoku == nil {
		t.Fatal("expected sudoku print payload from save hydration")
	}
	if got, want := doc.Puzzles[0].Sudoku.Givens[0][0], 5; got != want {
		t.Fatalf("sudoku givens[0][0] = %d, want %d", got, want)
	}
}

func TestParseJSONLFileHydratesWordSearchFromSave(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "wordsearch.jsonl")

	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Word Search",
			ModeSelection: "Standard",
			Count:         1,
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "sage-sky",
			Game:  "Word Search",
			Mode:  "Standard",
			Save:  json.RawMessage(`{"width":3,"height":3,"grid":"abc\ndef\nghi","words":[{"text":"ace"},{"text":"fig"}]}`),
		},
		Print: JSONLPrintData{
			Kind: "word-search",
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
	if doc.Puzzles[0].WordSearch == nil {
		t.Fatal("expected word search print payload from save hydration")
	}
	if got, want := len(doc.Puzzles[0].WordSearch.Words), 2; got != want {
		t.Fatalf("word count = %d, want %d", got, want)
	}
	if got, want := doc.Puzzles[0].WordSearch.Words[0], "ACE"; got != want {
		t.Fatalf("first word = %q, want %q", got, want)
	}
}

func TestParseJSONLFileHydratesNurikabeFromSave(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "nurikabe.jsonl")

	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Nurikabe",
			ModeSelection: "Mini",
			Count:         1,
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "mist-pond",
			Game:  "Nurikabe",
			Mode:  "Mini",
			Save:  json.RawMessage(`{"width":2,"height":2,"clues":"1,0\n0,2","marks":"??\n??"}`),
		},
		Print: JSONLPrintData{
			Kind: "nurikabe",
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
	if doc.Puzzles[0].Nurikabe == nil {
		t.Fatal("expected nurikabe print payload from save hydration")
	}
	if got, want := doc.Puzzles[0].Nurikabe.Clues[1][1], 2; got != want {
		t.Fatalf("nurikabe clues[1][1] = %d, want %d", got, want)
	}
}

func TestParseJSONLFileHydratesShikakuFromSave(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "shikaku.jsonl")

	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Shikaku",
			ModeSelection: "Mini",
			Count:         1,
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "ridge-glen",
			Game:  "Shikaku",
			Mode:  "Mini",
			Save:  json.RawMessage(`{"width":2,"height":2,"clues":[{"x":0,"y":0,"value":1},{"x":1,"y":1,"value":4}]}`),
		},
		Print: JSONLPrintData{
			Kind: "shikaku",
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
	if doc.Puzzles[0].Shikaku == nil {
		t.Fatal("expected shikaku print payload from save hydration")
	}
	if got, want := doc.Puzzles[0].Shikaku.Clues[1][1], 4; got != want {
		t.Fatalf("shikaku clues[1][1] = %d, want %d", got, want)
	}
}

func TestParseJSONLFileHydratesHashiFromSave(t *testing.T) {
	temp := t.TempDir()
	path := filepath.Join(temp, "hashi.jsonl")

	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Hashiwokakero",
			ModeSelection: "Standard",
			Count:         1,
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "reef-palm",
			Game:  "Hashiwokakero",
			Mode:  "Standard",
			Save:  json.RawMessage(`{"width":7,"height":7,"islands":[{"x":0,"y":0,"required":3},{"x":6,"y":6,"required":2}],"bridges":[]}`),
		},
		Print: JSONLPrintData{
			Kind: "hashi",
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
	if doc.Puzzles[0].Hashi == nil {
		t.Fatal("expected hashi print payload from save hydration")
	}
	if got, want := len(doc.Puzzles[0].Hashi.Islands), 2; got != want {
		t.Fatalf("hashi island count = %d, want %d", got, want)
	}
}
