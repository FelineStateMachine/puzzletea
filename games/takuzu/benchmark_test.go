package takuzu

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

type takuzuBenchMode struct {
	size      int
	prefilled float64
}

var takuzuBenchModes = []takuzuBenchMode{
	{size: 8, prefilled: 0.40},
	{size: 10, prefilled: 0.32},
	{size: 12, prefilled: 0.30},
	{size: 14, prefilled: 0.28},
}

func BenchmarkGeneratePuzzleSizes(b *testing.B) {
	skipBenchmarkInShortMode(b)

	for _, mode := range takuzuBenchModes {
		name := fmt.Sprintf("%dx%d", mode.size, mode.size)
		b.Run(name, func(b *testing.B) {
			rng := rand.New(rand.NewPCG(uint64(mode.size), uint64(mode.size*17+3)))
			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				complete := generateCompleteSeeded(mode.size, rng)
				puzzle, provided := generatePuzzleSeeded(complete, mode.size, mode.prefilled, rng)
				if len(puzzle) != mode.size || len(provided) != mode.size {
					b.Fatalf("unexpected puzzle dimensions for size %d", mode.size)
				}
			}
		})
	}
}

func BenchmarkCountSolutionsSizes(b *testing.B) {
	skipBenchmarkInShortMode(b)

	for _, mode := range takuzuBenchModes {
		name := fmt.Sprintf("%dx%d", mode.size, mode.size)
		b.Run(name, func(b *testing.B) {
			rng := rand.New(rand.NewPCG(uint64(mode.size*101), uint64(mode.size*37+11)))
			complete := generateCompleteSeeded(mode.size, rng)
			puzzle, _ := generatePuzzleSeeded(complete, mode.size, mode.prefilled, rng)

			b.ReportAllocs()
			b.ResetTimer()

			for range b.N {
				count := countSolutions(puzzle, mode.size, 2)
				if count != 1 {
					b.Fatalf("countSolutions returned %d, want 1", count)
				}
			}
		})
	}
}
