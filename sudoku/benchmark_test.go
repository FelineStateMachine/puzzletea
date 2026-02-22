package sudoku

import (
	"fmt"
	"math/rand/v2"
	"testing"
)

func skipBenchmarkInShortMode(b *testing.B) {
	b.Helper()
	if testing.Short() {
		b.Skip("skipping benchmark in short mode")
	}
}

func benchmarkSudokuModes() []SudokuMode {
	return []SudokuMode{
		NewMode("Beginner", "benchmark", 45),
		NewMode("Easy", "benchmark", 38),
		NewMode("Medium", "benchmark", 32),
		NewMode("Hard", "benchmark", 27),
		NewMode("Expert", "benchmark", 22),
		NewMode("Diabolical", "benchmark", 17),
	}
}

func BenchmarkGenerateProvidedCellsByMode(b *testing.B) {
	skipBenchmarkInShortMode(b)

	for modeIndex, mode := range benchmarkSudokuModes() {
		mode := mode
		modeIndex := modeIndex
		b.Run(mode.Title(), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				rng := rand.New(rand.NewPCG(uint64(i+1), uint64(modeIndex+10_000)))
				cells := GenerateProvidedCellsSeeded(mode, rng)
				if len(cells) < mode.ProvidedCount {
					b.Fatalf("provided count %d below target %d", len(cells), mode.ProvidedCount)
				}
			}
		})
	}
}

func BenchmarkCountSolutionsByDifficulty(b *testing.B) {
	skipBenchmarkInShortMode(b)

	for modeIndex, mode := range benchmarkSudokuModes() {
		mode := mode
		modeIndex := modeIndex
		b.Run(mode.Title(), func(b *testing.B) {
			rng := rand.New(rand.NewPCG(uint64(modeIndex+1), uint64(modeIndex+2)))
			cells := GenerateProvidedCellsSeeded(mode, rng)
			source := newGrid(cells)
			if got := countSolutions(&source, 2); got != 1 {
				b.Fatalf("expected unique seed puzzle for %s, got %d", mode.Title(), got)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				g := source
				if got := countSolutions(&g, 2); got != 1 {
					b.Fatalf("countSolutions(%s) = %d, want 1", mode.Title(), got)
				}
			}
		})
	}
}

func BenchmarkGenerateProvidedCellsSeededStable(b *testing.B) {
	skipBenchmarkInShortMode(b)

	mode := NewMode("Hard", "benchmark", 27)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rng := rand.New(rand.NewPCG(uint64(i+1), uint64(i+100)))
		cells := GenerateProvidedCellsSeeded(mode, rng)
		if len(cells) < mode.ProvidedCount {
			b.Fatalf("generated %d clues, want >= %d", len(cells), mode.ProvidedCount)
		}
	}
}

func BenchmarkCountSolutionsFixedPuzzle(b *testing.B) {
	skipBenchmarkInShortMode(b)

	mode := NewMode("Expert", "benchmark", 22)
	cells := GenerateProvidedCellsSeeded(mode, rand.New(rand.NewPCG(42, 84)))
	source := newGrid(cells)
	if got := countSolutions(&source, 2); got != 1 {
		b.Fatalf("seed puzzle not unique: %d", got)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g := source
		count := countSolutions(&g, 2)
		if count != 1 {
			b.Fatalf("iteration %d: countSolutions=%d, want 1", i, count)
		}
	}
}

func BenchmarkIsValidAllCells(b *testing.B) {
	skipBenchmarkInShortMode(b)

	mode := NewMode("Medium", "benchmark", 32)
	cells := GenerateProvidedCellsSeeded(mode, rand.New(rand.NewPCG(7, 11)))
	g := newGrid(cells)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for y := range gridSize {
			for x := range gridSize {
				val := g[y][x].v
				if val == 0 {
					val = ((x + y) % gridSize) + 1
				}
				if !isValid(&g, val, x, y) && g[y][x].v == val {
					b.Fatalf("unexpected invalid self-check at (%d,%d)", x, y)
				}
			}
		}
	}
}

func BenchmarkCountSolutionsModesSummary(b *testing.B) {
	skipBenchmarkInShortMode(b)

	modes := benchmarkSudokuModes()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mode := modes[i%len(modes)]
		rng := rand.New(rand.NewPCG(uint64(i+1), uint64(i+2)))
		cells := GenerateProvidedCellsSeeded(mode, rng)
		g := newGrid(cells)
		if got := countSolutions(&g, 2); got != 1 {
			b.Fatalf("mode %s uniqueness check failed: %d", mode.Title(), got)
		}
	}
}

func ExampleBenchmarkGenerateProvidedCellsByMode() {
	fmt.Println("run with: go test ./sudoku -run '^$' -bench BenchmarkGenerateProvidedCellsByMode -benchmem")
	// Output: run with: go test ./sudoku -run '^$' -bench BenchmarkGenerateProvidedCellsByMode -benchmem
}
