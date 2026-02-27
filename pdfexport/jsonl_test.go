package pdfexport

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"codeberg.org/go-pdf/fpdf"
	"github.com/FelineStateMachine/puzzletea/game"
)

func TestParseJSONLFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pack.jsonl")
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
			Save:  json.RawMessage(`{"width":2,"height":2,"row-hints":[[1],[1]],"col-hints":[[1],[1]],"state":"  \n  "}`),
		},
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	payload, ok := doc.Puzzles[0].PrintPayload.(*NonogramData)
	if !ok || payload == nil {
		t.Fatal("expected nonogram print payload from save hydration")
	}
}

func TestParseJSONLFileHydratesNonogramFromSaveWithoutSnippet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonogram-save-only.jsonl")
	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Nonogram",
			ModeSelection: "Mini",
			Count:         1,
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "fern-owl",
			Game:  "Nonogram",
			Mode:  "Mini",
			Save:  json.RawMessage(`{"width":2,"height":2,"row-hints":[[1],[1]],"col-hints":[[1],[1]],"state":"  \n  "}`),
		},
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	payload, ok := doc.Puzzles[0].PrintPayload.(*NonogramData)
	if !ok || payload == nil {
		t.Fatal("expected nonogram print payload from save hydration")
	}
}

func TestParseJSONLFileRejectsNonJSONLExtension(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pack.md")
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
	path := filepath.Join(t.TempDir(), "takuzu-pack.jsonl")
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
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	payload, ok := doc.Puzzles[0].PrintPayload.(*TakuzuData)
	if !ok || payload == nil {
		t.Fatal("expected takuzu print payload from save hydration")
	}
	if got, want := payload.Givens[1][1], ""; got != want {
		t.Fatalf("takuzu row 1 col 1 = %q, want empty", got)
	}
}

func TestParseJSONLFileHydratesSudokuFromSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sudoku.jsonl")
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
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	payload, ok := doc.Puzzles[0].PrintPayload.(*SudokuData)
	if !ok || payload == nil {
		t.Fatal("expected sudoku print payload from save hydration")
	}
	if got, want := payload.Givens[0][0], 5; got != want {
		t.Fatalf("sudoku givens[0][0] = %d, want %d", got, want)
	}
}

func TestParseJSONLFileHydratesWordSearchFromSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wordsearch.jsonl")
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
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	payload, ok := doc.Puzzles[0].PrintPayload.(*WordSearchData)
	if !ok || payload == nil {
		t.Fatal("expected word search print payload from save hydration")
	}
	if got, want := len(payload.Words), 2; got != want {
		t.Fatalf("word count = %d, want %d", got, want)
	}
	if got, want := payload.Words[0], "ACE"; got != want {
		t.Fatalf("first word = %q, want %q", got, want)
	}
}

func TestParseJSONLFileHydratesNurikabeFromSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nurikabe.jsonl")
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
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	payload, ok := doc.Puzzles[0].PrintPayload.(*NurikabeData)
	if !ok || payload == nil {
		t.Fatal("expected nurikabe print payload from save hydration")
	}
	if got, want := payload.Clues[1][1], 2; got != want {
		t.Fatalf("nurikabe clues[1][1] = %d, want %d", got, want)
	}
}

func TestParseJSONLFileHydratesShikakuFromSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shikaku.jsonl")
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
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	payload, ok := doc.Puzzles[0].PrintPayload.(*ShikakuData)
	if !ok || payload == nil {
		t.Fatal("expected shikaku print payload from save hydration")
	}
	if got, want := payload.Clues[1][1], 4; got != want {
		t.Fatalf("shikaku clues[1][1] = %d, want %d", got, want)
	}
}

func TestParseJSONLFileHydratesHashiFromSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hashi.jsonl")
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
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	payload, ok := doc.Puzzles[0].PrintPayload.(*HashiData)
	if !ok || payload == nil {
		t.Fatal("expected hashi print payload from save hydration")
	}
	if got, want := len(payload.Islands), 2; got != want {
		t.Fatalf("hashi island count = %d, want %d", got, want)
	}
}

