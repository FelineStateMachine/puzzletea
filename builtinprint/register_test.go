package builtinprint

import (
	"testing"

	"github.com/FelineStateMachine/puzzletea/pdfexport"
)

func TestRegisterBuiltins(t *testing.T) {
	Register()

	tests := []string{
		"Fillomino",
		"hashi",
		"Hitori",
		"Nonogram",
		"Nurikabe",
		"Ripple Effect",
		"Shikaku",
		"spellpuzzle",
		"Sudoku",
		"Sudoku RGB",
		"Takuzu",
		"binario+",
		"Word Search",
	}

	for _, name := range tests {
		if !pdfexport.HasPrintAdapter(name) {
			t.Fatalf("expected print adapter for %q", name)
		}
	}
}
