package sudoku

import (
	"math/rand/v2"
	"testing"
)

func TestGenerateProvidedCellsUniquenessAcrossModes(t *testing.T) {
	for modeIndex, mode := range benchmarkSudokuModes() {
		mode := mode
		modeIndex := modeIndex

		t.Run(mode.Title(), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(uint64(modeIndex+101), uint64(modeIndex+1001)))
			cells := GenerateProvidedCellsSeeded(mode, rng)
			if len(cells) < mode.ProvidedCount {
				t.Fatalf("generated %d clues, want >= %d", len(cells), mode.ProvidedCount)
			}

			g := newGrid(cells)
			if got := countSolutions(&g, 2); got != 1 {
				t.Fatalf("countSolutions(limit=2) = %d, want 1", got)
			}
		})
	}
}