func TestParseJSONLFileSilentlySkipsUnsupportedGame(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lights.jsonl")
	record := JSONLRecord{
		Schema: ExportSchemaV1,
		Pack: JSONLPackMeta{
			Generated:     "2026-02-22T10:00:00Z",
			Version:       "v-test",
			Category:      "Lights Out",
			ModeSelection: "Standard",
			Count:         1,
		},
		Puzzle: JSONLPuzzle{
			Index: 1,
			Name:  "glow-shore",
			Game:  "Lights Out",
			Mode:  "Standard",
			Save:  json.RawMessage(`{"size":5}`),
		},
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatalf("expected silent no-op for unsupported game, got: %v", err)
	}
	if got := len(doc.Puzzles); got != 0 {
		t.Fatalf("puzzles = %d, want 0", got)
	}
}

func TestParseJSONLFileMetadataFromFirstPrintableRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mixed.jsonl")

	records := []JSONLRecord{
		{
			Schema: ExportSchemaV1,
			Pack: JSONLPackMeta{
				Generated:     "2026-02-22T10:00:00Z",
				Version:       "v-test",
				Category:      "Lights Out",
				ModeSelection: "Standard",
				Count:         2,
				Seed:          "seed-unsupported",
			},
			Puzzle: JSONLPuzzle{
				Index: 1,
				Name:  "glow-shore",
				Game:  "Lights Out",
				Mode:  "Standard",
				Save:  json.RawMessage(`{"size":5}`),
			},
		},
		{
			Schema: ExportSchemaV1,
			Pack: JSONLPackMeta{
				Generated:     "2026-02-22T10:00:00Z",
				Version:       "v-test",
				Category:      "Sudoku",
				ModeSelection: "Easy",
				Count:         2,
				Seed:          "seed-supported",
			},
			Puzzle: JSONLPuzzle{
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
	if err := os.WriteFile(path, lines, 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatalf("expected parse success, got: %v", err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	if got, want := doc.Metadata.Category, "Sudoku"; got != want {
		t.Fatalf("metadata category = %q, want %q", got, want)
	}
	if got, want := doc.Metadata.ModeSelection, "Easy"; got != want {
		t.Fatalf("metadata mode = %q, want %q", got, want)
	}
	if got, want := doc.Metadata.Seed, "seed-supported"; got != want {
		t.Fatalf("metadata seed = %q, want %q", got, want)
	}
}

func writeSingleJSONLRecord(t *testing.T, path string, record JSONLRecord) {
	t.Helper()
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

var registerJSONLAdaptersOnce sync.Once

func init() {
	ensureJSONLTestAdapters()
}

func ensureJSONLTestAdapters() {
	registerJSONLAdaptersOnce.Do(func() {
		register := func(category string, build func(save []byte) (any, error), aliases ...string) {
			game.RegisterPrintAdapter(jsonlTestAdapter{
				category: category,
				aliases:  aliases,
				build:    build,
			})
		}

		register("Nonogram", buildPayloadAdapter("Nonogram", func(save []byte) (any, error) {
			return ParseNonogramPrintData(save)
		}), "nonogram")
		register("Takuzu", buildPayloadAdapter("Takuzu", func(save []byte) (any, error) {
			return ParseTakuzuPrintData(save)
		}), "takuzu")
		register("Sudoku", buildPayloadAdapter("Sudoku", func(save []byte) (any, error) {
			return ParseSudokuPrintData(save)
		}), "sudoku")
		register("Word Search", buildPayloadAdapter("Word Search", func(save []byte) (any, error) {
			return ParseWordSearchPrintData(save)
		}), "wordsearch")
		register("Nurikabe", buildPayloadAdapter("Nurikabe", func(save []byte) (any, error) {
			return ParseNurikabePrintData(save)
		}), "nurikabe")
		register("Shikaku", buildPayloadAdapter("Shikaku", func(save []byte) (any, error) {
			return ParseShikakuPrintData(save)
		}), "shikaku")
		register("Hashiwokakero", buildPayloadAdapter("Hashiwokakero", func(save []byte) (any, error) {
			return ParseHashiPrintData(save)
		}), "hashi")
	})
}

func buildPayloadAdapter(_ string, parse func(save []byte) (any, error)) func([]byte) (any, error) {
	return parse
}

type jsonlTestAdapter struct {
	category string
	aliases  []string
	build    func(save []byte) (any, error)
}

func (a jsonlTestAdapter) CanonicalGameType() string { return a.category }
func (a jsonlTestAdapter) Aliases() []string         { return a.aliases }
func (a jsonlTestAdapter) BuildPDFPayload(save []byte) (any, error) {
	return a.build(save)
}
func (a jsonlTestAdapter) RenderPDFBody(*fpdf.Fpdf, any) error { return nil }
