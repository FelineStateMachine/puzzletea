package packexport

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/export/pdf"
	"github.com/FelineStateMachine/puzzletea/puzzle"
)

func TestAnnotatePuzzlesForPrintUsesLegacyModeOrder(t *testing.T) {
	definitions := []puzzle.Definition{{
		Name: "Sudoku",
		Variants: []puzzle.VariantDef{
			puzzle.NewVariantDef(puzzle.VariantSpec{Title: "Sudoku", DefaultElo: 1200}),
		},
		LegacyModes: []puzzle.LegacyModeAlias{
			puzzle.NewLegacyModeAlias(puzzle.LegacyModeAliasSpec{
				Title:           "Easy",
				TargetVariantID: puzzle.CanonicalVariantID("Sudoku"),
				PresetElo:       600,
				XPWeight:        1,
			}),
			puzzle.NewLegacyModeAlias(puzzle.LegacyModeAliasSpec{
				Title:           "Medium",
				TargetVariantID: puzzle.CanonicalVariantID("Sudoku"),
				PresetElo:       1200,
				XPWeight:        3,
			}),
			puzzle.NewLegacyModeAlias(puzzle.LegacyModeAliasSpec{
				Title:           "Hard",
				TargetVariantID: puzzle.CanonicalVariantID("Sudoku"),
				PresetElo:       1800,
				XPWeight:        7,
			}),
		},
		Modes: []puzzle.ModeDef{
			{ID: puzzle.CanonicalModeID("Easy"), Title: "Easy"},
			{ID: puzzle.CanonicalModeID("Medium"), Title: "Medium"},
			{ID: puzzle.CanonicalModeID("Hard"), Title: "Hard"},
		},
	}}
	puzzles := []pdfexport.Puzzle{{
		Category:      "Sudoku",
		ModeSelection: "Hard",
	}}

	AnnotatePuzzlesForPrint(puzzles, definitions)

	if got, want := puzzles[0].DifficultyScore, 1.0; got != want {
		t.Fatalf("DifficultyScore = %v, want %v", got, want)
	}
	if got, want := puzzles[0].DifficultyConfidence, pdfexport.DifficultyConfidenceHigh; got != want {
		t.Fatalf("DifficultyConfidence = %q, want %q", got, want)
	}
	if got, want := puzzles[0].DifficultySource, "mode-order"; got != want {
		t.Fatalf("DifficultySource = %q, want %q", got, want)
	}
}
