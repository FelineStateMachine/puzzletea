package pdfexport

import (
	"reflect"
	"testing"
)

func TestOrderPuzzlesForPrintDeterministicWithSeed(t *testing.T) {
	input := []Puzzle{
		{Name: "a", Category: "Sudoku", DifficultyScore: 0.0},
		{Name: "b", Category: "Sudoku", DifficultyScore: 0.1},
		{Name: "c", Category: "Nonogram", DifficultyScore: 0.2},
		{Name: "d", Category: "Nonogram", DifficultyScore: 0.3},
		{Name: "e", Category: "Takuzu", DifficultyScore: 0.4},
		{Name: "f", Category: "Takuzu", DifficultyScore: 0.5},
		{Name: "g", Category: "Hitori", DifficultyScore: 0.6},
		{Name: "h", Category: "Hitori", DifficultyScore: 0.7},
		{Name: "i", Category: "Shikaku", DifficultyScore: 0.8},
		{Name: "j", Category: "Shikaku", DifficultyScore: 0.9},
	}

	orderedA := OrderPuzzlesForPrint(input, "zine-seed")
	orderedB := OrderPuzzlesForPrint(input, "zine-seed")
	if !reflect.DeepEqual(orderedA, orderedB) {
		t.Fatal("expected deterministic ordering with same seed")
	}

	orderedC := OrderPuzzlesForPrint(input, "different-seed")
	if reflect.DeepEqual(orderedA, orderedC) {
		t.Fatal("expected different ordering with a different seed")
	}

	if got, want := len(orderedA), len(input); got != want {
		t.Fatalf("length = %d, want %d", got, want)
	}
}
