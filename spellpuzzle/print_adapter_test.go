package spellpuzzle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/FelineStateMachine/puzzletea/game"
	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

func TestPrintAdapterReturnsNilForBlankSave(t *testing.T) {
	payload, err := printAdapter{}.BuildPDFPayload(nil)
	if err != nil {
		t.Fatal(err)
	}
	if payload != nil {
		t.Fatalf("payload = %#v, want nil", payload)
	}
}

func TestPrintAdapterBuildsPayloadAndIgnoresFoundState(t *testing.T) {
	save := Save{
		ModeTitle: "print-test",
		Bank:      "PATACN",
		Placements: []WordPlacement{
			{Text: "CATNAP", Start: Position{X: 3, Y: 0}, Orientation: Vertical, Found: true},
			{Text: "CAPTAN", Start: Position{X: 0, Y: 2}, Orientation: Vertical, Found: false},
			{Text: "CANT", Start: Position{X: 0, Y: 2}, Orientation: Horizontal, Found: true},
		},
		BonusWords: []string{"TAN", "ACT"},
		Trace:      []int{0, 1},
		BankCursor: 1,
		Solved:     true,
		Feedback:   feedback{Kind: feedbackSuccess, Message: "bonus word counted"},
	}

	data, err := json.Marshal(save)
	if err != nil {
		t.Fatal(err)
	}

	payload, err := printAdapter{}.BuildPDFPayload(data)
	if err != nil {
		t.Fatal(err)
	}

	got, ok := payload.(*printPayload)
	if !ok || got == nil {
		t.Fatal("expected print payload")
	}
	if got.Board.Width <= 0 || got.Board.Height <= 0 {
		t.Fatalf("board dims = %dx%d, want positive", got.Board.Width, got.Board.Height)
	}
	if string(got.Bank) != "PATACN" {
		t.Fatalf("bank = %q, want PATACN", string(got.Bank))
	}
	for y := range got.Board.Height {
		for x := range got.Board.Width {
			if got.Board.Cells[y][x].Revealed {
				t.Fatalf("cell (%d,%d) unexpectedly revealed", x, y)
			}
		}
	}
}

func TestPrintAdapterDeduplicatesDuplicatePlacements(t *testing.T) {
	save := Save{
		ModeTitle: "opal-crane",
		Bank:      "PATACN",
		Placements: []WordPlacement{
			{Text: "CATNAP", Start: Position{X: 3, Y: 0}, Orientation: Vertical, Found: true},
			{Text: "CAPTAN", Start: Position{X: 0, Y: 2}, Orientation: Vertical, Found: true},
			{Text: "CANT", Start: Position{X: 0, Y: 2}, Orientation: Horizontal, Found: false},
			{Text: "CANT", Start: Position{X: 0, Y: 2}, Orientation: Horizontal, Found: true},
		},
	}

	data, err := json.Marshal(save)
	if err != nil {
		t.Fatal(err)
	}

	payload, err := printAdapter{}.BuildPDFPayload(data)
	if err != nil {
		t.Fatal(err)
	}

	got := payload.(*printPayload)
	if len(got.Placements) != 3 {
		t.Fatalf("placement count = %d, want 3", len(got.Placements))
	}
}

func TestSpellPuzzlePrintAdapterRegistration(t *testing.T) {
	pdfexport.RegisterPrintAdapter(PDFPrintAdapter)

	for _, gameType := range []string{"Spell Puzzle", "spell", "spellpuzzle"} {
		if !game.HasPrintAdapter(gameType) {
			t.Fatalf("expected print adapter for %q", gameType)
		}
	}
}

func TestSpellPuzzleJSONLHydratesPrintPayload(t *testing.T) {
	pdfexport.RegisterPrintAdapter(PDFPrintAdapter)

	record := pdfexport.JSONLRecord{
		Schema: pdfexport.ExportSchemaV1,
		Pack: pdfexport.JSONLPackMeta{
			Generated:     "2026-03-10T10:00:00Z",
			Version:       "v-test",
			Category:      "Spell Puzzle",
			ModeSelection: "Beginner",
			Count:         1,
		},
		Puzzle: pdfexport.JSONLPuzzle{
			Index: 1,
			Name:  "print-bank",
			Game:  "Spell Puzzle",
			Mode:  "Beginner",
			Save: json.RawMessage(`{
				"mode_title":"print-bank",
				"bank":"PATACN",
				"placements":[
					{"text":"CATNAP","start":{"x":3,"y":0},"orientation":1,"found":true},
					{"text":"CAPTAN","start":{"x":0,"y":2},"orientation":1,"found":false},
					{"text":"CANT","start":{"x":0,"y":2},"orientation":0,"found":true}
				]
			}`),
		},
	}

	path := filepath.Join(t.TempDir(), "spellpuzzle.jsonl")
	data, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := pdfexport.ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(doc.Puzzles), 1; got != want {
		t.Fatalf("puzzles = %d, want %d", got, want)
	}
	if _, ok := doc.Puzzles[0].PrintPayload.(*printPayload); !ok {
		t.Fatalf("payload type = %T, want *printPayload", doc.Puzzles[0].PrintPayload)
	}
}

func TestSpellPuzzlePDFRenderSmokeForAllModes(t *testing.T) {
	pdfexport.RegisterPrintAdapter(PDFPrintAdapter)

	puzzles := make([]pdfexport.Puzzle, 0, len(Modes))
	for i, item := range Modes {
		mode := item.(SpellPuzzleMode)
		generated, err := GeneratePuzzleSeeded(mode, newTestRNG(uint64(100+i)))
		if err != nil {
			t.Fatalf("%s: %v", mode.Title(), err)
		}

		saveData, err := json.Marshal(Save{
			ModeTitle:  mode.Title(),
			Bank:       generated.Bank,
			Placements: generated.Placements,
		})
		if err != nil {
			t.Fatal(err)
		}

		payload, err := printAdapter{}.BuildPDFPayload(saveData)
		if err != nil {
			t.Fatalf("%s payload: %v", mode.Title(), err)
		}
		puzzles = append(puzzles, pdfexport.Puzzle{
			Category:        "Spell Puzzle",
			Name:            mode.Title(),
			Index:           i + 1,
			ModeSelection:   mode.Title(),
			PrintPayload:    payload,
			DifficultyScore: 0.5,
		})
	}

	output := filepath.Join(t.TempDir(), "spellpuzzle.pdf")
	err := pdfexport.WritePDF(output, nil, puzzles, pdfexport.RenderConfig{
		VolumeNumber: 1,
		GeneratedAt:  time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty pdf output")
	}
}
