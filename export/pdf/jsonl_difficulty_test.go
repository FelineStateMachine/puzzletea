package pdfexport

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestParseJSONLFilePrefersDifficultyElo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pack.jsonl")
	targetElo := 1200
	actualElo := 1800
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
			Index:                1,
			Name:                 "ember-newt",
			Game:                 "Nonogram",
			Mode:                 "Mini",
			TargetDifficultyElo:  &targetElo,
			ActualDifficultyElo:  &actualElo,
			DifficultyConfidence: "high",
			Save:                 json.RawMessage(`{"width":2,"height":2,"row-hints":[[1],[1]],"col-hints":[[1],[1]],"state":"  \n  "}`),
		},
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	puzzle := doc.Puzzles[0]
	if puzzle.TargetDifficultyElo == nil || *puzzle.TargetDifficultyElo != targetElo {
		t.Fatalf("TargetDifficultyElo = %v, want %d", puzzle.TargetDifficultyElo, targetElo)
	}
	if puzzle.ActualDifficultyElo == nil || *puzzle.ActualDifficultyElo != actualElo {
		t.Fatalf("ActualDifficultyElo = %v, want %d", puzzle.ActualDifficultyElo, actualElo)
	}
	if got, want := puzzle.DifficultyScore, 0.6; got != want {
		t.Fatalf("DifficultyScore = %v, want %v", got, want)
	}
	if got, want := puzzle.DifficultyConfidence, DifficultyConfidenceHigh; got != want {
		t.Fatalf("DifficultyConfidence = %q, want %q", got, want)
	}
	if got, want := puzzle.DifficultySource, "actual-elo"; got != want {
		t.Fatalf("DifficultySource = %q, want %q", got, want)
	}
}

func TestParseJSONLFileFallsBackToTargetDifficultyElo(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pack.jsonl")
	targetElo := 1500
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
			Index:                1,
			Name:                 "ember-newt",
			Game:                 "Nonogram",
			Mode:                 "Mini",
			TargetDifficultyElo:  &targetElo,
			DifficultyConfidence: "medium",
			Save:                 json.RawMessage(`{"width":2,"height":2,"row-hints":[[1],[1]],"col-hints":[[1],[1]],"state":"  \n  "}`),
		},
	}
	writeSingleJSONLRecord(t, path, record)

	doc, err := ParseJSONLFile(path)
	if err != nil {
		t.Fatal(err)
	}
	puzzle := doc.Puzzles[0]
	if got, want := puzzle.DifficultyScore, 0.5; got != want {
		t.Fatalf("DifficultyScore = %v, want %v", got, want)
	}
	if got, want := puzzle.DifficultySource, "target-elo"; got != want {
		t.Fatalf("DifficultySource = %q, want %q", got, want)
	}
}
